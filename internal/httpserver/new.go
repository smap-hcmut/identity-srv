package httpserver

import (
	"database/sql"
	"errors"

	"smap-api/pkg/discord"
	"smap-api/pkg/encrypter"
	"smap-api/pkg/log"

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

	// Authentication & Security Configuration
	JwtSecretKey string
	Encrypter    encrypter.Encrypter
	InternalKey  string

	// Monitoring & Notification Configuration
	Discord *discord.Discord
}

func New(l log.Logger, cfg Config) (*HTTPServer, error) {
	gin.SetMode(cfg.Mode)

	h := &HTTPServer{
		// Server Configuration
		l:    l,
		gin:  gin.Default(),
		host: cfg.Host,
		port: cfg.Port,
		mode: cfg.Mode,

		// Database Configuration
		postgresDB: cfg.PostgresDB,

		// Authentication & Security Configuration
		jwtSecretKey: cfg.JwtSecretKey,
		encrypter:    cfg.Encrypter,
		internalKey:  cfg.InternalKey,

		// Monitoring & Notification Configuration
		discord: cfg.Discord,
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
