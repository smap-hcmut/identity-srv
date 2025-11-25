package consumer

import (
	"github.com/nguyentantai21042004/smap-api/internal/dispatcher/delivery/rabbitmq"
)

// Consume start consume inbound queue.
func (c Consumer) Consume() {
	go c.consume(rabbitmq.InboundExchange, rabbitmq.InboundQueueName, rabbitmq.InboundRoutingPattern, c.dispatchWorker)
}
