package httpserver

import (
	authhttp "smap-api/internal/authentication/delivery/http"
	authproducer "smap-api/internal/authentication/delivery/rabbitmq/producer"
	authusecase "smap-api/internal/authentication/usecase"
	"smap-api/internal/middleware"
	planhttp "smap-api/internal/plan/delivery/http"
	planrepository "smap-api/internal/plan/repository/postgre"
	planusecase "smap-api/internal/plan/usecase"
	subscriptionhttp "smap-api/internal/subscription/delivery/http"
	subscriptionrepository "smap-api/internal/subscription/repository/postgre"
	subscriptionusecase "smap-api/internal/subscription/usecase"
	userhttp "smap-api/internal/user/delivery/http"
	userrepository "smap-api/internal/user/repository/postgre"
	userusecase "smap-api/internal/user/usecase"
	"smap-api/pkg/i18n"
	"smap-api/pkg/scope"

	// Import this to execute the init function in docs.go which setups the Swagger docs.
	_ "smap-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (srv HTTPServer) mapHandlers() error {
	scopeManager := scope.New(srv.jwtSecretKey)
	mw := middleware.New(srv.l, scopeManager, srv.cookieConfig)

	srv.registerMiddlewares(mw)
	srv.registerSystemRoutes()

	i18n.Init()

	// Initialize repositories
	userRepo := userrepository.New(srv.l, srv.postgresDB)
	planRepo := planrepository.New(srv.l, srv.postgresDB)
	subscriptionRepo := subscriptionrepository.New(srv.l, srv.postgresDB)

	// Initialize usecases
	userUC := userusecase.New(srv.l, srv.encrypter, userRepo)
	planUC := planusecase.New(srv.l, planRepo)
	subscriptionUC := subscriptionusecase.New(srv.l, subscriptionRepo, planUC)

	// Initialize authentication producer
	authProd := authproducer.New(srv.l, srv.amqpConn)
	if err := authProd.Run(); err != nil {
		return err
	}

	// Initialize authentication usecase with plan and subscription dependencies
	authUC := authusecase.New(srv.l, authProd, scopeManager, srv.encrypter, userUC, planUC, subscriptionUC)

	// Initialize HTTP handlers
	authHandler := authhttp.New(srv.l, authUC, srv.discord, srv.cookieConfig)
	planHandler := planhttp.New(srv.l, planUC, srv.discord)
	subscriptionHandler := subscriptionhttp.New(srv.l, subscriptionUC, srv.discord)
	userHandler := userhttp.New(srv.l, userUC, srv.discord)

	// Map routes (no prefix)
	authhttp.MapAuthRoutes(srv.gin.Group("/authentication"), authHandler, mw)
	planhttp.MapPlanRoutes(srv.gin.Group("/plans"), planHandler, mw)
	subscriptionhttp.MapSubscriptionRoutes(srv.gin.Group("/subscriptions"), subscriptionHandler, mw)
	userhttp.MapUserRoutes(srv.gin.Group("/users"), userHandler, mw)

	return nil
}

func (srv HTTPServer) registerMiddlewares(mw middleware.Middleware) {
	srv.gin.Use(middleware.Recovery(srv.l, srv.discord))

	corsConfig := middleware.DefaultCORSConfig()
	srv.gin.Use(middleware.CORS(corsConfig))

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
