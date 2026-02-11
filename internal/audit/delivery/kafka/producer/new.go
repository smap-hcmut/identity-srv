package producer

import (
	"smap-api/internal/audit"
	"sync"

	"github.com/IBM/sarama"
)

type publisher struct {
	producer  sarama.AsyncProducer
	topic     string
	buffer    []audit.AuditEvent
	bufferMu  sync.Mutex
	maxBuffer int
	logger    audit.Logger
}

// NewPublisher creates a new audit event publisher
func NewPublisher(producer sarama.AsyncProducer, topic string, logger audit.Logger) audit.Publisher {
	p := &publisher{
		producer:  producer,
		topic:     topic,
		buffer:    make([]audit.AuditEvent, 0, 1000),
		maxBuffer: 1000,
		logger:    logger,
	}

	// Start goroutine to handle producer errors and successes
	go p.handleProducerMessages()

	return p
}
