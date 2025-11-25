package consumer

import (
	"smap-collector/internal/dispatcher/delivery/rabbitmq"
)

// Consume start consume inbound queue.
func (c Consumer) Consume() {
	go c.consume(rabbitmq.InboundExchange, rabbitmq.InboundQueueName, rabbitmq.InboundRoutingPattern, c.dispatchWorker)
}
