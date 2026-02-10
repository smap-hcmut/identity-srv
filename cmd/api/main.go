package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"smap-api/config"
	configPostgre "smap-api/config/postgre"
	_ "smap-api/docs" // Import swagger docs
	authrepo "smap-api/internal/authentication/repository"
	authUsecase "smap-api/internal/authentication/usecase"
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
	// 1. Load configuration
	// Reads config from YAML file and environment variables
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return
	}

	// 2. Initialize logger
	logger := log.Init(log.ZapConfig{
		Level:        cfg.Logger.Level,
		Mode:         cfg.Logger.Mode,
		Encoding:     cfg.Logger.Encoding,
		ColorEnabled: cfg.Logger.ColorEnabled,
	})

	// 3. Register graceful shutdown
	registerGracefulShutdown(logger)

	// 4. Initialize encrypter
	encrypterInstance := encrypter.New(cfg.Encrypter.Key)

	// 5. Initialize PostgreSQL
	ctx := context.Background()
	postgresDB, err := configPostgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Error(ctx, "Failed to connect to PostgreSQL: ", err)
		return
	}
	defer configPostgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// 6. Initialize Discord (optional)
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

	// 7. Initialize Redis
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

	// 8. Initialize Redis for token blacklist
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

	// 9. Initialize JWT Manager
	// Load keys from database for rotation support, fallback to file-based keys
	jwtManager, err := initializeJWTManager(ctx, logger, cfg, postgresDB)
	if err != nil {
		logger.Error(ctx, "Failed to initialize JWT manager: ", err)
		return
	}
	logger.Infof(ctx, "JWT Manager initialized with algorithm: %s", cfg.JWT.Algorithm)

	// 10. Initialize Google Directory API client
	googleClient, err := pkgGoogle.New(ctx, pkgGoogle.Config{
		ServiceAccountKey: cfg.GoogleWorkspace.ServiceAccountKey,
		AdminEmail:        cfg.GoogleWorkspace.AdminEmail,
		Domain:            cfg.GoogleWorkspace.Domain,
	})
	if err != nil {
		logger.Warnf(ctx, "Google Directory API not configured: %v", err)
		googleClient = nil
	} else {
		// Test connection
		if err := googleClient.HealthCheck(ctx); err != nil {
			logger.Warnf(ctx, "Google Directory API health check failed: %v", err)
		} else {
			logger.Infof(ctx, "Google Directory API connected successfully for domain: %s", cfg.GoogleWorkspace.Domain)
		}
	}

	// 11. Initialize Redirect Validator
	// Validates OAuth redirect URLs against whitelist to prevent open redirect attacks
	redirectValidator := authUsecase.NewRedirectValidator(cfg.AccessControl.AllowedRedirectURLs)
	logger.Infof(ctx, "Redirect validator initialized with %d allowed URLs", len(cfg.AccessControl.AllowedRedirectURLs))

	// 12. Initialize Rate Limiter
	// Prevents brute force attacks by limiting failed login attempts per IP
	rateLimiter := authUsecase.NewRateLimiter(
		redisClient.GetClient(),
		cfg.RateLimit.MaxAttempts,
		time.Duration(cfg.RateLimit.WindowMinutes)*time.Minute,
		time.Duration(cfg.RateLimit.BlockMinutes)*time.Minute,
	)
	logger.Infof(ctx, "Rate limiter initialized (max %d attempts per %d minutes, block for %d minutes)",
		cfg.RateLimit.MaxAttempts, cfg.RateLimit.WindowMinutes, cfg.RateLimit.BlockMinutes)

	// 13. Initialize Kafka producer
	// Publishes audit events to Kafka for async processing and long-term storage
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

	// 14. Initialize HTTP server
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
		BlacklistRedis:    blacklistRedis,
		RedirectValidator: redirectValidator,
		RateLimiter:       rateLimiter,
		CookieConfig:      cfg.Cookie,
		Encrypter:         encrypterInstance,

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

// initializeJWTManager initializes JWT manager with database keys or file-based keys
func initializeJWTManager(ctx context.Context, logger log.Logger, cfg *config.Config, db *sql.DB) (*pkgJWT.Manager, error) {
	// Try to load keys from database first (for rotation support)
	jwtKeysRepo := authrepo.NewJWTKeysRepository(db)
	keys, err := jwtKeysRepo.GetActiveAndRotatingKeys(ctx)

	if err == nil && len(keys) > 0 {
		// Database keys found - use rotation-enabled mode
		logger.Infof(ctx, "Loading %d JWT keys from database (rotation enabled)", len(keys))

		// Convert to JWTKeyData format
		keyData := make([]*pkgJWT.JWTKeyData, len(keys))
		for i, key := range keys {
			keyData[i] = &pkgJWT.JWTKeyData{
				KID:        key.KID,
				PrivateKey: key.PrivateKey,
				PublicKey:  key.PublicKey,
				IsActive:   key.IsActive(),
			}
		}

		// Create manager with database keys
		manager := &pkgJWT.Manager{}
		if err := manager.LoadKeys(keyData); err != nil {
			return nil, fmt.Errorf("failed to load keys from database: %w", err)
		}

		// Set issuer, audience, TTL
		manager.SetConfig(cfg.JWT.Issuer, cfg.JWT.Audience, time.Duration(cfg.JWT.TTL)*time.Second)

		return manager, nil
	}

	// No database keys - fallback to file-based keys (legacy mode)
	logger.Warn(ctx, "No keys found in database, using file-based keys (rotation disabled)")
	return pkgJWT.New(pkgJWT.Config{
		PrivateKeyPath: cfg.JWT.PrivateKeyPath,
		PublicKeyPath:  cfg.JWT.PublicKeyPath,
		Issuer:         cfg.JWT.Issuer,
		Audience:       cfg.JWT.Audience,
		TTL:            time.Duration(cfg.JWT.TTL) * time.Second,
	})
}
