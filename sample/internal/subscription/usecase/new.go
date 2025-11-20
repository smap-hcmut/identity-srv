package usecase

import (
	"smap-api/internal/plan"
	"smap-api/internal/subscription"
	"smap-api/internal/subscription/repository"
	pkgLog "smap-api/pkg/log"
	"time"
)

type usecase struct {
	l      pkgLog.Logger
	repo   repository.Repository
	planUC plan.UseCase
	clock  func() time.Time
}

func New(l pkgLog.Logger, repo repository.Repository, planUC plan.UseCase) subscription.UseCase {
	return &usecase{
		l:      l,
		repo:   repo,
		planUC: planUC,
		clock:  time.Now,
	}
}

