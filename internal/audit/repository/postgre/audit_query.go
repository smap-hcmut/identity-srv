package postgre

import (
	"fmt"
	"smap-api/internal/audit"
	"smap-api/internal/audit/repository"
	"strings"
)

// buildBatchInsertQuery builds the batch insert SQL query and arguments
func (r *implRepository) buildBatchInsertQuery(events []audit.AuditEvent) (string, []interface{}) {
	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*8)

	for i, event := range events {
		metadataJSON := r.marshalMetadata(event.Metadata)

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

	query := fmt.Sprintf(`
		INSERT INTO audit_logs (
			user_id, action, resource_type, resource_id, 
			metadata, ip_address, user_agent, created_at, expires_at
		) VALUES %s
	`, strings.Join(valueStrings, ","))

	return query, valueArgs
}

// buildQueryFilter builds the WHERE clause and arguments for the Query method
func (r *implRepository) buildQueryFilter(opts repository.QueryOptions) (string, []interface{}, int) {
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

	return whereClause, args, argIndex
}

// buildCountQuery builds the count query with the given WHERE clause
func (r *implRepository) buildCountQuery(whereClause string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
}

// buildPaginatedQuery builds the paginated SELECT query
func (r *implRepository) buildPaginatedQuery(whereClause string, argIndex int) string {
	return fmt.Sprintf(`
		SELECT id, user_id, action, resource_type, resource_id, 
		       metadata, ip_address, user_agent, created_at, expires_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
}
