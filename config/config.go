package config

import (
	"github.com/caarlos0/env/v9"
)

type Config struct {
	// Server Configuration
	HTTPServer HTTPServerConfig
	Logger     LoggerConfig

	// Database Configuration
	Postgres PostgresConfig

	// Storage Configuration
	MinIO MinIOConfig

	// Authentication & Security Configuration
	JWT            JWTConfig
	Encrypter      EncrypterConfig
	InternalConfig InternalConfig

	// WebSocket Configuration
	WebSocket WebSocketConfig

	// Monitoring & Notification Configuration
	Discord DiscordConfig
}

// JWTConfig is the configuration for the JWT,
// which is used to generate and verify the JWT.
type JWTConfig struct {
	SecretKey string `env:"JWT_SECRET"`
}

// HTTPServerConfig is the configuration for the HTTP server,
// which is used to start, call API, etc.
type HTTPServerConfig struct {
	Host string `env:"HOST" envDefault:""`
	Port int    `env:"APP_PORT" envDefault:"8080"`
	Mode string `env:"API_MODE" envDefault:"debug"`
}

// LoggerConfig is the configuration for the logger,
// which is used to log the application.
type LoggerConfig struct {
	Level    string `env:"LOGGER_LEVEL" envDefault:"debug"`
	Mode     string `env:"LOGGER_MODE" envDefault:"debug"`
	Encoding string `env:"LOGGER_ENCODING" envDefault:"console"`
}

// PostgresConfig is the configuration for the Postgres,
// which is used to connect to the Postgres.
type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	DBName   string `env:"POSTGRES_DB" envDefault:"postgres"`
}

type MinIOConfig struct {
	Endpoint  string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	AccessKey string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	SecretKey string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	UseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	Region    string `env:"MINIO_REGION" envDefault:"us-east-1"`
	Bucket    string `env:"MINIO_BUCKET"`
}

type DiscordConfig struct {
	ReportBugID    string `env:"DISCORD_REPORT_BUG_ID"`
	ReportBugToken string `env:"DISCORD_REPORT_BUG_TOKEN"`
}

// EncrypterConfig is the configuration for the encrypter,
// which is used to encrypt and decrypt the data.
type EncrypterConfig struct {
	Key string `env:"ENCRYPT_KEY"`
}

// InternalConfig is the configuration for the internal,
// which is used to check the internal request.
type InternalConfig struct {
	InternalKey string `env:"INTERNAL_KEY"`
}

// WebSocketConfig is the configuration for the WebSocket,
// which is used to configure WebSocket settings.
type WebSocketConfig struct {
	ReadBufferSize  int `env:"WS_READ_BUFFER_SIZE" envDefault:"1024"`
	WriteBufferSize int `env:"WS_WRITE_BUFFER_SIZE" envDefault:"1024"`
	MaxMessageSize  int `env:"WS_MAX_MESSAGE_SIZE" envDefault:"512"`
	PongWait        int `env:"WS_PONG_WAIT" envDefault:"60"`
	PingPeriod      int `env:"WS_PING_PERIOD" envDefault:"54"`
	WriteWait       int `env:"WS_WRITE_WAIT" envDefault:"10"`
}

// Load is the function to load the configuration from the environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
