package postgre

import (
	"database/sql"
	"identity-srv/internal/authentication/repository"
	pkgLog "identity-srv/pkg/log"
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
