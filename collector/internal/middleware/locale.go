package middleware

import (
	"smap-collector/pkg/locale"

	"github.com/gin-gonic/gin"
)

func (m Middleware) Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("lang")

		l := locale.ParseLang(h)

		ctx := c.Request.Context()
		ctx = locale.SetLocaleToContext(ctx, l)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
