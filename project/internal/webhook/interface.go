package webhook

import (
	"context"
)

//go:generate mockery --name UseCase
type UseCase interface {
	HandleDryRunCallback(ctx context.Context, req CallbackRequest) error
	StoreJobMapping(ctx context.Context, jobID, userID, projectID string) error
}
