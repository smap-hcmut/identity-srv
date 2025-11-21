package usecase

import (
	"smap-api/internal/plan"
	"smap-api/internal/plan/repository"
	pkgLog "smap-api/pkg/log"
	"time"
)

type usecase struct {
	l     pkgLog.Logger
	repo  repository.Repository
	clock func() time.Time
}

func New(l pkgLog.Logger, repo repository.Repository) plan.UseCase {
	return &usecase{
		l:     l,
		repo:  repo,
		clock: time.Now,
	}
}

