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

	// TODO: Register other scheduled jobs here
	// if err := srv.registerReportGenerationJob(ctx); err != nil {
	//     return err
	// }

	return nil
}

// registerAuditCleanupJob registers the audit log cleanup job
func (srv *Scheduler) registerAuditCleanupJob(ctx context.Context) error {
	// Initialize audit repository
	auditRepo := auditPostgre.New(srv.postgresDB)

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
