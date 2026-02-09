package repository

import (
	"context"
	"smap-api/internal/audit"
	"smap-api/internal/model"
	"time"
)

// QueryOptions contains options for querying audit logs
type QueryOptions struct {
	UserID string
	Action string
	From   *time.Time
	To     *time.Time
	Page   int
	Limit  int
}

// Repository defines the interface for audit log storage
type Repository interface {
	// BatchInsert inserts multiple audit events into the database
	BatchInsert(ctx context.Context, events []audit.AuditEvent) error

	// DeleteExpired deletes audit logs older than the retention period
	DeleteExpired(ctx context.Context) (int64, error)

	// Query retrieves audit logs with pagination and filters
	Query(ctx context.Context, opts QueryOptions) ([]model.AuditLog, int, error)
}
