package middleware

import (
	"smap-api/config"
	pkgLog "smap-api/pkg/log"
	pkgScope "smap-api/pkg/scope"
)

type Middleware struct {
	l            pkgLog.Logger
	jwtManager   pkgScope.Manager
	cookieConfig config.CookieConfig
}

func New(l pkgLog.Logger, jwtManager pkgScope.Manager, cookieConfig config.CookieConfig) Middleware {
	return Middleware{
		l:            l,
		jwtManager:   jwtManager,
		cookieConfig: cookieConfig,
	}
}
