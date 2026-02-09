package postgre

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smap-api/internal/audit"
	"smap-api/internal/audit/repository"
	"smap-api/internal/model"
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

// Query retrieves audit logs with pagination and filters
func (r *auditRepository) Query(ctx context.Context, opts repository.QueryOptions) ([]model.AuditLog, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if opts.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, opts.UserID)
		argIndex++
	}

	if opts.Action != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, opts.Action)
		argIndex++
	}

	if opts.From != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *opts.From)
		argIndex++
	}

	if opts.To != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *opts.To)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get paginated results
	offset := (opts.Page - 1) * opts.Limit
	query := fmt.Sprintf(`
		SELECT id, user_id, action, resource_type, resource_id, 
		       metadata, ip_address, user_agent, created_at, expires_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, opts.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	logs := []model.AuditLog{}
	for rows.Next() {
		var log model.AuditLog
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&metadataJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
			&log.ExpiresAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				log.Metadata = make(map[string]interface{})
			}
		} else {
			log.Metadata = make(map[string]interface{})
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, totalCount, nil
}
