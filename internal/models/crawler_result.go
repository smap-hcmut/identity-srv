package models

import "time"

// CrawlerStatus mô tả trạng thái trả về của worker.
type CrawlerStatus string

const (
	CrawlerStatusSuccess CrawlerStatus = "success"
	CrawlerStatusSkipped CrawlerStatus = "skipped"
	CrawlerStatusFailed  CrawlerStatus = "failed"
)

// CrawlerResult là payload worker gửi ngược collector để quyết định retry/hoàn thành.
type CrawlerResult struct {
	JobID     string         `json:"job_id"`
	Platform  Platform       `json:"platform"`
	TaskType  TaskType       `json:"task_type"`
	Status    CrawlerStatus  `json:"status"`
	Payload   any            `json:"payload,omitempty"`
	Cursor    string         `json:"cursor,omitempty"`
	Metrics   ResultMetrics  `json:"metrics,omitempty"`
	Errors    []ResultError  `json:"errors,omitempty"`
	Attempt   int            `json:"attempt,omitempty"`
	EmittedAt time.Time      `json:"emitted_at"`
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
