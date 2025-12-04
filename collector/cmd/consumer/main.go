package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"smap-collector/config"
	"smap-collector/internal/consumer"
	"smap-collector/pkg/discord"
	pkgLog "smap-collector/pkg/log"
	"smap-collector/pkg/rabbitmq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	l := pkgLog.Init(pkgLog.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	// Initialize Discord webhook for error reporting
	discordWebhook, err := discord.NewDiscordWebhook(cfg.Discord.ReportBugID, cfg.Discord.ReportBugToken)
	if err != nil {
		l.Warnf(ctx, "failed to initialize Discord webhook: %v", err)
	}

	conn, err := rabbitmq.Dial(cfg.RabbitMQConfig.URL, true)
	if err != nil {
		l.Fatalf(ctx, "failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	srv, err := consumer.New(consumer.Config{
		Logger:        l,
		AMQPConn:      conn,
		Discord:       discordWebhook,
		ProjectConfig: cfg.Project,
	})
	if err != nil {
		l.Fatalf(ctx, "failed to init consumer: %v", err)
	}
	defer srv.Close()

	if err := srv.Run(ctx); err != nil {
		l.Fatalf(ctx, "consumer stopped with error: %v", err)
	}
}
