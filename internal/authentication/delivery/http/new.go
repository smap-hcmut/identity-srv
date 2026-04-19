package http

import (
	"identity-srv/config"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/discord"
	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/middleware"
)

type Handler interface {
	RegisterRoutes(r *gin.RouterGroup, mw *middleware.Middleware)
}

type handler struct {
	l            log.Logger
	uc           authentication.UseCase
	discord      discord.IDiscord
	cookieConfig config.CookieConfig
	config       *config.Config
	stateSecret  string // JWT secret reused for HMAC-signing OAuth state
}

func New(l log.Logger, uc authentication.UseCase, discord discord.IDiscord, cfg *config.Config) Handler {
	return handler{
		l:            l,
		uc:           uc,
		discord:      discord,
		cookieConfig: cfg.Cookie,
		config:       cfg,
		stateSecret:  cfg.JWT.SecretKey,
	}
}

func (h handler) isDevelopmentMode() bool {
	return h.config.Environment.Name == string(model.EnvironmentDevelopment)
}
