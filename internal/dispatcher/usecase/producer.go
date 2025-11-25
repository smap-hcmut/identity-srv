package usecase

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

// Producer triển khai interface dispatcher.Producer ngay trong implUseCase.
func (uc implUseCase) PublishTask(ctx context.Context, task models.CollectorTask) error {
	return uc.prod.PublishTask(ctx, task)
}
