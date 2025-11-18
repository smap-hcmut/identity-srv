package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	smtpConsumer "smap-api/internal/smtp/rabbitmq/consumer"
	pkgLog "smap-api/pkg/log"
)

type Consumer struct {
	l            pkgLog.Logger
	smtpConsumer smtpConsumer.Consumer
}

// Run starts all consumers and waits for termination signal
func (c *Consumer) Run() error {
	ctx := context.Background()
	c.l.Info(ctx, "Starting SMAP Consumer Service...")

	// Start SMTP email consumer
	c.l.Info(ctx, "Starting SMTP Email Consumer...")
	c.smtpConsumer.Consume()

	c.l.Info(ctx, "All consumers started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	c.l.Infof(ctx, "Received signal: %v. Shutting down gracefully...", sig)

	return nil
}

// Health returns the health status of consumers
func (c *Consumer) Health() error {
	// TODO: Implement health checks for RabbitMQ connections
	return nil
}

// String returns consumer service information
func (c *Consumer) String() string {
	return fmt.Sprintf("SMAP Consumer Service [SMTP Email Consumer]")
}
