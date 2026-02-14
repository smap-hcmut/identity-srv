package http

import (
	"smap-api/config"
	"smap-api/internal/authentication"
	"smap-api/internal/model"
	"smap-api/pkg/discord"
	pkgLog "smap-api/pkg/log"
)

type handler struct {
	l            pkgLog.Logger
	uc           authentication.UseCase
	discord      *discord.Discord
	cookieConfig config.CookieConfig
	config       *config.Config
}

func New(l pkgLog.Logger, uc authentication.UseCase, discord *discord.Discord, cfg *config.Config) handler {
	return handler{
		l:            l,
		uc:           uc,
		discord:      discord,
		cookieConfig: cfg.Cookie,
		config:       cfg,
	}
}

func (h handler) isDevelopmentMode() bool {
	return h.config.Environment.Name == string(model.EnvironmentDevelopment)
}
