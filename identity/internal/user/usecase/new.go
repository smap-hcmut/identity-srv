package usecase

import (
	"smap-api/internal/user"
	"smap-api/internal/user/repository"
	"smap-api/pkg/encrypter"
	pkgLog "smap-api/pkg/log"
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
