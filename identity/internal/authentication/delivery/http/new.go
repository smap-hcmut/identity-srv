package http

import (
	"smap-api/internal/authentication"
	"smap-api/pkg/discord"
	pkgLog "smap-api/pkg/log"
)

type handler struct {
	l       pkgLog.Logger
	uc      authentication.UseCase
	discord *discord.Discord
}

func New(l pkgLog.Logger, uc authentication.UseCase, discord *discord.Discord) handler {
	return handler{
		l:       l,
		uc:      uc,
		discord: discord,
	}
}
