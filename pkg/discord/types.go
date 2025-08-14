package discord

import (
	"context"
	"time"
)

// MessageType định nghĩa các loại message khác nhau
type MessageType string

const (
	MessageTypeInfo    MessageType = "info"
	MessageTypeSuccess MessageType = "success"
	MessageTypeWarning MessageType = "warning"
	MessageTypeError   MessageType = "error"
)

// MessageLevel định nghĩa mức độ ưu tiên của message
type MessageLevel int

const (
	LevelLow MessageLevel = iota
	LevelNormal
	LevelHigh
	LevelUrgent
)

// EmbedField đại diện cho một field trong Discord embed
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// EmbedFooter đại diện cho footer của Discord embed
type EmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// EmbedAuthor đại diện cho author của Discord embed
type EmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// Embed đại diện cho Discord embed
type Embed struct {
	Name        string          `json:"Name,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Color       int             `json:"color,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Footer      *EmbedFooter    `json:"footer,omitempty"`
	Author      *EmbedAuthor    `json:"author,omitempty"`
	Fields      []EmbedField    `json:"fields,omitempty"`
	Thumbnail   *EmbedThumbnail `json:"thumbnail,omitempty"`
	Image       *EmbedImage     `json:"image,omitempty"`
}

// EmbedThumbnail đại diện cho thumbnail của embed
type EmbedThumbnail struct {
	URL string `json:"url"`
}

// EmbedImage đại diện cho image của embed
type EmbedImage struct {
	URL string `json:"url"`
}

// WebhookPayload đại diện cho payload gửi đến Discord webhook
type WebhookPayload struct {
	Content   string  `json:"content,omitempty"`
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Embeds    []Embed `json:"embeds,omitempty"`
}

// MessageOptions chứa các tùy chọn cho message
type MessageOptions struct {
	Type        MessageType
	Level       MessageLevel
	Name        string
	Description string
	Fields      []EmbedField
	Footer      *EmbedFooter
	Author      *EmbedAuthor
	Thumbnail   *EmbedThumbnail
	Image       *EmbedImage
	Username    string
	AvatarURL   string
	Timestamp   time.Time
}

// DiscordService interface định nghĩa các method cho Discord service
type DiscordService interface {
	// SendMessage gửi message đơn giản
	SendMessage(ctx context.Context, content string) error

	// SendEmbed gửi embed message với options
	SendEmbed(ctx context.Context, options MessageOptions) error

	// SendError gửi error message
	SendError(ctx context.Context, Name, description string, err error) error

	// SendSuccess gửi success message
	SendSuccess(ctx context.Context, Name, description string) error

	// SendWarning gửi warning message
	SendWarning(ctx context.Context, Name, description string) error

	// SendInfo gửi info message
	SendInfo(ctx context.Context, Name, description string) error

	// ReportBug gửi bug report (backward compatibility)
	ReportBug(ctx context.Context, message string) error
}

// Config chứa cấu hình cho Discord service
type Config struct {
	WebhookID        string
	WebhookToken     string
	Timeout          time.Duration
	RetryCount       int
	RetryDelay       time.Duration
	DefaultUsername  string
	DefaultAvatarURL string
}

// DefaultConfig trả về config mặc định
func DefaultConfig() Config {
	return Config{
		Timeout:          30 * time.Second,
		RetryCount:       3,
		RetryDelay:       1 * time.Second,
		DefaultUsername:  "Smap Bot",
		DefaultAvatarURL: "",
	}
}
