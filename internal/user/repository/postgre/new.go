package postgres

import (
	"database/sql"
	"time"

	"identity-srv/internal/user/repository"

	"github.com/smap-hcmut/shared-libs/go/log"
)

type implRepository struct {
	l     log.Logger
	db    *sql.DB
	clock func() time.Time
}

var _ repository.Repository = &implRepository{}

func New(l log.Logger, db *sql.DB) *implRepository {
	return &implRepository{
		l:     l,
		db:    db,
		clock: time.Now,
	}
}
