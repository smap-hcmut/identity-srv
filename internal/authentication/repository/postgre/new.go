package postgre

import (
	"database/sql"
	"smap-api/internal/authentication/repository"
	pkgLog "smap-api/pkg/log"
)

type implRepository struct {
	l  pkgLog.Logger
	db *sql.DB
}

// New creates a new authentication repository
func New(l pkgLog.Logger, db *sql.DB) repository.Repository {
	return &implRepository{
		l:  l,
		db: db,
	}
}
