package consumer

import (
	"identity-srv/internal/audit/repository"
	pkgLog "identity-srv/pkg/log"
	"time"
)

// Consumer consumes audit events from Kafka and stores them in database
type Consumer struct {
	repo         repository.Repository
	logger       pkgLog.Logger
	batchSize    int
	batchTimeout time.Duration
}

// Config holds consumer configuration
type Config struct {
	BatchSize    int
	BatchTimeout time.Duration
}

// New creates a new audit consumer
func New(repo repository.Repository, cfg Config, logger pkgLog.Logger) *Consumer {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 5 * time.Second
	}

	return &Consumer{
		repo:         repo,
		logger:       logger,
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
	}
}
