package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapPlanRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware) {
	r.GET("", h.List)
	r.GET("/page", h.Get)
	r.GET("/:id", h.Detail)
	r.POST("", mw.Auth(), h.Create)
	r.PUT("/:id", mw.Auth(), h.Update)
	r.DELETE("/:id", mw.Auth(), h.Delete)
}
