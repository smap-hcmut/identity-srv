package kafka

import (
	"context"
	"encoding/json"
	"smap-api/internal/audit"
	"smap-api/internal/audit/repository"
	"time"

	"github.com/IBM/sarama"
)

// Consumer consumes audit events from Kafka and stores them in database
type Consumer struct {
	repo         repository.Repository
	logger       Logger
	batchSize    int
	batchTimeout time.Duration
}

// Logger interface for consumer
type Logger interface {
	Infof(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
}

// Config holds consumer configuration
type Config struct {
	BatchSize    int
	BatchTimeout time.Duration
}

// New creates a new audit consumer
func New(repo repository.Repository, cfg Config, logger Logger) *Consumer {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 5 * time.Second
	}

	return &Consumer{
		repo:         repo,
		logger:       logger,
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
	}
}

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
	logger    Logger
	batch     []audit.AuditEvent
	lastFlush time.Time
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.batch = make([]audit.AuditEvent, 0, h.consumer.batchSize)
	h.lastFlush = time.Now()
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	// Flush remaining events
	if len(h.batch) > 0 {
		ctx := context.Background()
		if err := h.flushBatch(ctx); err != nil {
			h.logger.Errorf(ctx, "Failed to flush batch on cleanup: %v", err)
		}
	}
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()

	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Parse audit event
			var event audit.AuditEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.logger.Errorf(ctx, "Failed to unmarshal audit event: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			// Add to batch
			h.batch = append(h.batch, event)

			// Mark message as processed
			session.MarkMessage(message, "")

			// Check if we should flush
			if len(h.batch) >= h.consumer.batchSize || time.Since(h.lastFlush) >= h.consumer.batchTimeout {
				if err := h.flushBatch(ctx); err != nil {
					h.logger.Errorf(ctx, "Failed to flush batch: %v", err)
					// Continue processing, don't fail
				}
			}

		case <-time.After(h.consumer.batchTimeout):
			// Timeout reached, flush batch if not empty
			if len(h.batch) > 0 {
				if err := h.flushBatch(ctx); err != nil {
					h.logger.Errorf(ctx, "Failed to flush batch on timeout: %v", err)
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}

// flushBatch inserts the batch into database
func (h *consumerGroupHandler) flushBatch(ctx context.Context) error {
	if len(h.batch) == 0 {
		return nil
	}

	h.logger.Infof(ctx, "Flushing batch of %d audit events", len(h.batch))

	if err := h.consumer.repo.BatchInsert(ctx, h.batch); err != nil {
		return err
	}

	h.logger.Infof(ctx, "Successfully inserted %d audit events", len(h.batch))

	// Clear batch
	h.batch = h.batch[:0]
	h.lastFlush = time.Now()

	return nil
}
