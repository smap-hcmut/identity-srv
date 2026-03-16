package http

import (
	"identity-srv/internal/audit/repository"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/discord"
	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/middleware"
)

// Handler defines the HTTP handler interface for Audit.
type Handler interface {
	RegisterRoutes(r *gin.RouterGroup, mw *middleware.Middleware)
}

type handler struct {
	l       log.Logger
	repo    repository.Repository
	discord discord.IDiscord
}

func New(l log.Logger, repo repository.Repository, discord discord.IDiscord) Handler {
	return handler{
		l:       l,
		repo:    repo,
		discord: discord,
	}
}
