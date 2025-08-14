package minio

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/pkg/minio"
)

const (
	defaultConnectTimeout = 10 * time.Second
	defaultRetryAttempts  = 3
)

var (
	instance minio.MinIO
	once     sync.Once
	mu       sync.RWMutex
)

// Connect initializes and connects to MinIO with retry logic
func Connect(ctx context.Context, cfg config.MinIOConfig) (minio.MinIO, error) {
	var err error

	once.Do(func() {
		// Create MinIO client with retry
		client, clientErr := minio.NewMinIOWithRetry(&cfg, defaultRetryAttempts)
		if clientErr != nil {
			err = fmt.Errorf("failed to create MinIO client: %w", clientErr)
			return
		}

		// Set timeout for connection
		connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
		defer cancel()

		// Test connection
		if healthErr := client.HealthCheck(connectCtx); healthErr != nil {
			err = fmt.Errorf("failed to connect to MinIO: %w", healthErr)
			return
		}

		instance = client
		log.Printf("MinIO client initialized successfully")
	})

	return instance, err
}

// Close closes the MinIO connection
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		if err := instance.Close(); err != nil {
			return fmt.Errorf("failed to close MinIO connection: %w", err)
		}
		instance = nil
		log.Printf("MinIO client closed successfully")
	}
	return nil
}

// HealthCheck performs a health check on the MinIO connection
func HealthCheck(ctx context.Context) error {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return fmt.Errorf("MinIO client not initialized")
	}

	return instance.HealthCheck(ctx)
}

// IsConnected checks if the MinIO client is connected
func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()

	return instance != nil
}

// Reconnect reinitializes the MinIO connection
func Reconnect(ctx context.Context, cfg config.MinIOConfig) error {
	mu.Lock()
	defer mu.Unlock()

	// Close existing connection
	if instance != nil {
		if err := instance.Close(); err != nil {
			log.Printf("Warning: failed to close existing MinIO connection: %v", err)
		}
		instance = nil
	}

	// Create new connection
	client, err := minio.NewMinIOWithRetry(&cfg, defaultRetryAttempts)
	if err != nil {
		return fmt.Errorf("failed to create new MinIO client: %w", err)
	}

	// Test connection
	connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	if healthErr := client.HealthCheck(connectCtx); healthErr != nil {
		return fmt.Errorf("failed to connect to MinIO: %w", healthErr)
	}

	instance = client
	log.Printf("MinIO client reconnected successfully")

	return nil
}
