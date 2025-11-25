package dispatcher

import "github.com/nguyentantai21042004/smap-api/internal/models"

// Options for dispatcher defaults.
type Options struct {
	DefaultMaxAttempts int
	SchemaVersion      int
	// PlatformQueues map platform -> queue name để fan-out (khi platform trống sẽ gửi tất cả).
	PlatformQueues map[models.Platform]string
}
