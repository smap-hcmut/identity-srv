package consumer

import (
	"identity-srv/internal/audit/repository"
	"time"

	"github.com/smap-hcmut/shared-libs/go/log"
)

// Consumer consumes audit events from Kafka and stores them in database
type Consumer struct {
	repo         repository.Repository
	logger       log.Logger
	batchSize    int
	batchTimeout time.Duration
}

// Config holds consumer configuration
type Config struct {
	BatchSize    int
	BatchTimeout time.Duration
}

// New creates a new audit consumer
func New(repo repository.Repository, cfg Config, logger log.Logger) *Consumer {
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
