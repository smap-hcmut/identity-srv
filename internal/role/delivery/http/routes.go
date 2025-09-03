package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/middleware"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	// Role routes
	
	r.Use(mw.Auth())
	r.POST("", h.Create)
	r.GET("/:id", h.GetOne)   // For public API (get single)
	r.PUT("", h.Update)
	r.DELETE("", h.Delete)
	r.GET("", h.Get)          // For public API (paginated list)
}