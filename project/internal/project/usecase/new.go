package usecase

import (
	"time"

	"smap-project/internal/keyword"
	"smap-project/internal/project"
	"smap-project/internal/project/repository"
	pkgLog "smap-project/pkg/log"
)

type usecase struct {
	l              pkgLog.Logger
	repo           repository.Repository
	clock          func() time.Time
	keywordService keyword.Service
}

func New(l pkgLog.Logger, repo repository.Repository, keywordSvc keyword.Service) project.UseCase {
	return &usecase{
		l:              l,
		repo:           repo,
		clock:          time.Now,
		keywordService: keywordSvc,
	}
}
