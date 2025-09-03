package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
)

type Handler interface {
	DetailMe(c *gin.Context)
	Detail(c *gin.Context)
	UpdateAvatar(c *gin.Context)
}

type handler struct {
	l  pkgLog.Logger
	uc user.UseCase
	d  *discord.Discord
}

func New(l pkgLog.Logger, uc user.UseCase, d *discord.Discord) Handler {
	h := handler{
		l:  l,
		uc: uc,
		d:  d,
	}
	return h
}
