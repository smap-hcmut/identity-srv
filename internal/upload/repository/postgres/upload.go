package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/upload"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

func (r *repository) Create(ctx context.Context, sc models.Scope, opts upload.CreateOptions) (models.Upload, error) {
	query := `
		INSERT INTO uploads (
			id, bucket_name, object_name, original_name, size, content_type, 
			etag, url, source, created_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, bucket_name, object_name, original_name, size, content_type, 
			etag, url, source, created_user_id, created_at, updated_at
	`

	var upload models.Upload
	err := r.database.QueryRowContext(ctx, query,
		opts.Upload.ID,
		opts.Upload.BucketName,
		opts.Upload.ObjectName,
		opts.Upload.OriginalName,
		opts.Upload.Size,
		opts.Upload.ContentType,
		opts.Upload.Etag,
		opts.Upload.URL,
		opts.Upload.Source,
		opts.Upload.CreatedUserID,
		opts.Upload.CreatedAt,
		opts.Upload.UpdatedAt,
	).Scan(
		&upload.ID,
		&upload.BucketName,
		&upload.ObjectName,
		&upload.OriginalName,
		&upload.Size,
		&upload.ContentType,
		&upload.Etag,
		&upload.URL,
		&upload.Source,
		&upload.CreatedUserID,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	)

	if err != nil {
		r.l.Error(ctx, "Failed to create upload", "error", err)
		return models.Upload{}, fmt.Errorf("failed to create upload: %w", err)
	}

	return upload, nil
}

func (r *repository) Detail(ctx context.Context, sc models.Scope, ID string) (models.Upload, error) {
	query := `
		SELECT id, bucket_name, object_name, original_name, size, content_type, 
			etag, url, source, created_user_id, created_at, updated_at
		FROM uploads 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var upload models.Upload
	err := r.database.QueryRowContext(ctx, query, ID).Scan(
		&upload.ID,
		&upload.BucketName,
		&upload.ObjectName,
		&upload.OriginalName,
		&upload.Size,
		&upload.ContentType,
		&upload.Etag,
		&upload.URL,
		&upload.Source,
		&upload.CreatedUserID,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.l.Warn(ctx, "Upload not found", "id", ID)
			return models.Upload{}, fmt.Errorf("upload not found")
		}
		r.l.Error(ctx, "Failed to get upload detail", "error", err, "id", ID)
		return models.Upload{}, fmt.Errorf("failed to get upload detail: %w", err)
	}

	return upload, nil
}

func (r *repository) Get(ctx context.Context, sc models.Scope, opts upload.GetOptions) ([]models.Upload, paginator.Paginator, error) {
	// Build WHERE clause
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 1

	if opts.Filter.ID != nil {
		whereClause += fmt.Sprintf(" AND id = $%d", argIndex)
		args = append(args, *opts.Filter.ID)
		argIndex++
	}

	if opts.Filter.BucketName != nil {
		whereClause += fmt.Sprintf(" AND bucket_name = $%d", argIndex)
		args = append(args, *opts.Filter.BucketName)
		argIndex++
	}

	if opts.Filter.OriginalName != nil {
		whereClause += fmt.Sprintf(" AND original_name ILIKE $%d", argIndex)
		args = append(args, "%"+*opts.Filter.OriginalName+"%")
		argIndex++
	}

	if opts.Filter.Source != nil {
		whereClause += fmt.Sprintf(" AND source = $%d", argIndex)
		args = append(args, *opts.Filter.Source)
		argIndex++
	}

	if opts.Filter.CreatedUserID != nil {
		whereClause += fmt.Sprintf(" AND created_user_id = $%d", argIndex)
		args = append(args, *opts.Filter.CreatedUserID)
		argIndex++
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM uploads %s", whereClause)
	var total int64
	err := r.database.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.l.Error(ctx, "Failed to count uploads", "error", err)
		return nil, paginator.Paginator{}, fmt.Errorf("failed to count uploads: %w", err)
	}

	// Calculate pagination
	page := opts.PagQuery.Page
	if page < 1 {
		page = 1
	}
	limit := opts.PagQuery.Limit
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * int(limit)

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, bucket_name, object_name, original_name, size, content_type, 
			etag, url, source, created_user_id, created_at, updated_at
		FROM uploads 
		%s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)
	rows, err := r.database.QueryContext(ctx, query, args...)
	if err != nil {
		r.l.Error(ctx, "Failed to get uploads", "error", err)
		return nil, paginator.Paginator{}, fmt.Errorf("failed to get uploads: %w", err)
	}
	defer rows.Close()

	var uploads []models.Upload
	for rows.Next() {
		var upload models.Upload
		err := rows.Scan(
			&upload.ID,
			&upload.BucketName,
			&upload.ObjectName,
			&upload.OriginalName,
			&upload.Size,
			&upload.ContentType,
			&upload.Etag,
			&upload.URL,
			&upload.Source,
			&upload.CreatedUserID,
			&upload.CreatedAt,
			&upload.UpdatedAt,
		)
		if err != nil {
			r.l.Error(ctx, "Failed to scan upload row", "error", err)
			continue
		}
		uploads = append(uploads, upload)
	}

	if err = rows.Err(); err != nil {
		r.l.Error(ctx, "Error iterating upload rows", "error", err)
		return nil, paginator.Paginator{}, fmt.Errorf("error iterating upload rows: %w", err)
	}

	// Create paginator
	paginator := paginator.Paginator{
		Total:       total,
		Count:       int64(len(uploads)),
		PerPage:     limit,
		CurrentPage: page,
	}

	return uploads, paginator, nil
}
