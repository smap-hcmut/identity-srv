package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapAuthRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware) {
	// Public routes
	r.GET("/login", h.OAuthLogin)
	r.GET("/callback", h.OAuthCallback)

	// Protected routes (require authentication)
	r.POST("/logout", mw.Auth(), h.Logout)
	r.GET("/me", mw.Auth(), h.GetMe)

	// Internal routes (require X-Service-Key header)
	internal := r.Group("/internal") //, mw.ServiceAuth())
	{
		internal.POST("/validate", h.ValidateToken)
		internal.POST("/revoke-token", mw.Admin(), h.RevokeToken)
		internal.GET("/users/:id", h.GetUserByID)
	}
}
