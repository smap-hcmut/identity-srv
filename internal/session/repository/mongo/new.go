package mongo

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/session"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
)

type implRepository struct {
	l     log.Logger
	db    mongo.Database
	clock func() time.Time
}

func New(l log.Logger, db mongo.Database) session.Repository {
	return &implRepository{
		l:     l,
		db:    db,
		clock: time.Now,
	}
}