package httpserver

import (
	"context"
	"fmt"
	audithttp "identity-srv/internal/audit/delivery/http"
	auditPostgre "identity-srv/internal/audit/repository/postgre"
	authhttp "identity-srv/internal/authentication/delivery/http"
	authusecase "identity-srv/internal/authentication/usecase"
	"identity-srv/internal/model"
	userrepository "identity-srv/internal/user/repository/postgre"
	userusecase "identity-srv/internal/user/usecase"
	"identity-srv/pkg/oauth"

	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (srv HTTPServer) mapHandlers() error {
	// Initialize middleware
	mw := middleware.New(middleware.Config{
		JWTManager:       srv.jwtManager,
		CookieName:       srv.cookieConfig.Name,
		ProductionDomain: srv.cookieConfig.Domain,
		InternalKey:      srv.config.InternalConfig.InternalKey,
		IsProduction:     srv.environment == string(model.EnvironmentProduction),
	})

	srv.registerMiddlewares()
	srv.registerSystemRoutes()

	// Initialize repositories
	userRepo := userrepository.New(srv.l, srv.postgresDB)

	// Initialize usecases
	userUC := userusecase.New(srv.l, srv.encrypter, userRepo)

	// Initialize authentication usecase - use scope manager from shared-libs
	scopeManager := auth.NewManager(srv.config.JWT.SecretKey)
	authUC := authusecase.New(srv.l, scopeManager, srv.encrypter, userUC)
	authUC.SetSessionManager(srv.sessionManager)
	authUC.SetBlacklistManager(srv.blacklistManager)
	authUC.SetJWTManager(srv.jwtManager)
	authUC.SetRoleMapper(srv.roleMapper)

	// Initialize OAuth provider
	oauthProvider, err := srv.initOAuthProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize OAuth provider: %w", err)
	}
	authUC.SetOAuthProvider(oauthProvider)

	authUC.SetRedirectValidator(srv.redirectValidator)

	// Initialize HTTP handlers with new dependencies
	authHandler := authhttp.New(srv.l, authUC, srv.discord, srv.config)

	// userHandler := userhttp.New(srv.l, userUC, srv.discord)

	// Initialize audit handler
	auditRepo := auditPostgre.New(srv.l, srv.postgresDB)
	auditHandler := audithttp.New(srv.l, auditRepo, srv.discord)

	// Map routes with middleware
	apiV1 := srv.gin.Group(model.APIV1Prefix)
	authHandler.RegisterRoutes(apiV1.Group("/authentication"), mw)
	auditHandler.RegisterRoutes(apiV1.Group("/audit-logs"), mw)

	return nil
}

func (srv HTTPServer) registerMiddlewares() {
	// Recovery middleware with Discord reporting
	srv.gin.Use(middleware.Recovery(srv.l, srv.discord))

	corsConfig := middleware.DefaultCORSConfig(srv.environment)
	srv.gin.Use(middleware.CORS(corsConfig))

	// Tracing middleware for centralized logging (trace_id)
	srv.gin.Use(middleware.Tracing())

	// Log CORS mode for visibility
	ctx := context.Background()
	if srv.environment == string(model.EnvironmentProduction) {
		srv.l.Infof(ctx, "CORS mode: production (strict origins only)")
	} else {
		srv.l.Infof(ctx, "CORS mode: %s (permissive - allows localhost and private subnets)", srv.environment)
	}

	// Add locale middleware to extract and set locale from request header
	srv.gin.Use(middleware.Locale())
}

func (srv HTTPServer) registerSystemRoutes() {
	srv.gin.GET("/health", srv.healthCheck)
	srv.gin.GET("/ready", srv.readyCheck)
	srv.gin.GET("/live", srv.liveCheck)

	// Test client (development only)
	if srv.environment != string(model.EnvironmentProduction) {
		srv.gin.StaticFile("/test", "./cmd/test-client/index.html")
	}

	// Swagger UI and docs
	srv.gin.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("doc.json"), // Use relative path
		ginSwagger.DefaultModelsExpandDepth(-1),
	))
}

// initOAuthProvider initializes the OAuth provider based on configuration
func (srv HTTPServer) initOAuthProvider() (oauth.Provider, error) {
	oauthCfg := oauth.Config{
		ClientID:     srv.config.OAuth2.ClientID,
		ClientSecret: srv.config.OAuth2.ClientSecret,
		RedirectURI:  srv.config.OAuth2.RedirectURI,
		Scopes:       srv.config.OAuth2.Scopes,
		ProviderType: srv.config.OAuth2.Provider,
		OktaDomain:   srv.config.OAuth2.OktaDomain,
	}

	provider, err := oauth.NewProvider(oauthCfg)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	srv.l.Infof(ctx, "OAuth provider initialized: %s", provider.GetProviderName())

	return provider, nil
}
