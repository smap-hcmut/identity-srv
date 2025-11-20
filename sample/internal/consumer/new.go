package consumer

import (
	"database/sql"

	"smap-api/config"
	smtpConsumer "smap-api/internal/smtp/rabbitmq/consumer"
	smtpUseCase "smap-api/internal/smtp/usecase"
	pkgLog "smap-api/pkg/log"
	pkgRabbitMQ "smap-api/pkg/rabbitmq"
)

// Config holds all dependencies needed for consumer service
type Config struct {
	Logger     pkgLog.Logger
	PostgresDB *sql.DB
	AMQPConn   *pkgRabbitMQ.Connection
	SMTPConfig config.SMTPConfig
}

// New creates a new Consumer service with all necessary dependencies
func New(cfg Config) (*Consumer, error) {
	// Validate required dependencies
	if cfg.Logger == nil {
		return nil, ErrLoggerRequired
	}
	if cfg.PostgresDB == nil {
		return nil, ErrPostgresDBRequired
	}
	if cfg.AMQPConn == nil {
		return nil, ErrAMQPConnRequired
	}
	if cfg.SMTPConfig.Host == "" {
		return nil, ErrSMTPConfigRequired
	}

	// Initialize SMTP use case
	smtpUC := smtpUseCase.New(cfg.Logger, cfg.SMTPConfig)

	// Initialize SMTP consumer
	smtpCons := smtpConsumer.NewConsumer(cfg.Logger, cfg.AMQPConn, smtpUC)

	return &Consumer{
		l:            cfg.Logger,
		smtpConsumer: smtpCons,
	}, nil
}
