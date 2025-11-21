package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapAuthRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.POST("/register", h.Register)
	r.POST("/send-otp", h.SendOTP)
	r.POST("/verify-otp", h.VerifyOTP)
	r.POST("/login", h.Login)
}
