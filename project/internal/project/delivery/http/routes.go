package http

import (
	"smap-project/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapProjectRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	// All routes require authentication
	r.GET("", mw.Auth(), h.List)
	r.GET("/page", mw.Auth(), h.Get)
	r.GET("/:id", mw.Auth(), h.Detail)
	r.POST("", mw.Auth(), h.Create)
	r.PUT("/:id", mw.Auth(), h.Update)
	r.DELETE("/:id", mw.Auth(), h.Delete)
}
