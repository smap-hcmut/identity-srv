package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"smap-api/config"
	"smap-api/config/postgre"
	"smap-api/internal/consumer"
	pkgLog "smap-api/pkg/log"
	pkgRabbitMQ "smap-api/pkg/rabbitmq"

	_ "github.com/lib/pq"
)

// @Name SMAP Consumer Service
// @description Consumer service for processing async tasks (Email, Notifications, etc.)
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
	postgresDB, err := postgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Errorf(ctx, "Failed to connect to PostgreSQL: %v", err)
		os.Exit(1)
	}
	defer postgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// Initialize RabbitMQ
	amqpConn, err := pkgRabbitMQ.Dial(cfg.RabbitMQ.URL, true)
	if err != nil {
		logger.Errorf(ctx, "Failed to connect to RabbitMQ: %v", err)
		os.Exit(1)
	}
	defer amqpConn.Close()
	logger.Infof(ctx, "RabbitMQ connected successfully")

	// Initialize Consumer Service
	srv, err := consumer.New(consumer.Config{
		Logger:     logger,
		PostgresDB: postgresDB,
		AMQPConn:   amqpConn,
		SMTPConfig: cfg.SMTP,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize consumer service: %v", err)
		os.Exit(1)
	}

	logger.Info(ctx, srv.String())

	// Run consumer service (blocks until shutdown)
	if err := srv.Run(); err != nil {
		logger.Errorf(ctx, "Consumer service error: %v", err)
		os.Exit(1)
	}

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
