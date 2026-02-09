package http

import (
	"smap-api/internal/authentication/usecase"
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapAuthRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware, rateLimiter *usecase.RateLimiter) {
	// Create rate limit middleware
	rateLimitMW := mw.LoginRateLimit(rateLimiter)

	// Public routes with rate limiting
	r.GET("/login", rateLimitMW, h.OAuthLogin)
	r.GET("/callback", rateLimitMW, h.OAuthCallback)

	// JWKS endpoint (public - for JWT verification by other services)
	r.GET("/.well-known/jwks.json", h.JWKS)

	// Protected routes (require authentication)
	r.POST("/logout", mw.Auth(), h.Logout)
	r.GET("/me", mw.Auth(), h.GetMe)

	// Internal routes (require X-Service-Key header)
	internal := r.Group("/internal", mw.ServiceAuth())
	{
		internal.POST("/validate", h.ValidateToken)
		internal.POST("/revoke-token", mw.Admin(), h.RevokeToken) // Also requires ADMIN role
		internal.GET("/users/:id", h.GetUserByID)
	}
}
