package httpserver

import (
	"database/sql"
	"errors"

	"smap-api/config"
	"smap-api/pkg/discord"
	"smap-api/pkg/encrypter"
	"smap-api/pkg/log"
	miniopkg "smap-api/pkg/minio"
	pkgRabbitMQ "smap-api/pkg/rabbitmq"

	"github.com/gin-gonic/gin"
)

const (
	productionMode = "production"
	debugMode      = "debug"
)

var (
	ginDebugMode   = gin.DebugMode
	ginReleaseMode = gin.ReleaseMode
	ginTestMode    = gin.TestMode
)

type HTTPServer struct {
	// Server Configuration
	gin  *gin.Engine
	l    log.Logger
	host string
	port int
	mode string

	// Database Configuration
	postgresDB *sql.DB

	// Storage Configuration
	minio miniopkg.MinIO

	// SMTP Configuration
	smtpConfig config.SMTPConfig

	// Message Queue Configuration
	amqpConn *pkgRabbitMQ.Connection

	// Authentication & Security Configuration
	jwtSecretKey string
	encrypter    encrypter.Encrypter
	internalKey  string

	// Monitoring & Notification Configuration
	discord *discord.Discord
}

type Config struct {
	// Server Configuration
	Logger log.Logger
	Host   string
	Port   int
	Mode   string

	// Database Configuration
	PostgresDB *sql.DB

	// Storage Configuration
	MinIO miniopkg.MinIO

	// SMTP Configuration
	SMTP config.SMTPConfig

	// Message Queue Configuration
	AmqpConn *pkgRabbitMQ.Connection

	// Authentication & Security Configuration
	JwtSecretKey string
	Encrypter    encrypter.Encrypter
	InternalKey  string

	// Monitoring & Notification Configuration
	Discord *discord.Discord
}

// New creates a new HTTPServer instance with the provided configuration.
func New(logger log.Logger, cfg Config) (*HTTPServer, error) {
	gin.SetMode(cfg.Mode)

	srv := &HTTPServer{
		// Server Configuration
		l:    logger,
		gin:  gin.Default(),
		host: cfg.Host,
		port: cfg.Port,
		mode: cfg.Mode,

		// Database Configuration
		postgresDB: cfg.PostgresDB,

		// Storage Configuration
		minio: cfg.MinIO,

		// SMTP Configuration
		smtpConfig: cfg.SMTP,

		// Message Queue Configuration
		amqpConn: cfg.AmqpConn,

		// Authentication & Security Configuration
		jwtSecretKey: cfg.JwtSecretKey,
		encrypter:    cfg.Encrypter,
		internalKey:  cfg.InternalKey,

		// Monitoring & Notification Configuration
		discord: cfg.Discord,
	}

	if err := srv.validate(); err != nil {
		return nil, err
	}

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
	if srv.host == "" {
		return errors.New("host is required")
	}
	if srv.port == 0 {
		return errors.New("port is required")
	}

	// Database Configuration
	if srv.postgresDB == nil {
		return errors.New("postgresDB is required")
	}

	// Storage Configuration
	if srv.minio == nil {
		return errors.New("minio is required")
	}

	// SMTP Configuration (optional for API - only consumer needs it)
	// API doesn't send emails directly, it uses RabbitMQ to queue email tasks
	// Commented out to make SMTP optional for API service
	// if srv.smtpConfig.Host == "" {
	// 	return errors.New("smtp host is required")
	// }
	// if srv.smtpConfig.Username == "" {
	// 	return errors.New("smtp username is required")
	// }
	// if srv.smtpConfig.Password == "" {
	// 	return errors.New("smtp password is required")
	// }
	// if srv.smtpConfig.From == "" {
	// 	return errors.New("smtp from address is required")
	// }

	// Message Queue Configuration
	if srv.amqpConn == nil {
		return errors.New("amqp connection is required")
	}

	// Authentication & Security Configuration
	if srv.jwtSecretKey == "" {
		return errors.New("jwtSecretKey is required")
	}
	if srv.encrypter == nil {
		return errors.New("encrypter is required")
	}
	if srv.internalKey == "" {
		return errors.New("internalKey is required")
	}

	// Monitoring & Notification Configuration
	if srv.discord == nil {
		return errors.New("discord is required")
	}

	return nil
}
