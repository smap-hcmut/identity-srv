package http

import (
	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/middleware"
)

func (h handler) RegisterRoutes(r *gin.RouterGroup, mw *middleware.Middleware) {
	// Public routes
	r.GET("/login", h.OAuthLogin)
	r.GET("/callback", h.OAuthCallback)

	// Protected routes (require authentication)
	r.POST("/logout", mw.Auth(), h.Logout)
	r.GET("/me", mw.Auth(), h.GetMe)

	// Internal routes (require X-Internal-Key header)
	internal := r.Group("/internal")
	internal.Use(mw.InternalAuth())
	{
		internal.POST("/validate", h.ValidateToken)
		internal.POST("/revoke-token", mw.Auth(), mw.AdminOnly(), h.RevokeToken)
		internal.GET("/users/:id", h.GetUserByID)
	}
}
