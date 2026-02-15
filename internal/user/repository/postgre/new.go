package postgres

import (
	"database/sql"
	"time"

	"identity-srv/internal/user/repository"
	pkgLog "identity-srv/pkg/log"
)

type implRepository struct {
	l     pkgLog.Logger
	db    *sql.DB
	clock func() time.Time
}

var _ repository.Repository = &implRepository{}

func New(l pkgLog.Logger, db *sql.DB) *implRepository {
	return &implRepository{
		l:     l,
		db:    db,
		clock: time.Now,
	}
}
