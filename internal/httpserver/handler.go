package httpserver

import (
	"github.com/nguyentantai21042004/smap-api/internal/middleware"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	"github.com/nguyentantai21042004/smap-api/pkg/i18n"
	"github.com/nguyentantai21042004/smap-api/pkg/scope"

	// Import this to execute the init function in docs.go which setups the Swagger docs.
	_ "github.com/nguyentantai21042004/smap-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/pkg/response"
)

const (
	Api         = "/api/v1"
	InternalApi = "internal/api/v1"
)

func (srv HTTPServer) mapHandlers() error {
	discord, err := discord.New(srv.l, srv.discord, discord.DefaultConfig())
	if err != nil {
		return err
	}
	srv.gin.Use(middleware.Recovery(discord))

	// Health check endpoint
	srv.gin.GET("/health", srv.healthCheck)
	srv.gin.GET("/ready", srv.readyCheck)
	srv.gin.GET("/live", srv.liveCheck)

	// Swagger UI
	srv.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	scopeUC := scope.New(srv.jwtSecretKey)
	// internalKey, err := srv.encrypter.Encrypt(srv.internalKey)
	// if err != nil {
	// 	srv.l.Fatal(context.Background(), err)
	// 	return err
	// }

	i18n.Init()

	// Middleware
	mw := middleware.New(srv.l, scopeUC)

	// Apply locale middleware
	srv.gin.Use(mw.Locale())
	// api := srv.gin.Group(Api)

	return nil
}

// healthCheck handles health check requests
// @Summary Health Check
// @Description Check if the API is healthy
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "API is healthy"
// @Router /health [get]
func (srv HTTPServer) healthCheck(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "healthy",
		"message": "From Tan Tai API V1 With Love",
		"version": "1.0.0",
		"service": "smap-api",
	})
}

// readyCheck handles readiness check requests
// @Summary Readiness Check
// @Description Check if the API is ready to serve traffic
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "API is ready"
// @Router /ready [get]
func (srv HTTPServer) readyCheck(c *gin.Context) {
	// Check database connection
	if err := srv.postgresDB.PingContext(c.Request.Context()); err != nil {
		c.JSON(503, gin.H{
			"status":  "not ready",
			"message": "Database connection failed",
			"error":   err.Error(),
		})
		return
	}

	response.OK(c, gin.H{
		"status":   "ready",
		"message":  "From Tan Tai API V1 With Love",
		"version":  "1.0.0",
		"service":  "smap-api",
		"database": "connected",
	})
}

// liveCheck handles liveness check requests
// @Summary Liveness Check
// @Description Check if the API is alive
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "API is alive"
// @Router /live [get]
func (srv HTTPServer) liveCheck(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "alive",
		"message": "From Tan Tai API V1 With Love",
		"version": "1.0.0",
		"service": "smap-api",
	})
}
