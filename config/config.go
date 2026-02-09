package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// Environment Configuration
	Environment EnvironmentConfig

	// Server Configuration
	HTTPServer HTTPServerConfig
	Logger     LoggerConfig

	// Database Configuration
	Postgres PostgresConfig

	// Message Queue Configuration (Kafka replaces RabbitMQ)
	Kafka KafkaConfig

	// Cache Configuration
	Redis RedisConfig

	// Authentication & Security Configuration
	OAuth2         OAuth2Config
	JWT            JWTConfig
	Cookie         CookieConfig
	Encrypter      EncrypterConfig
	InternalConfig InternalConfig

	// Google Workspace Integration
	GoogleWorkspace GoogleWorkspaceConfig

	// Access Control
	AccessControl AccessControlConfig

	// Session Configuration
	Session SessionConfig

	// Token Blacklist
	Blacklist BlacklistConfig

	// Monitoring & Notification Configuration
	Discord DiscordConfig
}

// EnvironmentConfig is the configuration for the deployment environment.
type EnvironmentConfig struct {
	Name string
}

// OAuth2Config is the configuration for OAuth2 authentication
type OAuth2Config struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

// GoogleWorkspaceConfig is the configuration for Google Workspace integration
type GoogleWorkspaceConfig struct {
	ServiceAccountKey string
	AdminEmail        string
	Domain            string
}

// AccessControlConfig is the configuration for access control
type AccessControlConfig struct {
	AllowedDomains []string
	BlockedEmails  []string
	RoleMapping    map[string][]string
	DefaultRole    string
}

// SessionConfig is the configuration for session management
type SessionConfig struct {
	TTL           int // in seconds
	RememberMeTTL int // in seconds
	Backend       string
}

// BlacklistConfig is the configuration for token blacklist
type BlacklistConfig struct {
	Enabled   bool
	Backend   string
	KeyPrefix string
}

// KafkaConfig is the configuration for Kafka
type KafkaConfig struct {
	Brokers []string
	Topic   string
}

// RedisConfig is the configuration for Redis
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// CookieConfig is the configuration for the cookie
type CookieConfig struct {
	Domain         string
	Secure         bool
	SameSite       string
	MaxAge         int
	MaxAgeRemember int
	Name           string
}

// JWTConfig is the configuration for JWT
type JWTConfig struct {
	Algorithm      string
	Issuer         string
	Audience       []string
	PrivateKeyPath string
	PublicKeyPath  string
	TTL            int // in seconds
}

// HTTPServerConfig is the configuration for the HTTP server
type HTTPServerConfig struct {
	Host string
	Port int
	Mode string
}

// LoggerConfig is the configuration for the logger
type LoggerConfig struct {
	Level        string
	Mode         string
	Encoding     string
	ColorEnabled bool
}

// PostgresConfig is the configuration for Postgres
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type DiscordConfig struct {
	WebhookID    string
	WebhookToken string
}

// EncrypterConfig is the configuration for the encrypter
type EncrypterConfig struct {
	Key string
}

// InternalConfig is the configuration for internal service authentication
type InternalConfig struct {
	ServiceKeys map[string]string
}

// Load loads configuration using Viper
func Load() (*Config, error) {
	// Set config file name and paths
	viper.SetConfigName("auth-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/smap/")

	// Enable environment variable override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	// Read config file (optional - will use env vars if file not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using environment variables
	}

	cfg := &Config{}

	// Environment
	cfg.Environment.Name = viper.GetString("environment.name")

	// HTTP Server
	cfg.HTTPServer.Host = viper.GetString("http_server.host")
	cfg.HTTPServer.Port = viper.GetInt("http_server.port")
	cfg.HTTPServer.Mode = viper.GetString("http_server.mode")

	// Logger
	cfg.Logger.Level = viper.GetString("logger.level")
	cfg.Logger.Mode = viper.GetString("logger.mode")
	cfg.Logger.Encoding = viper.GetString("logger.encoding")
	cfg.Logger.ColorEnabled = viper.GetBool("logger.color_enabled")

	// Postgres
	cfg.Postgres.Host = viper.GetString("postgres.host")
	cfg.Postgres.Port = viper.GetInt("postgres.port")
	cfg.Postgres.User = viper.GetString("postgres.user")
	cfg.Postgres.Password = viper.GetString("postgres.password")
	cfg.Postgres.DBName = viper.GetString("postgres.dbname")
	cfg.Postgres.SSLMode = viper.GetString("postgres.sslmode")

	// Redis
	cfg.Redis.Host = viper.GetString("redis.host")
	cfg.Redis.Port = viper.GetInt("redis.port")
	cfg.Redis.Password = viper.GetString("redis.password")
	cfg.Redis.DB = viper.GetInt("redis.db")

	// Kafka
	cfg.Kafka.Brokers = viper.GetStringSlice("kafka.brokers")
	cfg.Kafka.Topic = viper.GetString("kafka.topic")

	// OAuth2
	cfg.OAuth2.Provider = viper.GetString("oauth2.provider")
	cfg.OAuth2.ClientID = viper.GetString("oauth2.client_id")
	cfg.OAuth2.ClientSecret = viper.GetString("oauth2.client_secret")
	cfg.OAuth2.RedirectURI = viper.GetString("oauth2.redirect_uri")
	cfg.OAuth2.Scopes = viper.GetStringSlice("oauth2.scopes")

	// Google Workspace
	cfg.GoogleWorkspace.ServiceAccountKey = viper.GetString("google_workspace.service_account_key")
	cfg.GoogleWorkspace.AdminEmail = viper.GetString("google_workspace.admin_email")
	cfg.GoogleWorkspace.Domain = viper.GetString("google_workspace.domain")

	// JWT
	cfg.JWT.Algorithm = viper.GetString("jwt.algorithm")
	cfg.JWT.Issuer = viper.GetString("jwt.issuer")
	cfg.JWT.Audience = viper.GetStringSlice("jwt.audience")
	cfg.JWT.PrivateKeyPath = viper.GetString("jwt.private_key_path")
	cfg.JWT.PublicKeyPath = viper.GetString("jwt.public_key_path")
	cfg.JWT.TTL = viper.GetInt("jwt.ttl")

	// Cookie
	cfg.Cookie.Domain = viper.GetString("cookie.domain")
	cfg.Cookie.Secure = viper.GetBool("cookie.secure")
	cfg.Cookie.SameSite = viper.GetString("cookie.samesite")
	cfg.Cookie.MaxAge = viper.GetInt("cookie.max_age")
	cfg.Cookie.MaxAgeRemember = viper.GetInt("cookie.max_age_remember")
	cfg.Cookie.Name = viper.GetString("cookie.name")

	// Access Control
	cfg.AccessControl.AllowedDomains = viper.GetStringSlice("access_control.allowed_domains")
	cfg.AccessControl.BlockedEmails = viper.GetStringSlice("access_control.blocked_emails")
	cfg.AccessControl.DefaultRole = viper.GetString("access_control.default_role")

	// Role mapping
	roleMapping := make(map[string][]string)
	if viper.IsSet("access_control.role_mapping") {
		roleMappingRaw := viper.GetStringMap("access_control.role_mapping")
		for role, groups := range roleMappingRaw {
			if groupList, ok := groups.([]interface{}); ok {
				strGroups := make([]string, len(groupList))
				for i, g := range groupList {
					strGroups[i] = fmt.Sprint(g)
				}
				roleMapping[role] = strGroups
			}
		}
	}
	cfg.AccessControl.RoleMapping = roleMapping

	// Session
	cfg.Session.TTL = viper.GetInt("session.ttl")
	cfg.Session.RememberMeTTL = viper.GetInt("session.remember_me_ttl")
	cfg.Session.Backend = viper.GetString("session.backend")

	// Blacklist
	cfg.Blacklist.Enabled = viper.GetBool("blacklist.enabled")
	cfg.Blacklist.Backend = viper.GetString("blacklist.backend")
	cfg.Blacklist.KeyPrefix = viper.GetString("blacklist.key_prefix")

	// Encrypter
	cfg.Encrypter.Key = viper.GetString("encrypter.key")

	// Internal Service Keys
	serviceKeys := make(map[string]string)
	if viper.IsSet("internal.service_keys") {
		serviceKeysRaw := viper.GetStringMapString("internal.service_keys")
		for service, key := range serviceKeysRaw {
			serviceKeys[service] = key
		}
	}
	cfg.InternalConfig.ServiceKeys = serviceKeys

	// Discord
	cfg.Discord.WebhookID = viper.GetString("discord.webhook_id")
	cfg.Discord.WebhookToken = viper.GetString("discord.webhook_token")

	// Validate required fields
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults() {
	// Environment
	viper.SetDefault("environment.name", "production")

	// HTTP Server
	viper.SetDefault("http_server.host", "")
	viper.SetDefault("http_server.port", 8080)
	viper.SetDefault("http_server.mode", "debug")

	// Logger
	viper.SetDefault("logger.level", "debug")
	viper.SetDefault("logger.mode", "debug")
	viper.SetDefault("logger.encoding", "console")
	viper.SetDefault("logger.color_enabled", true)

	// Postgres
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.password", "postgres")
	viper.SetDefault("postgres.dbname", "postgres")
	viper.SetDefault("postgres.sslmode", "prefer")

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Kafka
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topic", "audit.events")

	// OAuth2
	viper.SetDefault("oauth2.provider", "google")
	viper.SetDefault("oauth2.scopes", []string{"openid", "email", "profile"})

	// JWT
	viper.SetDefault("jwt.algorithm", "RS256")
	viper.SetDefault("jwt.issuer", "smap-auth-service")
	viper.SetDefault("jwt.audience", []string{"smap-api"})
	viper.SetDefault("jwt.ttl", 28800) // 8 hours

	// Cookie
	viper.SetDefault("cookie.domain", ".smap.com")
	viper.SetDefault("cookie.secure", true)
	viper.SetDefault("cookie.samesite", "Lax")
	viper.SetDefault("cookie.max_age", 28800)           // 8 hours
	viper.SetDefault("cookie.max_age_remember", 604800) // 7 days
	viper.SetDefault("cookie.name", "smap_auth_token")

	// Access Control
	viper.SetDefault("access_control.default_role", "VIEWER")

	// Session
	viper.SetDefault("session.ttl", 28800)              // 8 hours
	viper.SetDefault("session.remember_me_ttl", 604800) // 7 days
	viper.SetDefault("session.backend", "redis")

	// Blacklist
	viper.SetDefault("blacklist.enabled", true)
	viper.SetDefault("blacklist.backend", "redis")
	viper.SetDefault("blacklist.key_prefix", "blacklist:")
}

func validate(cfg *Config) error {
	// Validate required OAuth2 fields
	if cfg.OAuth2.ClientID == "" {
		return fmt.Errorf("oauth2.client_id is required")
	}
	if cfg.OAuth2.ClientSecret == "" {
		return fmt.Errorf("oauth2.client_secret is required")
	}
	if cfg.OAuth2.RedirectURI == "" {
		return fmt.Errorf("oauth2.redirect_uri is required")
	}

	// Validate JWT fields
	if cfg.JWT.PrivateKeyPath == "" {
		return fmt.Errorf("jwt.private_key_path is required")
	}
	if cfg.JWT.PublicKeyPath == "" {
		return fmt.Errorf("jwt.public_key_path is required")
	}

	// Validate Google Workspace fields
	if cfg.GoogleWorkspace.ServiceAccountKey == "" {
		return fmt.Errorf("google_workspace.service_account_key is required")
	}
	if cfg.GoogleWorkspace.AdminEmail == "" {
		return fmt.Errorf("google_workspace.admin_email is required")
	}
	if cfg.GoogleWorkspace.Domain == "" {
		return fmt.Errorf("google_workspace.domain is required")
	}

	// Validate Access Control
	if len(cfg.AccessControl.AllowedDomains) == 0 {
		return fmt.Errorf("access_control.allowed_domains must have at least one domain")
	}

	// Validate Encrypter
	if cfg.Encrypter.Key == "" {
		return fmt.Errorf("encrypter.key is required")
	}

	return nil
}
