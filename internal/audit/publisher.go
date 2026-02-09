package audit

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// Publisher publishes audit events to Kafka
type Publisher interface {
	Publish(ctx context.Context, event AuditEvent) error
	Close() error
}

type publisher struct {
	producer  sarama.AsyncProducer
	topic     string
	buffer    []AuditEvent
	bufferMu  sync.Mutex
	maxBuffer int
	logger    Logger
}

// Logger interface for audit publisher
type Logger interface {
	Infof(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
}

// NewPublisher creates a new audit event publisher
func NewPublisher(producer sarama.AsyncProducer, topic string, logger Logger) Publisher {
	p := &publisher{
		producer:  producer,
		topic:     topic,
		buffer:    make([]AuditEvent, 0, 1000),
		maxBuffer: 1000,
		logger:    logger,
	}

	// Start goroutine to handle producer errors and successes
	go p.handleProducerMessages()

	return p
}

// Publish publishes an audit event to Kafka (non-blocking)
func (p *publisher) Publish(ctx context.Context, event AuditEvent) error {
	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Errorf(ctx, "Failed to marshal audit event: %v", err)
		return err
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(event.UserID), // Partition by user_id
	}

	// Try to send to Kafka (non-blocking)
	select {
	case p.producer.Input() <- msg:
		// Successfully queued
		return nil
	default:
		// Kafka producer queue is full, add to in-memory buffer
		p.bufferEvent(ctx, event)
		return nil
	}
}

// bufferEvent adds event to in-memory buffer when Kafka is unavailable
func (p *publisher) bufferEvent(ctx context.Context, event AuditEvent) {
	p.bufferMu.Lock()
	defer p.bufferMu.Unlock()

	// Check buffer size
	if len(p.buffer) >= p.maxBuffer {
		// Buffer is full, drop oldest event
		p.logger.Warnf(ctx, "Audit buffer full, dropping oldest event")
		p.buffer = p.buffer[1:]
	}

	p.buffer = append(p.buffer, event)
	p.logger.Infof(ctx, "Buffered audit event (buffer size: %d)", len(p.buffer))
}

// handleProducerMessages handles Kafka producer success and error messages
func (p *publisher) handleProducerMessages() {
	for {
		select {
		case success := <-p.producer.Successes():
			if success != nil {
				// Successfully published to Kafka
				// Try to flush buffer if there are buffered events
				p.flushBuffer()
			}

		case err := <-p.producer.Errors():
			if err != nil {
				p.logger.Errorf(context.Background(), "Failed to publish audit event to Kafka: %v", err.Err)
			}
		}
	}
}

// flushBuffer attempts to send buffered events to Kafka
func (p *publisher) flushBuffer() {
	p.bufferMu.Lock()
	defer p.bufferMu.Unlock()

	if len(p.buffer) == 0 {
		return
	}

	// Try to send buffered events
	sent := 0
	for i, event := range p.buffer {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: p.topic,
			Value: sarama.ByteEncoder(data),
			Key:   sarama.StringEncoder(event.UserID),
		}

		select {
		case p.producer.Input() <- msg:
			sent = i + 1
		default:
			// Producer queue full, stop flushing
			goto done
		}
	}

done:
	if sent > 0 {
		p.buffer = p.buffer[sent:]
		p.logger.Infof(context.Background(), "Flushed %d buffered audit events (remaining: %d)", sent, len(p.buffer))
	}
}

// Close closes the publisher
func (p *publisher) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
