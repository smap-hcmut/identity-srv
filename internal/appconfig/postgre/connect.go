package postgre

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nguyentantai21042004/smap-api/config"
)

const (
	defaultConnectTimeout  = 5 * time.Second  // Giảm từ 10s xuống 5s
	defaultMaxIdleConns    = 25               // Tăng từ 10 lên 25
	defaultMaxOpenConns    = 200              // Tăng từ 100 lên 200
	defaultConnMaxLifetime = 30 * time.Minute // Giảm từ 1h xuống 30m
	defaultConnMaxIdleTime = 5 * time.Minute  // Tăng từ 1m lên 5m
)

var (
	instance *sql.DB
	once     sync.Once
	mu       sync.RWMutex
)

// Connect initializes and connects to PostgreSQL with retry logic
func Connect(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	var err error

	once.Do(func() {
		// Set timeout for connection
		connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
		defer cancel()

		// Create connection string with SSL options
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require sslcert='' sslkey='' sslrootcert=''",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

		// Open database connection
		db, dbErr := sql.Open("postgres", dsn)
		if dbErr != nil {
			err = fmt.Errorf("failed to open PostgreSQL connection: %w", dbErr)
			return
		}

		// Configure connection pool
		db.SetMaxIdleConns(defaultMaxIdleConns)
		db.SetMaxOpenConns(defaultMaxOpenConns)
		db.SetConnMaxLifetime(defaultConnMaxLifetime)
		db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

		// Test connection
		if pingErr := db.PingContext(connectCtx); pingErr != nil {
			err = fmt.Errorf("failed to ping PostgreSQL: %w", pingErr)
			return
		}

		instance = db
	})

	return instance, err
}

// GetClient returns the singleton PostgreSQL client instance
func GetClient() *sql.DB {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		panic("PostgreSQL client not initialized. Call Connect() first")
	}
	return instance
}

// Disconnect closes the PostgreSQL connection
func Disconnect(ctx context.Context, db *sql.DB) error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		log.Println("Disconnecting from PostgreSQL...")

		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}

		instance = nil
		log.Println("PostgreSQL disconnected successfully")
	}
	return nil
}

// HealthCheck performs a health check on the PostgreSQL connection
func HealthCheck(ctx context.Context) error {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return fmt.Errorf("PostgreSQL client not initialized")
	}

	// Simple ping to check connection
	if err := instance.PingContext(ctx); err != nil {
		return fmt.Errorf("PostgreSQL health check failed: %w", err)
	}

	return nil
}

// IsConnected checks if the PostgreSQL client is connected
func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()

	return instance != nil
}

// Reconnect reinitializes the PostgreSQL connection
func Reconnect(ctx context.Context, cfg config.PostgresConfig) error {
	mu.Lock()
	defer mu.Unlock()

	// Close existing connection
	if instance != nil {
		if err := instance.Close(); err != nil {
			log.Printf("Warning: failed to close existing PostgreSQL connection: %v", err)
		}
		instance = nil
	}

	// Create new connection
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to create new PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	// Test connection
	connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	if pingErr := db.PingContext(connectCtx); pingErr != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", pingErr)
	}

	instance = db
	log.Println("PostgreSQL reconnected successfully")

	return nil
}
