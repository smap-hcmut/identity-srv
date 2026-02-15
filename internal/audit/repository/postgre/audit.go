package postgre

import (
	"context"
	"fmt"
	"identity-srv/internal/audit"
	"identity-srv/internal/audit/repository"
	"identity-srv/internal/model"
)

// BatchInsert inserts multiple audit events into the database
func (r *implRepository) BatchInsert(ctx context.Context, events []audit.AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	query, args := r.buildBatchInsertQuery(events)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch insert audit logs: %w", err)
	}

	return nil
}

// DeleteExpired deletes audit logs older than the retention period
func (r *implRepository) DeleteExpired(ctx context.Context) (int64, error) {
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
func (r *implRepository) Query(ctx context.Context, opts repository.QueryOptions) ([]model.AuditLog, int, error) {
	// Build filter
	whereClause, args, argIndex := r.buildQueryFilter(opts)

	// Get total count
	countQuery := r.buildCountQuery(whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get paginated results
	offset := (opts.Page - 1) * opts.Limit
	query := r.buildPaginatedQuery(whereClause, argIndex)
	args = append(args, opts.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	logs := []model.AuditLog{}
	for rows.Next() {
		log, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, totalCount, nil
}
