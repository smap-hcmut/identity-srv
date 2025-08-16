package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/internal/appconfig/minio"
	"github.com/nguyentantai21042004/smap-api/internal/appconfig/mongo"
	"github.com/nguyentantai21042004/smap-api/internal/httpserver"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	pkgCrt "github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
)

// @Name Smap API
// @description This is the API documentation for Smap.
// @version 1
// @host smap-api.ngtantai.pro
// @schemes https
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown()

	// Initialize logger first (needed for all other services)
	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	// Initialize Encrypter
	encrypter := pkgCrt.NewEncrypter(cfg.Encrypter.Key)

	// =============================================================================
	// DATABASE CONFIGURATION
	// =============================================================================

	// Initialize MongoDB
	client, err := mongo.Connect(cfg.Mongo, encrypter)
	if err != nil {
		panic(err)
	}
	db := client.Database(cfg.Mongo.Database)
	defer mongo.Disconnect(client)

	// =============================================================================
	// MESSAGE QUEUE CONFIGURATION
	// =============================================================================

	// =============================================================================
	// STORAGE CONFIGURATION
	// =============================================================================

	// Initialize MinIO
	// log.Println("Connecting to MinIO...")
	minioClient, err := minio.Connect(context.Background(), cfg.MinIO)
	if err != nil {
		log.Fatal("Failed to connect to MinIO: ", err)
	}
	defer minio.Close()

	// =============================================================================
	// AUTHENTICATION & SECURITY CONFIGURATION
	// =============================================================================

	// =============================================================================
	// EXTERNAL SERVICES CONFIGURATION
	// =============================================================================

	// SMTP is configured via config, no connection needed

	// =============================================================================
	// MONITORING & NOTIFICATION CONFIGURATION
	// =============================================================================

	// Initialize Discord Webhook
	discordWebhook, err := discord.NewDiscordWebhook(cfg.Discord.ReportBugID, cfg.Discord.ReportBugToken)
	if err != nil {
		log.Fatal("Failed to initialize Discord webhook: ", err)
	}

	// =============================================================================
	// HTTP SERVER CONFIGURATION
	// =============================================================================

	srv, err := httpserver.New(l, httpserver.Config{
		// Server Configuration
		Logger: l,
		Host:   cfg.HTTPServer.Host,
		Port:   cfg.HTTPServer.Port,
		Mode:   cfg.HTTPServer.Mode,

		// Database Configuration
		MongoDB: db,

		// Storage Configuration
		MinIOClient: minioClient,

		// Authentication & Security Configuration
		JwtSecretKey: cfg.JWT.SecretKey,
		Encrypter:    encrypter,
		InternalKey:  cfg.InternalConfig.InternalKey,

		// WebSocket Configuration
		WebSocketConfig: cfg.WebSocket,

		// Monitoring & Notification Configuration
		DiscordConfig: discordWebhook,
	})
	if err != nil {
		log.Fatal("Failed to initialize HTTP server: ", err)
	}

	// =============================================================================
	// START SERVER
	// =============================================================================

	if err := srv.Run(); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")

		log.Println("Cleanup completed")
		os.Exit(0)
	}()
}
