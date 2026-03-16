package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/response"
)

// AdminOnly validates that the authenticated user has admin role.
func (m Middleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		payload, ok := auth.GetPayloadFromContext(ctx)
		if !ok {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		sc := auth.NewScope(payload)

		if !sc.IsAdmin() {
			response.Forbidden(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
