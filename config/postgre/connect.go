package postgre

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"identity-srv/config"

	_ "github.com/lib/pq" // PostgreSQL driver
)

const (
	// defaultConnectTimeout is the maximum time to wait for initial connection
	defaultConnectTimeout = 5 * time.Second
	// defaultMaxIdleConns is the maximum number of idle connections in the pool
	defaultMaxIdleConns = 25
	// defaultMaxOpenConns is the maximum number of open connections to the database
	defaultMaxOpenConns = 200
	// defaultConnMaxLifetime is the maximum amount of time a connection may be reused
	defaultConnMaxLifetime = 30 * time.Minute
	// defaultConnMaxIdleTime is the maximum amount of time a connection may be idle
	defaultConnMaxIdleTime = 5 * time.Minute
)

var (
	instance *sql.DB
	once     sync.Once
	mu       sync.RWMutex
	initErr  error // Stores the last initialization error to allow retry
)

// Connect initializes and connects to PostgreSQL database using singleton pattern.
// If connection fails, it can be retried by calling Connect() again.
// Returns the existing connection instance if already connected.
func Connect(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	// Return existing instance if already connected
	if instance != nil {
		return instance, nil
	}

	// Reset sync.Once if previous initialization failed to allow retry
	if initErr != nil {
		once = sync.Once{}
		initErr = nil
	}

	var err error
	once.Do(func() {
		// Create context with timeout for connection attempt
		connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
		defer cancel()

		// Build connection string with configurable SSL mode
		// Supported modes: disable, require, verify-ca, verify-full
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "disable" // Default to disable for local development
		}
		searchPath := cfg.Schema
		if searchPath == "" {
			searchPath = "public"
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, sslMode, searchPath)

		fmt.Printf("[PostgreSQL] Attempting to connect to %s:%d/%s (SSL mode: %s)...\n",
			cfg.Host, cfg.Port, cfg.DBName, sslMode)

		// Open database connection (does not actually connect yet)
		db, dbErr := sql.Open("postgres", dsn)
		if dbErr != nil {
			err = fmt.Errorf("failed to open PostgreSQL connection: %w", dbErr)
			initErr = err
			fmt.Printf("[PostgreSQL] ERROR: Failed to open connection: %v\n", err)
			return
		}

		// Configure connection pool settings
		db.SetMaxIdleConns(defaultMaxIdleConns)
		db.SetMaxOpenConns(defaultMaxOpenConns)
		db.SetConnMaxLifetime(defaultConnMaxLifetime)
		db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

		fmt.Printf("[PostgreSQL] Pinging database...\n")

		// Verify connection by pinging the database
		if pingErr := db.PingContext(connectCtx); pingErr != nil {
			// Close connection to prevent resource leak
			_ = db.Close()
			err = fmt.Errorf("failed to ping PostgreSQL: %w", pingErr)
			initErr = err
			fmt.Printf("[PostgreSQL] ERROR: Failed to ping database: %v\n", pingErr)
			return
		}

		instance = db
		fmt.Printf("[PostgreSQL] Successfully connected to %s:%d/%s\n",
			cfg.Host, cfg.Port, cfg.DBName)
	})

	return instance, err
}

// Disconnect closes the PostgreSQL connection and resets the singleton instance.
// This allows a new connection to be established by calling Connect() again.
func Disconnect(ctx context.Context, db *sql.DB) error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		fmt.Printf("[PostgreSQL] Disconnecting...\n")
		if err := db.Close(); err != nil {
			fmt.Printf("[PostgreSQL] ERROR: Failed to close connection: %v\n", err)
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}

		instance = nil
		initErr = nil
		once = sync.Once{} // Reset to allow reconnection
		fmt.Printf("[PostgreSQL] Disconnected successfully\n")
	}
	return nil
}
