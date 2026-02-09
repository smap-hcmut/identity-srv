package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smap-api/config"
	configPostgre "smap-api/config/postgre"
	"smap-api/internal/audit/consumer"
	auditPostgre "smap-api/internal/audit/repository/postgre"
	pkgKafka "smap-api/pkg/kafka"
	pkgLog "smap-api/pkg/log"

	_ "github.com/lib/pq"
)

// @Name SMAP Consumer Service
// @description Consumer service for processing async tasks (Audit Logging, etc.)
// @version 1.0
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := pkgLog.Init(pkgLog.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	// Register graceful shutdown
	ctx := context.Background()
	registerGracefulShutdown(logger)

	// Initialize PostgreSQL
	postgresDB, err := configPostgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Errorf(ctx, "Failed to connect to PostgreSQL: %v", err)
		os.Exit(1)
	}
	defer configPostgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// Initialize Kafka consumer for audit logging (Task 2.7)
	kafkaConsumer, err := pkgKafka.NewConsumerGroup(pkgKafka.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers,
		GroupID: "audit-consumer-group",
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to create Kafka consumer: %v", err)
		os.Exit(1)
	}
	defer kafkaConsumer.Close()
	logger.Infof(ctx, "Kafka consumer connected successfully to %v", cfg.Kafka.Brokers)

	// Initialize audit repository
	auditRepo := auditPostgre.New(postgresDB)

	// Initialize audit consumer
	auditConsumer := consumer.New(
		kafkaConsumer,
		auditRepo,
		consumer.Config{
			Topic:        "audit.events",
			GroupID:      "audit-consumer-group",
			BatchSize:    100,
			BatchTimeout: 5 * time.Second,
		},
		logger,
	)

	// Start consuming in a goroutine
	go func() {
		logger.Info(ctx, "Starting audit consumer...")
		if err := auditConsumer.Start(ctx); err != nil {
			logger.Errorf(ctx, "Audit consumer error: %v", err)
		}
	}()

	logger.Info(ctx, "Consumer service ready - processing audit events")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info(ctx, "Shutting down consumer service...")
	auditConsumer.Close()
	logger.Info(ctx, "Consumer service stopped gracefully")
}

// registerGracefulShutdown registers a signal handler for graceful shutdown.
func registerGracefulShutdown(logger pkgLog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info(context.Background(), "Shutting down consumer service gracefully...")

		// Add cleanup logic here if needed

		logger.Info(context.Background(), "Cleanup completed")
		os.Exit(0)
	}()
}
