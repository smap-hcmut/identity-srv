package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	"github.com/nguyentantai21042004/smap-api/pkg/log"
)

type Handler interface {
	Create(c *gin.Context)
	GetOne(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Get(c *gin.Context)
}

type handler struct {
	l  log.Logger
	uc role.UseCase
	d  *discord.Discord
}

func New(l log.Logger, uc role.UseCase, d *discord.Discord) Handler {
	return handler{
		l:  l,
		uc: uc,
		d:  d,
	}
}

