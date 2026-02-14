package usecase

import (
	"context"
	"time"
)

// Start starts the cleanup job (runs daily at 2 AM)
func (j *CleanupJob) Start() error {
	// Schedule cleanup job to run daily at 2 AM
	_, err := j.cron.AddFunc("0 2 * * *", func() {
		ctx := context.Background()
		j.logger.Infof(ctx, "Starting audit log cleanup job...")

		startTime := time.Now()
		deleted, err := j.repo.DeleteExpired(ctx)
		if err != nil {
			j.logger.Errorf(ctx, "Failed to delete expired audit logs: %v", err)
			return
		}

		duration := time.Since(startTime)
		j.logger.Infof(ctx, "Audit log cleanup completed: deleted %d records in %v", deleted, duration)
	})

	if err != nil {
		return err
	}

	j.cron.Start()
	j.logger.Infof(context.Background(), "Audit log cleanup job scheduled (daily at 2 AM)")

	return nil
}

// Stop stops the cleanup job
func (j *CleanupJob) Stop() {
	if j.cron != nil {
		j.cron.Stop()
	}
}

// RunNow runs the cleanup job immediately (for testing)
func (j *CleanupJob) RunNow(ctx context.Context) (int64, error) {
	j.logger.Infof(ctx, "Running audit log cleanup job manually...")

	startTime := time.Now()
	deleted, err := j.repo.DeleteExpired(ctx)
	if err != nil {
		j.logger.Errorf(ctx, "Failed to delete expired audit logs: %v", err)
		return 0, err
	}

	duration := time.Since(startTime)
	j.logger.Infof(ctx, "Audit log cleanup completed: deleted %d records in %v", deleted, duration)

	return deleted, nil
}
