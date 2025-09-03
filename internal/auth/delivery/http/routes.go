package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/internal/middleware"
)

func MapAuthRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.POST("/register", h.Register)
	r.POST("/send-otp", h.SendOTP)
	r.POST("/verify-otp", h.VerifyOTP)
	r.POST("/login", h.Login)
	r.POST(":provider", h.SocialLogin)
	r.GET("/:provider/callback", h.SocialCallback)
}
