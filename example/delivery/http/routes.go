package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/gma-vietnam/tanca-connect/internal/middleware"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.Use(mw.Auth())
	r.POST("", h.Create)
	r.GET("/:event_id/:id", h.Detail)
	r.PUT("/:event_id/:id", h.Update)
	r.DELETE("/:event_id/:id", h.Delete)
	r.PATCH("/attendance/:event_id/:id", h.UpdateAttendance)
}

func MapCalendarRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.Use(mw.Auth())
	r.GET("", h.List)
}
