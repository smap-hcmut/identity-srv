package http

import (
	"smap-api/internal/subscription"
	pkgLog "smap-api/pkg/log"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
	Detail(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Cancel(c *gin.Context)
	GetMySubscription(c *gin.Context)
}

type handler struct {
	l  pkgLog.Logger
	uc subscription.UseCase
}

func New(l pkgLog.Logger, uc subscription.UseCase) Handler {
	h := handler{
		l:  l,
		uc: uc,
	}
	return h
}

