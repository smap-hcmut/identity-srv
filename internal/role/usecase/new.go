package usecase

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
)

type implUsecase struct {
	l     log.Logger
	repo  role.Repository
	clock func() time.Time
}

var _ role.UseCase = &implUsecase{}

func New(l log.Logger, repo role.Repository) role.UseCase {
	return &implUsecase{
		l:     l,
		repo:  repo,
		clock: time.Now,
	}
}
