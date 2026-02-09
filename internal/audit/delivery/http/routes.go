package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapAuditRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware) {
	// Protected routes (require authentication + ADMIN role)
	r.GET("", mw.Auth(), mw.Admin(), h.GetAuditLogs)
}
