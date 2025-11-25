package dispatcher

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name=UseCase
type UseCase interface {
	// Dispatch nhận CrawlRequest, chuẩn hóa thành các CollectorTask (kèm payload typed) và publish tới từng worker queue theo strategy (fan-out khi platform trống).
	Dispatch(ctx context.Context, req models.CrawlRequest) ([]models.CollectorTask, error)
}
