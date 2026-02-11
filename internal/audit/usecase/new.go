package usecase

import (
	"smap-api/internal/audit/repository"
	pkgLog "smap-api/pkg/log"

	"github.com/robfig/cron/v3"
)

// CleanupJob handles periodic cleanup of expired audit logs
type CleanupJob struct {
	repo   repository.Repository
	cron   *cron.Cron
	logger pkgLog.Logger
}

// NewCleanupJob creates a new audit log cleanup job
func NewCleanupJob(repo repository.Repository, logger pkgLog.Logger) *CleanupJob {
	return &CleanupJob{
		repo:   repo,
		cron:   cron.New(),
		logger: logger,
	}
}
