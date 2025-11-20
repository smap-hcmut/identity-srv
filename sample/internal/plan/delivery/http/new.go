package http

import (
	"smap-api/internal/plan"
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
}

type handler struct {
	l  pkgLog.Logger
	uc plan.UseCase
}

func New(l pkgLog.Logger, uc plan.UseCase) Handler {
	h := handler{
		l:  l,
		uc: uc,
	}
	return h
}

