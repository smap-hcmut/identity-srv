package usecase

import (
	"smap-api/internal/user"
	"smap-api/internal/user/repository"
	pkgLog "smap-api/pkg/log"
	"time"
)

type usecase struct {
	l     pkgLog.Logger
	repo  repository.Repository
	clock func() time.Time
}

func New(l pkgLog.Logger, repo repository.Repository) user.UseCase {
	return &usecase{
		l:     l,
		repo:  repo,
		clock: time.Now,
	}
}
