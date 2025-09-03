package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/middleware"
)

func MapUserRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.Use(mw.Auth())
	r.GET("/detail/me", h.DetailMe)
	r.GET("/detail/:id", h.Detail)
	r.PATCH("/avatar", h.UpdateAvatar)
}
