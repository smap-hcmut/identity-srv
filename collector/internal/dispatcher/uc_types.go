package dispatcher

import "smap-collector/internal/models"

// Options for dispatcher defaults.
type Options struct {
	DefaultMaxAttempts int
	SchemaVersion      int
	// PlatformQueues map platform -> queue name để fan-out (khi platform trống sẽ gửi tất cả).
	PlatformQueues map[models.Platform]string
}
