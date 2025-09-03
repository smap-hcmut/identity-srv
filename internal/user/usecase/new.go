package usecase

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
)

type implUsecase struct {
	l      log.Logger
	repo   user.Repository
	roleUC role.UseCase
	clock  func() time.Time
}

var _ user.UseCase = &implUsecase{}

func New(l log.Logger, repo user.Repository, roleUC role.UseCase) user.UseCase {
	return &implUsecase{
		l:      l,
		repo:   repo,
		roleUC: roleUC,
		clock:  time.Now,
	}
}
