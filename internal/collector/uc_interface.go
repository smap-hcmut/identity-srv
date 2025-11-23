package collector

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name=ResultUseCase
type ResultUseCase interface {
	// HandleResult xử lý kết quả từ worker, quyết định retry/ack và cập nhật trạng thái.
	HandleResult(ctx context.Context, res models.CrawlerResult) error
}

// UseCase export cho module collector (fan-in + retry).
type UseCase interface {
	ResultUseCase
}
