package postgre

import (
	"database/sql"
	"smap-api/internal/audit/repository"
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
