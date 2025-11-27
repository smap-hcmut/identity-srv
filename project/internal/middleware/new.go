package middleware

import (
	"smap-project/config"
	pkgLog "smap-project/pkg/log"
	pkgScope "smap-project/pkg/scope"
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
