package scheduler

import (
	"context"
	"time"

	"smap-api/internal/audit"
	"smap-api/pkg/jwt/rotation"
	pkgLog "smap-api/pkg/log"

	"github.com/robfig/cron/v3"
)

// KeyRotationJob handles JWT key rotation on a schedule
type KeyRotationJob struct {
	manager        *rotation.Manager
	logger         pkgLog.Logger
	auditPublisher audit.Publisher
	cron           *cron.Cron
}

// NewKeyRotationJob creates a new key rotation job
func NewKeyRotationJob(manager *rotation.Manager, logger pkgLog.Logger, auditPublisher audit.Publisher) *KeyRotationJob {
	return &KeyRotationJob{
		manager:        manager,
		logger:         logger,
		auditPublisher: auditPublisher,
		cron:           cron.New(),
	}
}

// Start starts the key rotation job (runs daily at 3 AM)
func (j *KeyRotationJob) Start() error {
	ctx := context.Background()

	// Schedule job to run daily at 3 AM
	_, err := j.cron.AddFunc("0 3 * * *", func() {
		j.run()
	})
	if err != nil {
		return err
	}

	j.cron.Start()
	j.logger.Info(ctx, "Key rotation job scheduled (daily at 3 AM)")

	return nil
}

// Stop stops the key rotation job
func (j *KeyRotationJob) Stop() {
	ctx := context.Background()
	j.logger.Info(ctx, "Stopping key rotation job...")
	j.cron.Stop()
	j.logger.Info(ctx, "Key rotation job stopped")
}

// run executes the key rotation logic
func (j *KeyRotationJob) run() {
	ctx := context.Background()
	startTime := time.Now()

	j.logger.Info(ctx, "Starting key rotation check...")

	// Perform key rotation
	if err := j.manager.RotateKeys(ctx); err != nil {
		j.logger.Errorf(ctx, "Key rotation failed: %v", err)

		// Publish audit event for failed rotation
		if j.auditPublisher != nil {
			j.auditPublisher.Publish(ctx, audit.AuditEvent{
				UserID:       "system",
				Action:       "key_rotation_failed",
				ResourceType: "jwt_keys",
				Metadata: map[string]string{
					"error": err.Error(),
				},
			})
		}
		return
	}

	duration := time.Since(startTime)
	j.logger.Infof(ctx, "Key rotation check completed successfully (took %v)", duration)

	// Publish audit event for successful rotation
	if j.auditPublisher != nil {
		j.auditPublisher.Publish(ctx, audit.AuditEvent{
			UserID:       "system",
			Action:       "key_rotation_completed",
			ResourceType: "jwt_keys",
			Metadata: map[string]string{
				"duration_ms": duration.String(),
			},
		})
	}
}
