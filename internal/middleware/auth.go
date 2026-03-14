package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/response"
)

// InternalAuth validates the internal key from the Authorization header (Bearer <key> or raw key).
// If internalKey is empty, all requests are rejected with 401.
func (m Middleware) InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// Support both "Bearer <key>" and raw key
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
		if m.internalKey == "" || tokenString != m.internalKey {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// Admin validates that the user has admin role
func (m Middleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get scope from context (should be set by Auth middleware)
		sc := auth.GetScopeFromContext(ctx)
		if sc.UserID == "" {
			response.Forbidden(c)
			c.Abort()
			return
		}

		// Check if user has admin role
		if !sc.IsAdmin() {
			response.Forbidden(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
