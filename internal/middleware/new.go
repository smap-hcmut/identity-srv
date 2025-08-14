package middleware

import (
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	pkgScope "github.com/nguyentantai21042004/smap-api/pkg/scope"
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
