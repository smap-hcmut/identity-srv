package middleware

import (
	pkgLog "smap-collector/pkg/log"
	pkgScope "smap-collector/pkg/scope"
)

type Middleware struct {
	l          pkgLog.Logger
	jwtManager pkgScope.Manager
}

func New(l pkgLog.Logger, jwtManager pkgScope.Manager) Middleware {
	return Middleware{
		l:          l,
		jwtManager: jwtManager,
	}
}
