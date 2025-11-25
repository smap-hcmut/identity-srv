package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/internal/consumer"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	conn, err := rabbitmq.Dial(cfg.RabbitMQConfig.URL, true)
	if err != nil {
		l.Fatalf(ctx, "failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	srv, err := consumer.New(consumer.Config{
		Logger:   l,
		AMQPConn: conn,
	})
	if err != nil {
		l.Fatalf(ctx, "failed to init consumer: %v", err)
	}
	defer srv.Close()

	if err := srv.Run(ctx); err != nil {
		l.Fatalf(ctx, "consumer stopped with error: %v", err)
	}
}
