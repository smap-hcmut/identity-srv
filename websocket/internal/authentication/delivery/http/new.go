package http

import (
	"smap-api/internal/authentication"
	"smap-api/pkg/discord"
	pkgLog "smap-api/pkg/log"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Register(c *gin.Context)
	SendOTP(c *gin.Context)
	VerifyOTP(c *gin.Context)
	Login(c *gin.Context)
}

type handler struct {
	l  pkgLog.Logger
	uc authentication.UseCase
	d  *discord.Discord
}

func New(l pkgLog.Logger, uc authentication.UseCase, d *discord.Discord) Handler {
	h := handler{
		l:  l,
		uc: uc,
		d:  d,
	}
	return h
}
