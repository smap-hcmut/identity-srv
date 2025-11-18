package usecase

import (
	"smap-api/internal/authentication"
	"smap-api/internal/authentication/delivery/rabbitmq/producer"
	"smap-api/internal/user"
	"smap-api/pkg/encrypter"
	pkgLog "smap-api/pkg/log"
	"smap-api/pkg/scope"
	"time"
)

type implUsecase struct {
	l       pkgLog.Logger
	prod    producer.Producer
	scope   scope.Manager
	encrypt encrypter.Encrypter
	userUC  user.UseCase
	clock   func() time.Time
}

func New(l pkgLog.Logger, prod producer.Producer, scope scope.Manager, encrypt encrypter.Encrypter, userUC user.UseCase) authentication.UseCase {
	return &implUsecase{
		l:       l,
		prod:    prod,
		scope:   scope,
		encrypt: encrypt,
		userUC:  userUC,
		clock:   time.Now,
	}
}
