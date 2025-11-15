package httpserver

import (

	// Import this to execute the init function in docs.go which setups the Swagger docs.
	_ "smap-api/docs" // TODO: Generate docs package

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	Api         = "/api/v1"
	InternalApi = "internal/api/v1"
)

func (srv HTTPServer) mapHandlers() error {
	// discord, err := discord.New(srv.l, srv.discord)
	// if err != nil {
	// 	return err
	// }
	// // srv.gin.Use(middleware.Recovery(discord))

	// Health check endpoint
	srv.gin.GET("/health", srv.healthCheck)
	srv.gin.GET("/ready", srv.readyCheck)
	srv.gin.GET("/live", srv.liveCheck)

	// Swagger UI
	srv.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// routes
	// api := srv.gin.Group(Api)

	return nil
}
