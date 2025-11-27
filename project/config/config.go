package config

import (
	"fmt"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	// Server Configuration
	HTTPServer HTTPServerConfig
	Logger     LoggerConfig

	// Database Configuration
	Postgres PostgresConfig

	// // Storage Configuration
	// MinIO MinIOConfig

	// // Message Queue Configuration
	// RabbitMQ RabbitMQConfig

	// Authentication & Security Configuration
	JWT            JWTConfig
	Cookie         CookieConfig
	Encrypter      EncrypterConfig
	InternalConfig InternalConfig

	// Monitoring & Notification Configuration
	Discord DiscordConfig
}

// JWTConfig is the configuration for the JWT,
// which is used to generate and verify the JWT.
type JWTConfig struct {
	SecretKey string `env:"JWT_SECRET"`
}

// CookieConfig is the configuration for HttpOnly cookies,
// which is used for secure authentication token storage.
type CookieConfig struct {
	Domain         string `env:"COOKIE_DOMAIN" envDefault:".smap.com"`
	Secure         bool   `env:"COOKIE_SECURE" envDefault:"true"`
	SameSite       string `env:"COOKIE_SAMESITE" envDefault:"Lax"`
	MaxAge         int    `env:"COOKIE_MAX_AGE" envDefault:"7200"`
	MaxAgeRemember int    `env:"COOKIE_MAX_AGE_REMEMBER" envDefault:"2592000"`
	Name           string `env:"COOKIE_NAME" envDefault:"smap_auth_token"`
}

// HTTPServerConfig is the configuration for the HTTP server,
// which is used to start, call API, etc.
type HTTPServerConfig struct {
	Host string `env:"HOST" envDefault:""`
	Port int    `env:"APP_PORT" envDefault:"8080"`
	Mode string `env:"API_MODE" envDefault:"debug"`
}

// RabbitMQConfig is the configuration for the RabbitMQ,
// which is used to connect to the RabbitMQ.
type RabbitMQConfig struct {
	URL string `env:"RABBITMQ_URL"`
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
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"prefer"`
}

type MinIOConfig struct {
	Endpoint  string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	AccessKey string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	SecretKey string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	UseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	Region    string `env:"MINIO_REGION" envDefault:"us-east-1"`
	Bucket    string `env:"MINIO_BUCKET"`

	// Async upload settings
	AsyncUploadWorkers   int `env:"MINIO_ASYNC_UPLOAD_WORKERS" envDefault:"4"`
	AsyncUploadQueueSize int `env:"MINIO_ASYNC_UPLOAD_QUEUE_SIZE" envDefault:"100"`
}

type DiscordConfig struct {
	WebhookID    string `env:"DISCORD_WEBHOOK_ID"`
	WebhookToken string `env:"DISCORD_WEBHOOK_TOKEN"`
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

// Load is the function to load the configuration from the environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Printf("Error loading configuration: %v", err)
		return nil, err
	}
	return cfg, nil
}
