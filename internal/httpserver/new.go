package httpserver

import (
	"database/sql"
	"errors"
	"time"

	"identity-srv/config"
	"identity-srv/internal/authentication/usecase"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/auth"
	"github.com/smap-hcmut/shared-libs/go/middleware"
	"github.com/smap-hcmut/shared-libs/go/discord"
	"github.com/smap-hcmut/shared-libs/go/encrypter"
	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/redis"
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
	jwtManager        auth.Manager
	redisClient       redis.IRedis
	sessionManager    *usecase.SessionManager
	blacklistManager  *usecase.BlacklistManager
	roleMapper        *usecase.RoleMapper
	redirectValidator *usecase.RedirectValidator
	cookieConfig      config.CookieConfig
	encrypter         encrypter.Encrypter

	// Monitoring & Notification Configuration
	discord discord.IDiscord
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
	JWTManager        auth.Manager
	RedisClient       redis.IRedis
	RedirectValidator *usecase.RedirectValidator
	CookieConfig      config.CookieConfig
	Encrypter         encrypter.Encrypter

	// Monitoring & Notification Configuration
	Discord discord.IDiscord
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
	srv.gin.Use(middleware.Logger(srv.l, srv.environment))
	srv.gin.Use(gin.Recovery())

	return srv, nil
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
