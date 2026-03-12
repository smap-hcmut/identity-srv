package postgre

import (
	"database/sql"
	"identity-srv/internal/authentication/repository"

	"github.com/smap-hcmut/shared-libs/go/log"
)

type implRepository struct {
	l  log.Logger
	db *sql.DB
}

// New creates a new authentication repository
func New(l log.Logger, db *sql.DB) repository.Repository {
	return &implRepository{
		l:  l,
		db: db,
	}
}
