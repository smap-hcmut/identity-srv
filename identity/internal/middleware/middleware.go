package middleware

import (
	"strings"

	"smap-api/pkg/response"
	"smap-api/pkg/scope"

	"github.com/gin-gonic/gin"
)

func (m Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, attempt to read token from cookie (preferred method)
		tokenString, err := c.Cookie(m.cookieConfig.Name)

		// Fallback to Authorization header for backward compatibility
		// This allows existing clients to continue working during migration
		// TODO: Remove this fallback after all clients have migrated to cookie-based auth
		if err != nil || tokenString == "" {
			tokenString = strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
		}

		// If no token found in either location, return unauthorized
		if tokenString == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Verify JWT token
		payload, err := m.jwtManager.Verify(tokenString)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Set payload and scope in context for downstream handlers
		ctx := c.Request.Context()
		ctx = scope.SetPayloadToContext(ctx, payload)
		sc := scope.NewScope(payload)
		ctx = scope.SetScopeToContext(ctx, sc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
