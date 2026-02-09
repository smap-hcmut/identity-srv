package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"smap-api/config"
	configPostgre "smap-api/config/postgre"
	_ "smap-api/docs" // Import swagger docs
	auditPostgre "smap-api/internal/audit/repository/postgre"
	"smap-api/internal/audit/usecase"
	"smap-api/internal/httpserver"
	"smap-api/pkg/discord"
	"smap-api/pkg/encrypter"
	pkgGoogle "smap-api/pkg/google"
	pkgJWT "smap-api/pkg/jwt"
	pkgKafka "smap-api/pkg/kafka"
	"smap-api/pkg/log"
	pkgRedis "smap-api/pkg/redis"
	"syscall"
	"time"
)

// @title       SMAP Identity Service API
// @description SMAP Identity Service API documentation.
// @version     1
// @host        smap-api.tantai.dev
// @schemes     https
// @BasePath    /identity
//
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name smap_auth_token
// @description Authentication token stored in HttpOnly cookie. Set automatically by /login endpoint.
//
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Legacy Bearer token authentication (deprecated - use cookie authentication instead). Format: "Bearer {token}"
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return
	}

	// Initialize logger
	logger := log.Init(log.ZapConfig{
		Level:        cfg.Logger.Level,
		Mode:         cfg.Logger.Mode,
		Encoding:     cfg.Logger.Encoding,
		ColorEnabled: cfg.Logger.ColorEnabled,
	})

	// Register graceful shutdown
	registerGracefulShutdown(logger)

	// Initialize encrypter
	encrypterInstance := encrypter.New(cfg.Encrypter.Key)

	// Initialize PostgreSQL
	ctx := context.Background()
	postgresDB, err := configPostgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Error(ctx, "Failed to connect to PostgreSQL: ", err)
		return
	}
	defer configPostgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// Initialize Discord (optional)
	discordClient, err := discord.New(logger, &discord.DiscordWebhook{
		ID:    cfg.Discord.WebhookID,
		Token: cfg.Discord.WebhookToken,
	})
	if err != nil {
		logger.Warnf(ctx, "Discord webhook not configured (optional): %v", err)
		discordClient = nil // Continue without Discord
	} else {
		logger.Infof(ctx, "Discord webhook initialized successfully")
	}

	// Initialize Redis
	redisClient, err := pkgRedis.New(pkgRedis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Error(ctx, "Failed to connect to Redis: ", err)
		return
	}
	logger.Infof(ctx, "Redis connected successfully to %s:%d (DB %d)", cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)

	// Initialize Redis for token blacklist (Task 2.9)
	// Use separate DB (DB=1) for blacklist to avoid key conflicts
	blacklistRedis, err := pkgRedis.New(pkgRedis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       1, // Separate DB for blacklist
	})
	if err != nil {
		logger.Error(ctx, "Failed to connect to Redis for blacklist: ", err)
		return
	}
	logger.Infof(ctx, "Redis blacklist connected successfully to %s:%d (DB 1)", cfg.Redis.Host, cfg.Redis.Port)
	_ = blacklistRedis // Will be used in Task 3.5 for token blacklist manager

	// Initialize JWT Manager
	jwtManager, err := pkgJWT.New(pkgJWT.Config{
		PrivateKeyPath: cfg.JWT.PrivateKeyPath,
		PublicKeyPath:  cfg.JWT.PublicKeyPath,
		Issuer:         cfg.JWT.Issuer,
		Audience:       cfg.JWT.Audience,
		TTL:            time.Duration(cfg.JWT.TTL) * time.Second,
	})
	if err != nil {
		logger.Error(ctx, "Failed to initialize JWT manager: ", err)
		return
	}
	logger.Infof(ctx, "JWT Manager initialized with algorithm: %s", cfg.JWT.Algorithm)

	// Initialize Google Directory API client (optional - needed for Day 3 Groups RBAC)
	googleClient, err := pkgGoogle.New(ctx, pkgGoogle.Config{
		ServiceAccountKey: cfg.GoogleWorkspace.ServiceAccountKey,
		AdminEmail:        cfg.GoogleWorkspace.AdminEmail,
		Domain:            cfg.GoogleWorkspace.Domain,
	})
	if err != nil {
		logger.Warnf(ctx, "Google Directory API not configured (needed for Day 3 Groups RBAC): %v", err)
		googleClient = nil // Continue without Google Directory API
	} else {
		// Test connection
		if err := googleClient.HealthCheck(ctx); err != nil {
			logger.Warnf(ctx, "Google Directory API health check failed (will retry on demand): %v", err)
		} else {
			logger.Infof(ctx, "Google Directory API connected successfully for domain: %s", cfg.GoogleWorkspace.Domain)
		}
	}

	// Initialize Kafka producer
	kafkaProducer, err := pkgKafka.NewProducer(pkgKafka.Config{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	if err != nil {
		logger.Warnf(ctx, "Failed to initialize Kafka producer (audit logging will be buffered): %v", err)
		kafkaProducer = nil // Continue without Kafka - will use in-memory buffer
	} else {
		logger.Infof(ctx, "Kafka producer initialized successfully for topic: %s", cfg.Kafka.Topic)
	}
	if kafkaProducer != nil {
		defer kafkaProducer.Close()
	}

	// Initialize audit log cleanup job (Task 2.8)
	auditRepo := auditPostgre.New(postgresDB)
	cleanupJob := usecase.NewCleanupJob(auditRepo, logger)
	if err := cleanupJob.Start(); err != nil {
		logger.Warnf(ctx, "Failed to start audit cleanup job: %v", err)
	} else {
		logger.Infof(ctx, "Audit log cleanup job started (runs daily at 2 AM)")
		defer cleanupJob.Stop()
	}

	// Initialize HTTP server
	httpServer, err := httpserver.New(logger, httpserver.Config{
		// Server Configuration
		Logger:      logger,
		Host:        cfg.HTTPServer.Host,
		Port:        cfg.HTTPServer.Port,
		Mode:        cfg.HTTPServer.Mode,
		Environment: cfg.Environment.Name,

		// Database Configuration
		PostgresDB: postgresDB,

		// Authentication & Security Configuration
		Config:         cfg,
		JWTManager:     jwtManager,
		RedisClient:    redisClient,
		BlacklistRedis: blacklistRedis,
		CookieConfig:   cfg.Cookie,
		Encrypter:      encrypterInstance,

		// Google Workspace Integration
		GoogleClient: googleClient,

		// Kafka Integration
		KafkaProducer: kafkaProducer,

		// Monitoring & Notification Configuration
		Discord: discordClient,
	})
	if err != nil {
		logger.Error(ctx, "Failed to initialize HTTP server: ", err)
		return
	}

	if err := httpServer.Run(); err != nil {
		logger.Error(ctx, "Failed to run server: ", err)
		return
	}
}

// registerGracefulShutdown registers a signal handler for graceful shutdown.
func registerGracefulShutdown(logger log.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info(context.Background(), "Shutting down gracefully...")

		logger.Info(context.Background(), "Cleanup completed")
		os.Exit(0)
	}()
}
