package httpserver

import (
	"context"
	"database/sql"

	audithttp "smap-api/internal/audit/delivery/http"
	auditPostgre "smap-api/internal/audit/repository/postgre"
	authhttp "smap-api/internal/authentication/delivery/http"
	authusecase "smap-api/internal/authentication/usecase"
	"smap-api/internal/middleware"
	userrepository "smap-api/internal/user/repository/postgre"
	userusecase "smap-api/internal/user/usecase"
	"smap-api/pkg/i18n"
	"smap-api/pkg/scope"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (srv HTTPServer) mapHandlers() error {
	scopeManager := scope.New("temporary-secret-not-used-in-oauth") // Legacy scope manager - not used in OAuth flow
	mw := middleware.New(srv.l, scopeManager, srv.cookieConfig, "", srv.config, srv.encrypter)

	srv.registerMiddlewares(mw)
	srv.registerSystemRoutes()

	i18n.Init()

	// Initialize repositories
	userRepo := userrepository.New(srv.l, srv.postgresDB)

	// Initialize usecases
	userUC := userusecase.New(srv.l, srv.encrypter, userRepo)

	// Initialize authentication usecase
	authUC := authusecase.New(srv.l, scopeManager, srv.encrypter, userUC)

	// Set database connection for direct user operations (bypassing legacy user usecase)
	if authUCWithDB, ok := authUC.(interface{ SetDB(*sql.DB) }); ok {
		authUCWithDB.SetDB(srv.postgresDB)
	}

	// Initialize HTTP handlers with new dependencies
	authHandler := authhttp.New(srv.l, authUC, srv.discord, srv.config, srv.jwtManager, srv.sessionManager, srv.blacklistManager, srv.googleClient, srv.groupsManager, srv.roleMapper, userRepo)

	// Initialize OAuth2 config
	authHandler.InitOAuth2Config(authhttp.OAuthConfig{
		ClientID:     srv.config.OAuth2.ClientID,
		ClientSecret: srv.config.OAuth2.ClientSecret,
		RedirectURI:  srv.config.OAuth2.RedirectURI,
		Scopes:       srv.config.OAuth2.Scopes,
	})

	// userHandler := userhttp.New(srv.l, userUC, srv.discord)

	// Initialize audit handler
	auditRepo := auditPostgre.New(srv.postgresDB)
	auditHandler := audithttp.New(srv.l, auditRepo, srv.discord)

	// Map routes (no prefix)
	authhttp.MapAuthRoutes(srv.gin.Group("/authentication"), authHandler, mw)
	audithttp.MapAuditRoutes(srv.gin.Group("/audit-logs"), auditHandler, mw)
	// userhttp.MapUserRoutes(srv.gin.Group("/users"), userHandler, mw) // Temporarily disabled for Task 1.9

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
