package mongo

import (
	"time"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/mongo"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/util"
)

type implRepository struct {
	l     log.Logger
	db    mongo.Database
	clock func() time.Time
}

func New(
	l log.Logger,
	db mongo.Database,
) repository.Repository {
	now := util.Now
	return &implRepository{
		l:     l,
		db:    db,
		clock: now,
	}
}
