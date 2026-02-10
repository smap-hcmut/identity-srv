package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"smap-api/config"
	configPostgre "smap-api/config/postgre"
	"smap-api/internal/audit"
	authrepo "smap-api/internal/authentication/repository"
	"smap-api/internal/scheduler"
	"smap-api/pkg/jwt/rotation"
	pkgLog "smap-api/pkg/log"

	_ "github.com/lib/pq"
)

// @Name SMAP Scheduler Service
// @description Scheduler service for running periodic jobs (Audit Cleanup, etc.)
// @version 1.0
func main() {
	// 1. Load configuration
	// Reads config from YAML file and environment variables
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize logger
	// Sets up structured logging with configurable level and format
	logger := pkgLog.Init(pkgLog.ZapConfig{
		Level:        cfg.Logger.Level,
		Mode:         cfg.Logger.Mode,
		Encoding:     cfg.Logger.Encoding,
		ColorEnabled: cfg.Logger.ColorEnabled,
	})

	// 3. Initialize PostgreSQL
	// Main database connection for persistent data storage
	ctx := context.Background()
	postgresDB, err := configPostgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Errorf(ctx, "Failed to connect to PostgreSQL: %v", err)
		os.Exit(1)
	}
	defer configPostgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// 4. Initialize JWT keys repository
	jwtKeysRepo := authrepo.NewJWTKeysRepository(postgresDB)

	// 5. Initialize key rotation manager
	rotationManager := rotation.NewManager(
		jwtKeysRepo,
		cfg.KeyRotation.RotationDays,
		cfg.KeyRotation.GraceDays,
	)

	// 6. Initialize audit publisher (optional - can be nil)
	var auditPublisher audit.Publisher
	// TODO: Initialize Kafka audit publisher if needed
	// auditPublisher = kafka.NewPublisher(...)

	// 7. Initialize scheduler service
	// Manages all periodic background jobs (similar to httpserver)
	schedulerService, err := scheduler.New(logger, scheduler.Config{
		Logger:          logger,
		PostgresDB:      postgresDB,
		Config:          cfg,
		JWTKeysRepo:     jwtKeysRepo,
		RotationManager: rotationManager,
		AuditPublisher:  auditPublisher,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize scheduler service: %v", err)
		os.Exit(1)
	}

	// 5. Start scheduler
	// Jobs are registered internally in scheduler/handler.go
	if err := schedulerService.Start(); err != nil {
		logger.Errorf(ctx, "Failed to start scheduler service: %v", err)
		os.Exit(1)
	}

	logger.Info(ctx, "Scheduler service ready - running periodic jobs")

	// 8. Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info(ctx, "Shutting down scheduler service gracefully...")
	schedulerService.Stop()
	logger.Info(ctx, "Scheduler service stopped gracefully")
}
