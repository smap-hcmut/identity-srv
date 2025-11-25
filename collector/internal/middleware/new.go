package middleware

import (
<<<<<<<< HEAD:identity/internal/middleware/new.go
	pkgLog "smap-api/pkg/log"
	pkgScope "smap-api/pkg/scope"
========
	pkgLog "smap-collector/pkg/log"
	pkgScope "smap-collector/pkg/scope"
>>>>>>>> 9c65a15b02994a6cc9940a129c9a3c4f61fd0697:collector/internal/middleware/new.go
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
