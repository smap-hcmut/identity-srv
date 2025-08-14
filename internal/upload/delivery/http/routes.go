package http

import (
	"github.com/nguyentantai21042004/smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapUploadRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.POST("", mw.Auth(), h.Create)
	r.GET("", mw.Auth(), h.Get)
	r.GET("/:id", mw.Auth(), h.Detail)
}
