package http

import (
	"identity-srv/config"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"

	"github.com/smap-hcmut/shared-libs/go/discord"
	"github.com/smap-hcmut/shared-libs/go/log"
)

type handler struct {
	l            log.Logger
	uc           authentication.UseCase
	discord      discord.IDiscord
	cookieConfig config.CookieConfig
	config       *config.Config
}

func New(l log.Logger, uc authentication.UseCase, discord discord.IDiscord, cfg *config.Config) handler {
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
