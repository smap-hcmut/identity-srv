package usecase

import (
	"context"

	"smap-collector/internal/models"
)

// Producer triển khai interface dispatcher.Producer ngay trong implUseCase.
func (uc implUseCase) PublishTask(ctx context.Context, task models.CollectorTask) error {
	return uc.prod.PublishTask(ctx, task)
}
