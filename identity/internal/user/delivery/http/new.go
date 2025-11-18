package http

import (
	"smap-api/internal/user"
	pkgLog "smap-api/pkg/log"
)

type Handler struct {
	l  pkgLog.Logger
	uc user.UseCase
}

func New(l pkgLog.Logger, uc user.UseCase) Handler {
	return Handler{
		l:  l,
		uc: uc,
	}
}
