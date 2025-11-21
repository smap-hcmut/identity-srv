package usecase

import (
	"time"

	"smap-project/internal/project"
	"smap-project/internal/project/repository"
	pkgLog "smap-project/pkg/log"
)

type usecase struct {
	l     pkgLog.Logger
	repo  repository.Repository
	clock func() time.Time
}

func New(l pkgLog.Logger, repo repository.Repository) project.UseCase {
	return &usecase{
		l:     l,
		repo:  repo,
		clock: time.Now,
	}
}
