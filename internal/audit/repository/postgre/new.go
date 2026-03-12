package postgre

import (
	"database/sql"
	"identity-srv/internal/audit/repository"

	"github.com/smap-hcmut/shared-libs/go/log"
)

var _ repository.Repository = (*implRepository)(nil)

type implRepository struct {
	l  log.Logger
	db *sql.DB
}

// New creates a new audit repository
func New(l log.Logger, db *sql.DB) repository.Repository {
	return &implRepository{
		l:  l,
		db: db,
	}
}
