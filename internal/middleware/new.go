package middleware

import (
	"identity-srv/config"
	"identity-srv/pkg/encrypter"
	pkgLog "identity-srv/pkg/log"
	pkgScope "identity-srv/pkg/scope"
)

type Middleware struct {
	l            pkgLog.Logger
	jwtManager   pkgScope.Manager
	cookieConfig config.CookieConfig
	InternalKey  string
	config       *config.Config
	encrypter    encrypter.Encrypter
}

func New(l pkgLog.Logger, jwtManager pkgScope.Manager, cookieConfig config.CookieConfig, internalKey string, cfg *config.Config, enc encrypter.Encrypter) Middleware {
	return Middleware{
		l:            l,
		jwtManager:   jwtManager,
		cookieConfig: cookieConfig,
		InternalKey:  internalKey,
		config:       cfg,
		encrypter:    enc,
	}
}
