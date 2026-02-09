package middleware

import (
	"net/http"
	"smap-api/internal/authentication/usecase"

	"github.com/gin-gonic/gin"
)

// LoginRateLimit creates a middleware that applies rate limiting to login endpoints
func (m *Middleware) LoginRateLimit(rateLimiter *usecase.RateLimiter) gin.HandlerFunc {
	if rateLimiter == nil {
		// If rate limiter not configured, return no-op middleware
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()

		// Check if IP is blocked
		blocked, err := rateLimiter.IsBlocked(ctx, ip)
		if err != nil {
			// Log error but don't block request on Redis failure
			c.Next()
			return
		}

		if blocked {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Too many failed login attempts. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
