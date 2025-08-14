package http

import "github.com/gin-gonic/gin"

type Handler interface {
	Create(c *gin.Context)
	Detail(c *gin.Context)
	Get(c *gin.Context)
}
