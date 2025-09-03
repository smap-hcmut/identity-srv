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

	// Auth
	authHTTP "github.com/nguyentantai21042004/smap-api/internal/auth/delivery/http"
	authProducer "github.com/nguyentantai21042004/smap-api/internal/auth/delivery/rabbitmq/producer"
	authUC "github.com/nguyentantai21042004/smap-api/internal/auth/usecase"

	// Core SMTP
	smtpUC "github.com/nguyentantai21042004/smap-api/internal/core/smtp/usecase"

	// Role
	// roleDelivery "github.com/nguyentantai21042004/smap-api/internal/role/delivery/http"
	roleRepo "github.com/nguyentantai21042004/smap-api/internal/role/repository/mongo"
	roleUC "github.com/nguyentantai21042004/smap-api/internal/role/usecase"

	// User
	userRepo "github.com/nguyentantai21042004/smap-api/internal/user/repository/mongo"
	userUC "github.com/nguyentantai21042004/smap-api/internal/user/usecase"
	userHTTP "github.com/nguyentantai21042004/smap-api/internal/user/delivery/http"
	
	// Session
	sessionRepo "github.com/nguyentantai21042004/smap-api/internal/session/repository/mongo"
	sessionUC "github.com/nguyentantai21042004/smap-api/internal/session/usecase"
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

	// Role
	roleRepo := roleRepo.New(srv.l, srv.mongoDB)
	roleUC := roleUC.New(srv.l, roleRepo)
	// roleHandler := roleDelivery.New(srv.l, roleUC, discord)

	// User
	userRepo := userRepo.New(srv.l, srv.mongoDB)
	userUC := userUC.New(srv.l, userRepo, roleUC)
	userH := userHTTP.New(srv.l, userUC, discord)

	// Session
	sessionRepo := sessionRepo.New(srv.l, srv.mongoDB)
	sessionUC := sessionUC.New(srv.l, sessionRepo, userUC)

	// SMTP Core
	smtpUC := smtpUC.New(srv.l, srv.smtpConfig)

	// Auth Producer
	authProd := authProducer.NewProducer(srv.l, srv.amqpConn)
	if err := authProd.Run(); err != nil {
		return err
	}
	authUC := authUC.New(
		srv.l,
		authProd,
		srv.encrypter,
		srv.oauthConfig,
		scopeUC,
		smtpUC,
		userUC,
		roleUC,
		sessionUC,
	)
	authH := authHTTP.New(srv.l, authUC, discord)

	// Middleware
	mw := middleware.New(srv.l, scopeUC)

	// Apply locale middleware
	srv.gin.Use(mw.Locale())
	api := srv.gin.Group(Api)

	// Map Routes
	// roleDelivery.MapRoutes(api.Group("/roles"), roleHandler, mw)
	authHTTP.MapAuthRoutes(api.Group("/auth"), authH, mw)
	userHTTP.MapUserRoutes(api.Group("/user"), userH, mw)

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
	if err := srv.mongoDB.Client().Ping(c.Request.Context()); err != nil {
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
