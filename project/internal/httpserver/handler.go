package httpserver

import (
	"context"
	"smap-project/internal/keyword/usecase"
	"smap-project/internal/middleware"
	projecthttp "smap-project/internal/project/delivery/http"
	projectProd "smap-project/internal/project/delivery/rabbitmq/producer"
	projectrepository "smap-project/internal/project/repository/postgre"
	projectusecase "smap-project/internal/project/usecase"
	webhookhttp "smap-project/internal/webhook/delivery/http"
	webhookusecase "smap-project/internal/webhook/usecase"
	"smap-project/pkg/collector"
	"smap-project/pkg/i18n"
	"smap-project/pkg/llm"
	"smap-project/pkg/scope"

	// Import this to execute the init function in docs.go which setups the Swagger docs.
	// Uncomment after running: make swagger
	_ "smap-project/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (srv HTTPServer) mapHandlers() error {
	scopeManager := scope.New(srv.jwtSecretKey)
	mw := middleware.New(srv.l, scopeManager, srv.cookieConfig, srv.internalKey)

	srv.registerMiddlewares(mw)
	srv.registerSystemRoutes()

	i18n.Init()

	// Initialize LLM provider
	llmProvider, err := llm.NewProvider(srv.llmConfig, srv.l)
	if err != nil {
		return err
	}

	// Initialize Collector client
	collectorClient := collector.NewClient(srv.collectorConfig, srv.l)

	// Initialize project repository
	projectRepo := projectrepository.New(srv.postgresDB, srv.l)

	// Initialize keyword usecase
	keywordUC := usecase.New(srv.l, llmProvider, collectorClient)

	projectProd := projectProd.New(srv.l, *srv.amqpConn)
	if err := projectProd.Run(); err != nil {
		return err
	}

	// Initialize webhook usecase first (needed by project usecase)
	webhookUC := webhookusecase.New(srv.l, srv.redisClient)
	webhookHandler := webhookhttp.New(srv.l, webhookUC, srv.discord, srv.internalKey)
	webhookhttp.MapWebhookRoutes(srv.gin.Group("/internal"), webhookHandler, mw)

	// Initialize project usecase (pass webhookUC)
	projectUC := projectusecase.New(srv.l, projectRepo, keywordUC, projectProd, webhookUC)
	// Initialize project HTTP handler
	projectHandler := projecthttp.New(srv.l, projectUC, srv.discord)
	// Map routes (no prefix)
	projecthttp.MapProjectRoutes(srv.gin.Group("/projects"), projectHandler, mw)

	return nil
}

func (srv HTTPServer) registerMiddlewares(mw middleware.Middleware) {
	srv.gin.Use(middleware.Recovery(srv.l, srv.discord))

	corsConfig := middleware.DefaultCORSConfig(srv.environment)
	srv.gin.Use(middleware.CORS(corsConfig))

	// Log CORS mode for visibility
	ctx := context.Background()
	if srv.environment == "production" {
		srv.l.Infof(ctx, "CORS mode: production (strict origins only)")
	} else {
		srv.l.Infof(ctx, "CORS mode: %s (permissive - allows localhost and private subnets)", srv.environment)
	}

	// Add locale middleware to extract and set locale from request header
	srv.gin.Use(mw.Locale())
}

func (srv HTTPServer) registerSystemRoutes() {
	srv.gin.GET("/health", srv.healthCheck)
	srv.gin.GET("/ready", srv.readyCheck)
	srv.gin.GET("/live", srv.liveCheck)

	// Swagger UI and docs
	srv.gin.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("doc.json"), // Use relative path
		ginSwagger.DefaultModelsExpandDepth(-1),
	))
}
