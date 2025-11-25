package models

import "time"

// TaskType liệt kê các loại tác vụ crawler hỗ trợ.
type TaskType string

const (
	TaskTypeResearchKeyword  TaskType = "research_keyword"
	TaskTypeCrawlLinks       TaskType = "crawl_links"
	TaskTypeResearchAndCrawl TaskType = "research_and_crawl"
)

// Platform liệt kê các worker/platform hiện có.
type Platform string

const (
	PlatformYouTube Platform = "youtube"
	PlatformTikTok  Platform = "tiktok"
)

// CollectorTask là hợp đồng chuẩn hóa mà collector publish tới từng worker.
// Payload sẽ được map sang struct riêng theo platform + task_type.
type CollectorTask struct {
	JobID         string         `json:"job_id"`
	Platform      Platform       `json:"platform"`
	TaskType      TaskType       `json:"task_type"`
	Payload       any            `json:"payload"`
	TimeRange     int            `json:"time_range,omitempty"`
	Attempt       int            `json:"attempt,omitempty"`
	MaxAttempts   int            `json:"max_attempts,omitempty"`
	Retry         bool           `json:"retry,omitempty"`
	SchemaVersion int            `json:"schema_version,omitempty"`
	TraceID       string         `json:"trace_id,omitempty"`
	RoutingKey    string         `json:"routing_key,omitempty"`
	EmittedAt     time.Time      `json:"emitted_at"`
	Headers       map[string]any `json:"headers,omitempty"`
}

// --- YouTube payloads ---

type YouTubeResearchKeywordPayload struct {
	Keyword   string `json:"keyword"`
	Limit     int    `json:"limit,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	TimeRange int    `json:"time_range,omitempty"`
}

type YouTubeCrawlLinksPayload struct {
	VideoURLs       []string `json:"video_urls"`
	IncludeChannel  bool     `json:"include_channel,omitempty"`
	IncludeComments bool     `json:"include_comments,omitempty"`
	MaxComments     int      `json:"max_comments,omitempty"`
	DownloadMedia   bool     `json:"download_media,omitempty"`
	MediaType       string   `json:"media_type,omitempty"`
	TimeRange       int      `json:"time_range,omitempty"`
}

type YouTubeResearchAndCrawlPayload struct {
	Keywords        []string `json:"keywords"`
	LimitPerKeyword int      `json:"limit_per_keyword,omitempty"`
	IncludeComments bool     `json:"include_comments,omitempty"`
	IncludeChannel  bool     `json:"include_channel,omitempty"`
	MaxComments     int      `json:"max_comments,omitempty"`
	DownloadMedia   bool     `json:"download_media,omitempty"`
	TimeRange       int      `json:"time_range,omitempty"`
}

// --- TikTok payloads ---

type TikTokResearchKeywordPayload struct {
	Keyword   string `json:"keyword"`
	Limit     int    `json:"limit,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	TimeRange int    `json:"time_range,omitempty"`
}

type TikTokCrawlLinksPayload struct {
	VideoURLs       []string `json:"video_urls"`
	IncludeComments bool     `json:"include_comments,omitempty"`
	IncludeCreator  bool     `json:"include_creator,omitempty"`
	MaxComments     int      `json:"max_comments,omitempty"`
	DownloadMedia   bool     `json:"download_media,omitempty"`
	MediaType       string   `json:"media_type,omitempty"`
	MediaSaveDir    string   `json:"media_save_dir,omitempty"`
	TimeRange       int      `json:"time_range,omitempty"`
}

type TikTokResearchAndCrawlPayload struct {
	Keywords        []string `json:"keywords"`
	LimitPerKeyword int      `json:"limit_per_keyword,omitempty"`
	SortBy          string   `json:"sort_by,omitempty"`
	IncludeComments bool     `json:"include_comments,omitempty"`
	IncludeCreator  bool     `json:"include_creator,omitempty"`
	MaxComments     int      `json:"max_comments,omitempty"`
	DownloadMedia   bool     `json:"download_media,omitempty"`
	MediaType       string   `json:"media_type,omitempty"`
	MediaSaveDir    string   `json:"media_save_dir,omitempty"`
	TimeRange       int      `json:"time_range,omitempty"`
}
