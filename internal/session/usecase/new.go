package usecase

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/session"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
)

type implUsecase struct {
	l      log.Logger
	repo   session.Repository
	userUC user.UseCase
	clock  func() time.Time
}

var _ session.UseCase = &implUsecase{}

func New(l log.Logger, repo session.Repository, userUC user.UseCase) session.UseCase {
	return &implUsecase{
		l:      l,
		repo:   repo,
		userUC: userUC,
		clock:  time.Now,
	}
}

func (uc *implUsecase) SetUserUseCase(userUC user.UseCase) {
	uc.userUC = userUC
}
