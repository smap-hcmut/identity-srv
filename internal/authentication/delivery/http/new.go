package http

import (
	"identity-srv/config"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"
	"identity-srv/pkg/discord"
	pkgLog "identity-srv/pkg/log"
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
