package httpserver

import (
	authhttp "smap-api/internal/authentication/delivery/http"
	authproducer "smap-api/internal/authentication/delivery/rabbitmq/producer"
	authusecase "smap-api/internal/authentication/usecase"
	"smap-api/internal/middleware"
	userrepository "smap-api/internal/user/repository/postgre"
	userusecase "smap-api/internal/user/usecase"
	"smap-api/pkg/i18n"
	"smap-api/pkg/scope"

	// Import this to execute the init function in docs.go which setups the Swagger docs.
	_ "smap-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const apiPrefix = "/api/v1"

func (srv HTTPServer) mapHandlers() error {
	srv.registerMiddlewares()
	srv.registerSystemRoutes()

	scopeManager := scope.New(srv.jwtSecretKey)
	mw := middleware.New(srv.l, scopeManager)

	i18n.Init()

	userRepo := userrepository.New(srv.l, srv.postgresDB)
	userUC := userusecase.New(srv.l, userRepo)

	authProd := authproducer.New(srv.l, srv.amqpConn)
	if err := authProd.Run(); err != nil {
		return err
	}

	authUC := authusecase.New(srv.l, authProd, scopeManager, srv.encrypter, userUC)
	authHandler := authhttp.New(srv.l, authUC, srv.discord)

	api := srv.gin.Group(apiPrefix)
	authhttp.MapAuthRoutes(api.Group("/auth"), authHandler, mw)

	return nil
}

func (srv HTTPServer) registerMiddlewares() {
	srv.gin.Use(middleware.Recovery(srv.l, srv.discord))

	corsConfig := middleware.DefaultCORSConfig()
	srv.gin.Use(middleware.CORS(corsConfig))
}

func (srv HTTPServer) registerSystemRoutes() {
	srv.gin.GET("/health", srv.healthCheck)
	srv.gin.GET("/ready", srv.readyCheck)
	srv.gin.GET("/live", srv.liveCheck)
	srv.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
