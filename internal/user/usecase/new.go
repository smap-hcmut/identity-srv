package usecase

import (
	"identity-srv/internal/user"
	"identity-srv/internal/user/repository"
	"time"

	"github.com/smap-hcmut/shared-libs/go/encrypter"
	"github.com/smap-hcmut/shared-libs/go/log"
)

type usecase struct {
	l       log.Logger
	encrypt encrypter.Encrypter
	repo    repository.Repository
	clock   func() time.Time
}

func New(l log.Logger, encrypt encrypter.Encrypter, repo repository.Repository) user.UseCase {
	return &usecase{
		l:       l,
		encrypt: encrypt,
		repo:    repo,
		clock:   time.Now,
	}
}
