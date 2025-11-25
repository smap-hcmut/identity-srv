package dispatcher

import (
	"context"

	"smap-collector/internal/models"
)

//go:generate mockery --name=UseCase
type UseCase interface {
	// Dispatch nhận CrawlRequest, chuẩn hóa thành các CollectorTask (kèm payload typed) và publish tới từng worker queue theo strategy (fan-out khi platform trống).
	Dispatch(ctx context.Context, req models.CrawlRequest) ([]models.CollectorTask, error)
	Producer
}

// Producer để usecase gọi publish task (implemented ở delivery layer).
type Producer interface {
	PublishTask(ctx context.Context, task models.CollectorTask) error
}
