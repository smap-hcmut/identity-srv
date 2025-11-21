package consumer

import "errors"

var (
	ErrLoggerRequired     = errors.New("logger is required")
	ErrPostgresDBRequired = errors.New("postgres database is required")
	ErrAMQPConnRequired   = errors.New("amqp connection is required")
	ErrSMTPConfigRequired = errors.New("smtp config is required")
)
