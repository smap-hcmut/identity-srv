package config

import (
	"fmt"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	// Logger Configuration
	Logger LoggerConfig

	// Message Queue Configuration
	RabbitMQConfig RabbitMQConfig

	// External Services
	Project ProjectConfig

	// Monitoring & Notification Configuration
	Discord DiscordConfig
}

// LoggerConfig is the configuration for the logger.
type LoggerConfig struct {
	Level    string `env:"LOGGER_LEVEL" envDefault:"debug"`
	Mode     string `env:"LOGGER_MODE" envDefault:"debug"`
	Encoding string `env:"LOGGER_ENCODING" envDefault:"console"`
}

// DiscordConfig is the configuration for Discord webhooks.
type DiscordConfig struct {
	ReportBugID    string `env:"DISCORD_REPORT_BUG_ID"`
	ReportBugToken string `env:"DISCORD_REPORT_BUG_TOKEN"`
}

// RabbitMQConfig is the configuration for RabbitMQ,
// which is used to connect to RabbitMQ server.
type RabbitMQConfig struct {
	URL string `env:"RABBITMQ_URL"`
}

// ProjectConfig is the configuration for the Project Service.
type ProjectConfig struct {
	BaseURL              string `env:"PROJECT_SERVICE_URL" envDefault:"http://localhost:8080"`
	Timeout              int    `env:"PROJECT_TIMEOUT" envDefault:"10"`
	InternalKey          string `env:"PROJECT_INTERNAL_KEY"`
	WebhookRetryAttempts int    `env:"WEBHOOK_RETRY_ATTEMPTS" envDefault:"5"`
	WebhookRetryDelay    int    `env:"WEBHOOK_RETRY_DELAY" envDefault:"1"`
}

// Load is the function to load the configuration from the environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	// Print all config for testing
	fmt.Printf("%+v\n", cfg)
	return cfg, nil
}
