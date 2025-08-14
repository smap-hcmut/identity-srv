package httpserver

import (
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/pkg/discord"
	pkgCrt "github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/minio"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
)

type HTTPServer struct {
	// Server Configuration
	gin  *gin.Engine
	l    pkgLog.Logger
	host string
	port int
	mode string

	// Database Configuration
	postgresDB *sql.DB

	// Storage Configuration
	minioClient minio.MinIO

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
	PostgresDB *sql.DB
	MongoDB    mongo.Client

	// Storage Configuration
	MinIOClient minio.MinIO

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
		postgresDB: cfg.PostgresDB,

		// Storage Configuration
		minioClient: cfg.MinIOClient,

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
		{s.postgresDB, "postgresDB is required"},

		// Storage Configuration
		{s.minioClient, "minioClient is required"},

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
