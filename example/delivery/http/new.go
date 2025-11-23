package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
)

type Handler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Detail(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	UpdateAttendance(c *gin.Context)
}

type handler struct {
	l  log.Logger
	uc event.UseCase
}

func New(l log.Logger, uc event.UseCase) Handler {
	return handler{
		l:  l,
		uc: uc,
	}
}
