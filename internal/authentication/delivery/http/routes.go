package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapAuthRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware) {
	// Public routes
	r.GET("/login", h.OAuthLogin)
	r.GET("/callback", h.OAuthCallback)

	// JWKS endpoint (public - for JWT verification by other services)
	r.GET("/.well-known/jwks.json", h.JWKS)

	// Protected routes (require authentication)
	r.POST("/logout", mw.Auth(), h.Logout)
	r.GET("/me", mw.Auth(), h.GetMe)
}
