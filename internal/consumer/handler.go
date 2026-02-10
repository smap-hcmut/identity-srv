package consumer

import (
	"context"
	"time"

	auditKafka "smap-api/internal/audit/delivery/kafka"
	auditPostgre "smap-api/internal/audit/repository/postgre"
)

// registerAuditConsumer registers the audit consumer
// Similar to how httpserver/handler.go initializes and maps audit routes
func (srv *Consumer) registerAuditConsumer(ctx context.Context) error {
	// Initialize audit repository
	auditRepo := auditPostgre.New(srv.postgresDB)

	// Initialize audit consumer
	auditConsumer := auditKafka.New(
		auditRepo,
		auditKafka.Config{
			BatchSize:    100,
			BatchTimeout: 5 * time.Second,
		},
		srv.logger,
	)

	// Register the consumer
	if err := srv.register(auditConsumer); err != nil {
		return err
	}

	srv.logger.Info(ctx, "Audit consumer registered")

	return nil
}

// TODO: Add more consumer registration methods here
// Example:
// func (srv *Consumer) registerNotificationConsumer(ctx context.Context) error {
//     notificationRepo := notificationPostgre.New(srv.postgresDB)
//     notificationConsumer := notificationKafka.New(notificationRepo, srv.logger)
//     return srv.register(notificationConsumer)
// }
