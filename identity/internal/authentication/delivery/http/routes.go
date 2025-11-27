package http

import (
	"smap-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MapAuthRoutes(r *gin.RouterGroup, h handler, mw middleware.Middleware) {
	r.POST("/register", h.Register)
	r.POST("/send-otp", h.SendOTP)
	r.POST("/verify-otp", h.VerifyOTP)
	r.POST("/login", h.Login)

	// Protected routes (require authentication)
	r.POST("/logout", mw.Auth(), h.Logout)
	r.GET("/me", mw.Auth(), h.GetMe)
}
