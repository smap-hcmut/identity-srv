package models

// CrawlerStatus mô tả trạng thái trả về của worker.
type CrawlerStatus string

const (
	CrawlerStatusSuccess CrawlerStatus = "success"
	CrawlerStatusSkipped CrawlerStatus = "skipped"
	CrawlerStatusFailed  CrawlerStatus = "failed"
)

// CrawlerResult là payload worker gửi ngược collector.
// Format mới chỉ có 2 fields: success và payload (array of content items).
// Metadata như job_id, platform được lấy từ payload[].meta
type CrawlerResult struct {
	Success bool `json:"success"`
	Payload any  `json:"payload"` // Array of content items with meta, content, interaction, author, comments
}

// ResultMetrics thống kê kết quả crawl.
type ResultMetrics struct {
	Documents  int64 `json:"documents,omitempty"`
	Bytes      int64 `json:"bytes,omitempty"`
	DurationMs int64 `json:"duration_ms,omitempty"`
}

// ResultError chứa mã lỗi máy đọc được từ worker.
type ResultError struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
