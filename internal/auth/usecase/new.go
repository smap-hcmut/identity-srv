package usecase

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/appconfig/oauth"
	"github.com/nguyentantai21042004/smap-api/internal/auth"
	"github.com/nguyentantai21042004/smap-api/internal/auth/delivery/rabbitmq/producer"
	"github.com/nguyentantai21042004/smap-api/internal/core/smtp"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/internal/session"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/scope"
)

type implUsecase struct {
	l         log.Logger
	prod      producer.Producer
	encrypt   encrypter.Encrypter
	oauth     oauth.OauthConfig
	scope     scope.Manager
	smtp      smtp.UseCase
	userUC    user.UseCase
	roleUC    role.UseCase
	sessionUC session.UseCase
	clock     func() time.Time
}

var _ auth.UseCase = &implUsecase{}

func New(l log.Logger,
	prod producer.Producer,
	encrypt encrypter.Encrypter,
	oauth oauth.OauthConfig,
	scope scope.Manager,
	smtp smtp.UseCase,
	userUC user.UseCase,
	roleUC role.UseCase,
	sessionUC session.UseCase,
) auth.UseCase {
	return &implUsecase{
		l:         l,
		prod:      prod,
		encrypt:   encrypt,
		oauth:     oauth,
		scope:     scope,
		smtp:      smtp,
		userUC:    userUC,
		roleUC:    roleUC,
		sessionUC: sessionUC,
		clock:     time.Now,
	}
}
