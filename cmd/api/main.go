package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"smap-api/config"
	"smap-api/config/postgre"
	"smap-api/internal/httpserver"
	"smap-api/pkg/discord"
	"smap-api/pkg/encrypter"
	"smap-api/pkg/log"
	"syscall"
)

// @Name Smap API
// @description This is the API documentation for Smap.
// @version 1
// @host localhost:8080
// @schemes http
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return
	}

	// Initialize logger
	l := log.Init(log.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	// Register graceful shutdown
	registerGracefulShutdown(l)

	// Initialize encrypter
	enc := encrypter.New(cfg.Encrypter.Key)

	// Initialize PostgreSQL
	postgresDB, err := postgre.Connect(context.Background(), cfg.Postgres)
	if err != nil {
		l.Error(context.Background(), "Failed to connect to PostgreSQL: ", err)
		return
	}
	defer postgre.Disconnect(context.Background(), postgresDB)

	// Initialize Discord
	discordClient, err := discord.New(l, &discord.DiscordWebhook{
		ID:    cfg.Discord.WebhookID,
		Token: cfg.Discord.WebhookToken,
	})
	if err != nil {
		l.Error(context.Background(), "Failed to initialize Discord: ", err)
		return
	}

	// Initialize HTTP server
	srv, err := httpserver.New(l, httpserver.Config{
		// Server Configuration
		Logger: l,
		Host:   cfg.HTTPServer.Host,
		Port:   cfg.HTTPServer.Port,
		Mode:   cfg.HTTPServer.Mode,

		// Database Configuration
		PostgresDB: postgresDB,

		// Authentication & Security Configuration
		JwtSecretKey: cfg.JWT.SecretKey,
		Encrypter:    enc,
		InternalKey:  cfg.InternalConfig.InternalKey,

		// Monitoring & Notification Configuration
		Discord: discordClient,
	})
	if err != nil {
		l.Error(context.Background(), "Failed to initialize HTTP server: ", err)
		return
	}

	if err := srv.Run(); err != nil {
		l.Error(context.Background(), "Failed to run server: ", err)
		return
	}
}
func registerGracefulShutdown(l log.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		l.Info(context.Background(), "Shutting down gracefully...")

		l.Info(context.Background(), "Cleanup completed")
		os.Exit(0)
	}()
}
