package usecase

import (
	"context"
	"smap-api/internal/audit"
)

// PublishAuditEvent publishes an audit event (non-blocking)
func (u *implUsecase) PublishAuditEvent(ctx context.Context, event audit.AuditEvent) {
	if u.auditPublisher == nil {
		u.l.Warnf(ctx, "Audit publisher not configured, skipping audit event")
		return
	}

	if err := u.auditPublisher.Publish(ctx, event); err != nil {
		u.l.Errorf(ctx, "Failed to publish audit event: %v", err)
	}
}
