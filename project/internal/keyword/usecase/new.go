package usecase

import (
	"smap-project/internal/keyword"
	pkgLog "smap-project/pkg/log"
	"time"
)

type usecase struct {
	l     pkgLog.Logger
	clock func() time.Time
}

// New creates a new keyword use case
func New(l pkgLog.Logger) keyword.UseCase {
	return &usecase{
		l:     l,
		clock: time.Now,
	}
}
