package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// MapUserRoutes maps user routes to the router
func MapUserRoutes(group *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	// Public/authenticated user endpoints
	group.Use(mw.Auth())
	{
		// Current user endpoints
		group.GET("/me", h.GetMe)
		group.PUT("/me", h.UpdateProfile)
		group.POST("/me/change-password", h.ChangePassword)
	}

	// Admin-only endpoints
	adminGroup := group.Group("")
	adminGroup.Use(mw.AdminOnly())
	{
		adminGroup.GET("", h.List)
		adminGroup.GET("/page", h.Get)
		adminGroup.GET("/:id", h.GetDetail)
	}
}
