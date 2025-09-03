package httpserver

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	pkgCrt "github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"github.com/nguyentantai21042004/smap-api/pkg/redis"
	pkgRabbitMQ "github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
	"github.com/nguyentantai21042004/smap-api/internal/appconfig/oauth"
)

type HTTPServer struct {
	// Server Configuration
	gin  *gin.Engine
	l    pkgLog.Logger
	host string
	port int
	mode string

	// Database Configuration
	mongoDB mongo.Database

	// Cache Configuration
	redisClient *redis.Client

	// Message Queue Configuration
	amqpConn *pkgRabbitMQ.Connection

	// Authentication & Security Configuration
	jwtSecretKey string
	encrypter    pkgCrt.Encrypter
	internalKey  string
	oauthConfig  oauth.OauthConfig

	// External Services Configuration
	smtpConfig config.SMTPConfig

	// WebSocket Configuration
	wsConfig config.WebSocketConfig

	// Monitoring & Notification Configuration
	discord *discord.DiscordWebhook
}

type Config struct {
	// Server Configuration
	Logger pkgLog.Logger
	Host   string
	Port   int
	Mode   string

	// Database Configuration
	MongoDB mongo.Database

	// Cache Configuration
	RedisClient *redis.Client

	// Message Queue Configuration
	AMQPConn *pkgRabbitMQ.Connection

	// Authentication & Security Configuration
	JwtSecretKey string
	Encrypter    pkgCrt.Encrypter
	InternalKey  string
	OauthConfig  oauth.OauthConfig

	// External Services Configuration
	SMTPConfig config.SMTPConfig

	// WebSocket Configuration
	WebSocketConfig config.WebSocketConfig

	// Monitoring & Notification Configuration
	DiscordConfig *discord.DiscordWebhook
}

func New(l pkgLog.Logger, cfg Config) (*HTTPServer, error) {
	if cfg.Mode == productionMode {
		ginMode = gin.ReleaseMode
	}

	gin.SetMode(ginMode)

	h := &HTTPServer{
		// Server Configuration
		l:    l,
		gin:  gin.Default(),
		host: cfg.Host,
		port: cfg.Port,
		mode: cfg.Mode,

		// Database Configuration
		mongoDB: cfg.MongoDB,

		// Cache Configuration
		redisClient: cfg.RedisClient,

		// Message Queue Configuration
		amqpConn: cfg.AMQPConn,

		// Authentication & Security Configuration
		jwtSecretKey: cfg.JwtSecretKey,
		encrypter:    cfg.Encrypter,
		internalKey:  cfg.InternalKey,

		// WebSocket Configuration
		wsConfig: cfg.WebSocketConfig,

		// Monitoring & Notification Configuration
		discord: cfg.DiscordConfig,
	}

	if err := h.validate(); err != nil {
		return nil, err
	}

	return h, nil
}

func (s HTTPServer) validate() error {
	requiredDeps := []struct {
		dep interface{}
		msg string
	}{
		// Server Configuration
		{s.l, "logger is required"},
		{s.mode, "mode is required"},
		{s.host, "host is required"},
		{s.port, "port is required"},

		// Database Configuration
		{s.mongoDB, "mongoDB is required"},

		// Cache Configuration
		{s.redisClient, "redisClient is required"},

		// Message Queue Configuration
		{s.amqpConn, "amqpConn is required"},


		// Authentication & Security Configuration
		{s.jwtSecretKey, "jwtSecretKey is required"},
		{s.encrypter, "encrypter is required"},
		{s.internalKey, "internalKey is required"},
		{s.oauthConfig, "oauthConfig is required"},

		// External Services Configuration
		{s.smtpConfig, "smtpConfig is required"},
		
		// WebSocket Configuration
		{s.wsConfig, "wsConfig is required"},

		// Monitoring & Notification Configuration
		{s.discord, "discord is required"},
	}

	for _, dep := range requiredDeps {
		if dep.dep == nil {
			return errors.New(dep.msg)
		}
	}

	return nil
}
