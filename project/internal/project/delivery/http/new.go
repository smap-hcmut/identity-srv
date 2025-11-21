package http

import (
	"smap-project/internal/project"
	pkgLog "smap-project/pkg/log"

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
	uc project.UseCase
}

func New(l pkgLog.Logger, uc project.UseCase) Handler {
	h := handler{
		l:  l,
		uc: uc,
	}
	return h
}
