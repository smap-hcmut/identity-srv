package httpserver

import (
	"errors"

	"smap-collector/config"
	"smap-collector/pkg/discord"
	pkgCrt "smap-collector/pkg/encrypter"
	pkgLog "smap-collector/pkg/log"
	"smap-collector/pkg/mongo"

	"github.com/gin-gonic/gin"
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

	// Authentication & Security Configuration
	jwtSecretKey string
	encrypter    pkgCrt.Encrypter
	internalKey  string

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

	// Authentication & Security Configuration
	JwtSecretKey string
	Encrypter    pkgCrt.Encrypter
	InternalKey  string

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

		// Authentication & Security Configuration
		{s.jwtSecretKey, "jwtSecretKey is required"},
		{s.encrypter, "encrypter is required"},
		{s.internalKey, "internalKey is required"},

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
