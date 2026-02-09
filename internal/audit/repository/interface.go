package repository

import (
	"context"
	"smap-api/internal/audit"
)

// Repository defines the interface for audit log storage
type Repository interface {
	// BatchInsert inserts multiple audit events into the database
	BatchInsert(ctx context.Context, events []audit.AuditEvent) error

	// DeleteExpired deletes audit logs older than the retention period
	DeleteExpired(ctx context.Context) (int64, error)
}
