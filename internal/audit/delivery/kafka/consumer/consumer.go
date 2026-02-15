package consumer

import (
	"identity-srv/internal/audit"
	pkgLog "identity-srv/pkg/log"
	"time"

	"github.com/IBM/sarama"
)

// GetTopics returns the topics this consumer wants to subscribe to
func (c *Consumer) GetTopics() []string {
	return []string{"audit.events"}
}

// GetGroupID returns the consumer group ID
func (c *Consumer) GetGroupID() string {
	return "audit-consumer-group"
}

// CreateHandler creates a Sarama consumer group handler
func (c *Consumer) CreateHandler() sarama.ConsumerGroupHandler {
	return &consumerGroupHandler{
		consumer:  c,
		logger:    c.logger,
		batch:     make([]audit.AuditEvent, 0, c.batchSize),
		lastFlush: time.Now(),
	}
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer  *Consumer
	logger    pkgLog.Logger
	batch     []audit.AuditEvent
	lastFlush time.Time
}
