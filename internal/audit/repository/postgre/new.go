package postgre

import (
	"database/sql"
	"smap-api/internal/audit/repository"
	pkgLog "smap-api/pkg/log"
)

var _ repository.Repository = (*implRepository)(nil)

type implRepository struct {
	l  pkgLog.Logger
	db *sql.DB
}

// New creates a new audit repository
func New(l pkgLog.Logger, db *sql.DB) repository.Repository {
	return &implRepository{
		l:  l,
		db: db,
	}
}
