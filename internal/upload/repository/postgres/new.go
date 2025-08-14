package postgres

import (
	"database/sql"

	"github.com/nguyentantai21042004/smap-api/internal/upload"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
)

type repository struct {
	l        pkgLog.Logger
	database *sql.DB
}

func New(l pkgLog.Logger, database *sql.DB) upload.Repository {
	return &repository{
		l:        l,
		database: database,
	}
}
