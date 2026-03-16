package http

import (
	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/middleware"
)

func (h handler) RegisterRoutes(r *gin.RouterGroup, mw *middleware.Middleware) {
	// Protected routes (require authentication + ADMIN role)
	r.GET("", mw.Auth(), mw.AdminOnly(), h.GetAuditLogs)
}
