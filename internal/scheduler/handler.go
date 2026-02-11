package scheduler

import (
	"context"

	auditPostgre "smap-api/internal/audit/repository/postgre"
	"smap-api/internal/audit/usecase"
)

// registerJobs registers all scheduled jobs (similar to mapHandlers in httpserver)
func (srv *Scheduler) registerJobs() error {
	ctx := context.Background()

	// Register audit cleanup job
	if err := srv.registerAuditCleanupJob(ctx); err != nil {
		return err
	}

	// Register key rotation job
	if err := srv.registerKeyRotationJob(ctx); err != nil {
		return err
	}

	return nil
}

// registerAuditCleanupJob registers the audit log cleanup job
func (srv *Scheduler) registerAuditCleanupJob(ctx context.Context) error {
	// Initialize audit repository
	auditRepo := auditPostgre.New(srv.logger, srv.postgresDB)

	// Initialize audit log cleanup job
	cleanupJob := usecase.NewCleanupJob(auditRepo, srv.logger)

	// Start the job
	if err := cleanupJob.Start(); err != nil {
		return err
	}

	srv.logger.Info(ctx, "Audit log cleanup job started (runs daily at 2 AM)")

	// Store job for cleanup on shutdown
	srv.jobs = append(srv.jobs, cleanupJob)

	return nil
}

// TODO: Add more job registration methods here
// Example:
// func (srv *Scheduler) registerReportGenerationJob(ctx context.Context) error {
//     reportJob := reports.NewGenerationJob(srv.postgresDB, srv.logger)
//     if err := reportJob.Start(); err != nil {
//         return err
//     }
//     srv.jobs = append(srv.jobs, reportJob)
//     return nil
// }

// registerKeyRotationJob registers the JWT key rotation job
func (srv *Scheduler) registerKeyRotationJob(ctx context.Context) error {
	// Check if key rotation is enabled
	if !srv.config.KeyRotation.Enabled {
		srv.logger.Info(ctx, "Key rotation is disabled in configuration")
		return nil
	}

	// Ensure at least one active key exists on startup
	if err := srv.rotationManager.EnsureActiveKey(ctx); err != nil {
		return err
	}

	// Create key rotation job
	rotationJob := NewKeyRotationJob(srv.rotationManager, srv.logger, srv.auditPublisher)

	// Start the job (runs daily at 3 AM)
	if err := rotationJob.Start(); err != nil {
		return err
	}

	srv.logger.Infof(ctx, "Key rotation job started (runs daily at 3 AM, rotation period: %d days, grace period: %d days)",
		srv.config.KeyRotation.RotationDays, srv.config.KeyRotation.GraceDays)

	// Store job for cleanup on shutdown
	srv.jobs = append(srv.jobs, rotationJob)

	return nil
}
