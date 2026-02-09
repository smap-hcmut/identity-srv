package postgre

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smap-api/internal/audit"
	"smap-api/internal/audit/repository"
	"strings"
	"time"
)

type auditRepository struct {
	db *sql.DB
}

// New creates a new audit repository
func New(db *sql.DB) repository.Repository {
	return &auditRepository{
		db: db,
	}
}

// BatchInsert inserts multiple audit events into the database
func (r *auditRepository) BatchInsert(ctx context.Context, events []audit.AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Build batch insert query
	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*8)

	for i, event := range events {
		// Marshal metadata to JSON
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}

		// Calculate expires_at (90 days from created_at)
		expiresAt := event.Timestamp.Add(90 * 24 * time.Hour)

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8,
		))

		valueArgs = append(valueArgs,
			event.UserID,
			string(event.Action),
			event.ResourceType,
			event.ResourceID,
			string(metadataJSON),
			event.IPAddress,
			event.UserAgent,
			event.Timestamp,
		)

		// Note: expires_at will be calculated in the query
		_ = expiresAt // Will use in query
	}

	query := fmt.Sprintf(`
		INSERT INTO audit_logs (
			user_id, action, resource_type, resource_id, 
			metadata, ip_address, user_agent, created_at, expires_at
		) VALUES %s
	`, strings.Join(valueStrings, ","))

	// Update query to include expires_at calculation
	valueStrings = make([]string, 0, len(events))
	valueArgs = make([]interface{}, 0, len(events)*8)

	for i, event := range events {
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}

		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d::timestamp + interval '90 days')",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8, i*8+8,
		))

		valueArgs = append(valueArgs,
			event.UserID,
			string(event.Action),
			event.ResourceType,
			event.ResourceID,
			string(metadataJSON),
			event.IPAddress,
			event.UserAgent,
			event.Timestamp,
		)
	}

	query = fmt.Sprintf(`
		INSERT INTO audit_logs (
			user_id, action, resource_type, resource_id, 
			metadata, ip_address, user_agent, created_at, expires_at
		) VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert audit logs: %w", err)
	}

	return nil
}

// DeleteExpired deletes audit logs older than the retention period
func (r *auditRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM audit_logs WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
