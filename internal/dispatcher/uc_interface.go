package dispatcher

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name=UseCase
type UseCase interface {
	// Dispatch nhận CrawlRequest, chuẩn hóa thành CollectorTask (kèm payload typed) và publish tới worker queue.
	Dispatch(ctx context.Context, req models.CrawlRequest) (models.CollectorTask, error)
}
