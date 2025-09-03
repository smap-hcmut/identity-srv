package mongo

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"github.com/nguyentantai21042004/smap-api/pkg/util"
)

type implRepository struct {
	l     log.Logger
	db    mongo.Database
	clock func() time.Time
}

var _ role.Repository = implRepository{}

func New(
	l log.Logger,
	db mongo.Database,
) role.Repository {
	now := util.Now
	return &implRepository{
		l:     l,
		db:    db,
		clock: now,
	}
}
