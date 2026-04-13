package main

import (
	"context"
	"fmt"
	"identity-srv/config"
	configPostgre "identity-srv/config/postgre"
	_ "identity-srv/docs" // Import swagger docs
	authUsecase "identity-srv/internal/authentication/usecase"
	"identity-srv/internal/consumer"
	"identity-srv/internal/httpserver"
	"os"
	"os/signal"
	"syscall"

	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/discord"
	"github.com/smap-hcmut/shared-libs/go/encrypter"
	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/redis"
	_ "github.com/smap-hcmut/shared-libs/go/response" // For swagger type definitions
)

// @title       SMAP Identity Service API
// @description SMAP Identity Service API documentation.
// @version     1
// @schemes     https http
// @BasePath    /identity/api/v1
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
	// 1. Load configuration
	// Reads config from YAML file and environment variables
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return
	}

	// 2. Initialize logger
	logger := log.NewZapLogger(log.ZapConfig{
		Level:        cfg.Logger.Level,
		Mode:         cfg.Logger.Mode,
		Encoding:     cfg.Logger.Encoding,
		ColorEnabled: cfg.Logger.ColorEnabled,
	})

	// 3. Context with signal-based cancellation (replaces registerGracefulShutdown)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info(ctx, "Starting Identity Service...")

	// 4. Initialize encrypter
	encrypterInstance := encrypter.New(cfg.Encrypter.Key)

	postgresDB, err := configPostgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Error(ctx, "Failed to connect to PostgreSQL: ", err)
		return
	}
	defer configPostgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// 6. Initialize Discord (optional)
	var webhookURL string
	if cfg.Discord.WebhookID != "" && cfg.Discord.WebhookToken != "" {
		webhookURL = fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", cfg.Discord.WebhookID, cfg.Discord.WebhookToken)
	}
	discordClient, err := discord.New(logger, webhookURL)
	if err != nil {
		logger.Warnf(ctx, "Discord webhook not configured (optional): %v", err)
		discordClient = nil // Continue without Discord
	} else {
		logger.Infof(ctx, "Discord webhook initialized successfully")
	}

	// 7. Initialize Redis
	redisClient, err := redis.New(redis.RedisConfig{
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

	// 9. Initialize JWT Manager
	jwtManager := auth.NewManager(cfg.JWT.SecretKey)
	logger.Infof(ctx, "JWT Manager initialized")

	// 10. Initialize Redirect Validator
	// Validates OAuth redirect URLs against whitelist to prevent open redirect attacks
	redirectValidator := authUsecase.NewRedirectValidator(cfg.AccessControl.AllowedRedirectURLs)
	logger.Infof(ctx, "Redirect validator initialized with %d allowed URLs", len(cfg.AccessControl.AllowedRedirectURLs))

	// ── Consumer (Kafka) ────────────────────────────────────────────────────
	consumerService, err := consumer.New(logger, consumer.Config{
		PostgresDB:   postgresDB,
		KafkaBrokers: cfg.Kafka.Brokers,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to initialize consumer service: %v", err)
		return
	}
	defer consumerService.Close()

	go func() {
		if err := consumerService.Start(ctx); err != nil {
			logger.Errorf(ctx, "Consumer service error: %v", err)
		}
	}()
	logger.Info(ctx, "Consumer service started in background")

	// ── HTTP Server ─────────────────────────────────────────────────────────
	// Initialize HTTP server
	// Main application server that handles all HTTP requests and routes
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
		Config:            cfg,
		JWTManager:        jwtManager,
		RedisClient:       redisClient,
		RedirectValidator: redirectValidator,
		CookieConfig:      cfg.Cookie,
		Encrypter:         encrypterInstance,

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
	logger.Info(ctx, "Identity service stopped gracefully")
}
