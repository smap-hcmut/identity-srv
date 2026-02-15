package usecase

import (
	"identity-srv/internal/user"
	"identity-srv/internal/user/repository"
	"identity-srv/pkg/encrypter"
	pkgLog "identity-srv/pkg/log"
	"time"
)

type usecase struct {
	l       pkgLog.Logger
	encrypt encrypter.Encrypter
	repo    repository.Repository
	clock   func() time.Time
}

func New(l pkgLog.Logger, encrypt encrypter.Encrypter, repo repository.Repository) user.UseCase {
	return &usecase{
		l:       l,
		encrypt: encrypt,
		repo:    repo,
		clock:   time.Now,
	}
}
