package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/auth"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
)

type Handler interface {
	Register(c *gin.Context)
	SendOTP(c *gin.Context)
	VerifyOTP(c *gin.Context)
	Login(c *gin.Context)
	SocialLogin(c *gin.Context)
	SocialCallback(c *gin.Context)
}

type handler struct {
	l  pkgLog.Logger
	uc auth.UseCase
	d  *discord.Discord
}

func New(l pkgLog.Logger, uc auth.UseCase, d *discord.Discord) Handler {
	h := handler{
		l:  l,
		uc: uc,
		d:  d,
	}
	return h
}
