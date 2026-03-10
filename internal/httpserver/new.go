package httpserver

import (
	"database/sql"
	"errors"

	"identity-srv/config"
	"identity-srv/internal/authentication/usecase"
	"identity-srv/pkg/discord"
	"identity-srv/pkg/encrypter"
	pkgJWT "identity-srv/pkg/jwt"
	"identity-srv/pkg/log"
	pkgRedis "identity-srv/pkg/redis"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HTTPServer struct {
	// Server Configuration
	gin         *gin.Engine
	l           log.Logger
	host        string
	port        int
	mode        string
	environment string

	// Database Configuration
	postgresDB *sql.DB

	// Storage Configuration (disabled for OAuth flow)
	// minio miniopkg.MinIO

	// Authentication & Security Configuration
	config            *config.Config
	jwtManager        *pkgJWT.Manager
	redisClient       *pkgRedis.Client
	sessionManager    *usecase.SessionManager
	blacklistManager  *usecase.BlacklistManager
	roleMapper        *usecase.RoleMapper
	redirectValidator *usecase.RedirectValidator
	cookieConfig      config.CookieConfig
	encrypter         encrypter.Encrypter

	// Monitoring & Notification Configuration
	discord *discord.Discord
}

type Config struct {
	// Server Configuration
	Logger      log.Logger
	Host        string
	Port        int
	Mode        string
	Environment string

	// Database Configuration
	PostgresDB *sql.DB

	// Storage Configuration (disabled for OAuth flow)
	// MinIO miniopkg.MinIO

	// Authentication & Security Configuration
	Config            *config.Config
	JWTManager        *pkgJWT.Manager
	RedisClient       *pkgRedis.Client
	RedirectValidator *usecase.RedirectValidator
	CookieConfig      config.CookieConfig
	Encrypter         encrypter.Encrypter

	// Monitoring & Notification Configuration
	Discord *discord.Discord
}

// New creates a new HTTPServer instance with the provided configuration.
func New(logger log.Logger, cfg Config) (*HTTPServer, error) {
	gin.SetMode(cfg.Mode)

	// Initialize session manager
	sessionTTL := time.Duration(cfg.Config.Session.TTL) * time.Second
	sessionManager := usecase.NewSessionManager(cfg.RedisClient, sessionTTL)

	// Initialize blacklist manager (using same Redis client as session)
	blacklistManager := usecase.NewBlacklistManager(cfg.RedisClient)

	// Initialize role mapper
	roleMapper := usecase.NewRoleMapper(cfg.Config)

	srv := &HTTPServer{
		// Server Configuration
		l:           logger,
		gin:         gin.New(),
		host:        cfg.Host,
		port:        cfg.Port,
		mode:        cfg.Mode,
		environment: cfg.Environment,

		// Database Configuration
		postgresDB: cfg.PostgresDB,

		// Storage Configuration (disabled for OAuth flow)
		// minio: cfg.MinIO,

		// Authentication & Security Configuration
		config:            cfg.Config,
		jwtManager:        cfg.JWTManager,
		redisClient:       cfg.RedisClient,
		sessionManager:    sessionManager,
		blacklistManager:  blacklistManager,
		roleMapper:        roleMapper,
		redirectValidator: cfg.RedirectValidator,
		cookieConfig:      cfg.CookieConfig,
		encrypter:         cfg.Encrypter,

		// Monitoring & Notification Configuration
		discord: cfg.Discord,
	}

	if err := srv.validate(); err != nil {
		return nil, err
	}

	// Add middlewares
	srv.gin.Use(srv.zapLoggerMiddleware())
	srv.gin.Use(gin.Recovery())

	return srv, nil
}

func (srv *HTTPServer) zapLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if path == "/health" || path == "/ready" || path == "/live" {
			return
		}

		if srv.environment == "production" {
			srv.l.Info(c.Request.Context(), "HTTP Request",
				zap.Int("status", status),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.Duration("latency", latency),
				zap.String("user-agent", c.Request.UserAgent()),
			)
		} else {
			// In development, you might still want standard gin logs or a simpler format
			srv.l.Infof(c.Request.Context(), "%s %s %d %s %s", c.Request.Method, path, status, latency, c.ClientIP())
		}
	}
}

// validate validates that all required dependencies are provided.
func (srv HTTPServer) validate() error {
	// Server Configuration
	if srv.l == nil {
		return errors.New("logger is required")
	}
	if srv.mode == "" {
		return errors.New("mode is required")
	}
	// host can be empty (listen on all interfaces)
	if srv.port == 0 {
		return errors.New("port is required")
	}

	// Database Configuration
	if srv.postgresDB == nil {
		return errors.New("postgresDB is required")
	}

	// Authentication & Security Configuration
	if srv.config == nil {
		return errors.New("config is required")
	}
	if srv.jwtManager == nil {
		return errors.New("jwtManager is required")
	}
	if srv.redisClient == nil {
		return errors.New("redisClient is required")
	}
	if srv.sessionManager == nil {
		return errors.New("sessionManager is required")
	}
	if srv.blacklistManager == nil {
		return errors.New("blacklistManager is required")
	}
	if srv.encrypter == nil {
		return errors.New("encrypter is required")
	}

	// Monitoring & Notification Configuration (optional)
	// if srv.discord == nil {
	// 	return errors.New("discord is required")
	// }

	return nil
}
