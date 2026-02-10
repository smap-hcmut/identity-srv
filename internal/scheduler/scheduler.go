package scheduler

import (
	"context"
	"database/sql"
	"errors"

	"smap-api/config"
	"smap-api/internal/audit"
	"smap-api/internal/authentication/repository"
	"smap-api/pkg/jwt/rotation"
	pkgLog "smap-api/pkg/log"
)

// Job represents a scheduled job interface
type Job interface {
	Stop()
}

// Scheduler represents the scheduler service
type Scheduler struct {
	logger          pkgLog.Logger
	postgresDB      *sql.DB
	config          *config.Config
	jwtKeysRepo     *repository.JWTKeysRepository
	rotationManager *rotation.Manager
	auditPublisher  audit.Publisher
	jobs            []Job
}

// Config holds scheduler service configuration
type Config struct {
	Logger          pkgLog.Logger
	PostgresDB      *sql.DB
	Config          *config.Config
	JWTKeysRepo     *repository.JWTKeysRepository
	RotationManager *rotation.Manager
	AuditPublisher  audit.Publisher
}

// New creates a new scheduler service
func New(logger pkgLog.Logger, cfg Config) (*Scheduler, error) {
	srv := &Scheduler{
		logger:          logger,
		postgresDB:      cfg.PostgresDB,
		config:          cfg.Config,
		jwtKeysRepo:     cfg.JWTKeysRepo,
		rotationManager: cfg.RotationManager,
		auditPublisher:  cfg.AuditPublisher,
		jobs:            make([]Job, 0),
	}

	if err := srv.validate(); err != nil {
		return nil, err
	}

	return srv, nil
}

// validate validates that all required dependencies are provided
func (srv *Scheduler) validate() error {
	if srv.logger == nil {
		return errors.New("logger is required")
	}
	if srv.postgresDB == nil {
		return errors.New("postgresDB is required")
	}
	return nil
}

// Start starts all scheduled jobs
func (srv *Scheduler) Start() error {
	ctx := context.Background()

	// Register all jobs (similar to mapHandlers in httpserver)
	if err := srv.registerJobs(); err != nil {
		return err
	}

	srv.logger.Infof(ctx, "Scheduler service started with %d job(s)", len(srv.jobs))

	return nil
}

// Stop stops all scheduled jobs
func (srv *Scheduler) Stop() {
	ctx := context.Background()
	srv.logger.Info(ctx, "Stopping scheduler service...")

	for _, job := range srv.jobs {
		job.Stop()
	}

	srv.logger.Info(ctx, "Scheduler service stopped")
}
