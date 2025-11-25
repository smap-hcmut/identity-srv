package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapSubscriptionRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware) {
	r.GET("", mw.Auth(), h.List)
	r.GET("/page", mw.Auth(), h.Get)
	r.GET("/me", mw.Auth(), h.GetMySubscription)
	r.GET("/:id", mw.Auth(), h.Detail)
	r.POST("", mw.Auth(), h.Create)
	r.PUT("/:id", mw.Auth(), h.Update)
	r.DELETE("/:id", mw.Auth(), h.Delete)
	r.POST("/:id/cancel", mw.Auth(), h.Cancel)
}
