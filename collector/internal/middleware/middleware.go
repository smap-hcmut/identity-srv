package middleware

import (
	"strings"

<<<<<<<< HEAD:identity/internal/middleware/middleware.go
	"smap-api/pkg/response"
	"smap-api/pkg/scope"
========
	"smap-collector/pkg/response"
	"smap-collector/pkg/scope"
>>>>>>>> 9c65a15b02994a6cc9940a129c9a3c4f61fd0697:collector/internal/middleware/middleware.go

	"github.com/gin-gonic/gin"
)

func (m Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
		if tokenString == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		payload, err := m.jwtManager.Verify(tokenString)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = scope.SetPayloadToContext(ctx, payload)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
