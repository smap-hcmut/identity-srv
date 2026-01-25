package middleware

import (
	"smap-api/pkg/response"
	"smap-api/pkg/scope"

	"github.com/gin-gonic/gin"
)

func (m Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, attempt to read token from cookie (preferred method)
		tokenString, err := c.Cookie(m.cookieConfig.Name)

		// If no token found in cookie, return unauthorized
		if err != nil || tokenString == "" {
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

func (m Middleware) InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, attempt to read token from cookie (preferred method)
		tokenString := c.GetHeader("Authorization")

		// If no token found in cookie, return unauthorized
		if tokenString == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		if tokenString != m.InternalKey {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
