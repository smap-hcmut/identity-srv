package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"smap-project/config"
	"smap-project/config/postgre"
	"smap-project/internal/httpserver"
	"smap-project/pkg/discord"
	"smap-project/pkg/encrypter"
	"smap-project/pkg/log"
	"syscall"
)

// @title       Smap API
// @description SMAP Project Service API documentation.
// @version     1
// @host        localhost:8080
// @schemes     http
// @BasePath    /project
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return
	}

	// Initialize logger
	logger := log.Init(log.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	// Register graceful shutdown
	registerGracefulShutdown(logger)

	// Initialize encrypter
	encrypterInstance := encrypter.New(cfg.Encrypter.Key)

	// Initialize PostgreSQL
	ctx := context.Background()
	postgresDB, err := postgre.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Error(ctx, "Failed to connect to PostgreSQL: ", err)
		return
	}
	defer postgre.Disconnect(ctx, postgresDB)
	logger.Infof(ctx, "PostgreSQL connected successfully to %s:%d/%s", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// // Initialize MinIO
	// minioClient, err := minio.Connect(ctx, cfg.MinIO)
	// if err != nil {
	// 	logger.Error(ctx, "Failed to connect to MinIO: ", err)
	// 	return
	// }
	// defer minio.Disconnect(ctx)
	// logger.Infof(ctx, "MinIO connected successfully to %s", cfg.MinIO.Endpoint)

	// // Initialize RabbitMQ
	// amqpConn, err := rabbitmq.Dial(cfg.RabbitMQ.URL, true)
	// if err != nil {
	// 	logger.Error(ctx, "Failed to connect to RabbitMQ: ", err)
	// 	return
	// }
	// defer amqpConn.Close()

	// Initialize Discord
	discordClient, err := discord.New(logger, &discord.DiscordWebhook{
		ID:    cfg.Discord.WebhookID,
		Token: cfg.Discord.WebhookToken,
	})
	if err != nil {
		logger.Error(ctx, "Failed to initialize Discord: ", err)
		return
	}

	// Initialize HTTP server
	httpServer, err := httpserver.New(logger, httpserver.Config{
		// Server Configuration
		Logger: logger,
		Host:   cfg.HTTPServer.Host,
		Port:   cfg.HTTPServer.Port,
		Mode:   cfg.HTTPServer.Mode,

		// Database Configuration
		PostgresDB: postgresDB,

		// // Storage Configuration
		// MinIO: minioClient,

		// // Message Queue Configuration
		// AmqpConn: amqpConn,

		// Authentication & Security Configuration
		JwtSecretKey: cfg.JWT.SecretKey,
		Encrypter:    encrypterInstance,
		InternalKey:  cfg.InternalConfig.InternalKey,

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
