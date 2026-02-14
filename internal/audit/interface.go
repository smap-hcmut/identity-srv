package audit

import "context"

// Publisher publishes audit events to Kafka
type Publisher interface {
	Publish(ctx context.Context, event AuditEvent) error
	Close() error
}

// Logger interface for audit module
type Logger interface {
	Infof(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
}
