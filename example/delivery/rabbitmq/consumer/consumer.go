package consumer

import (
	"log"

	"gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
)

func (c Consumer) Consume() {
	go c.consume(rabbitmq.CreateSystemEventExchange, rabbitmq.CreateSystemEventQueueName, c.createSystemEventWorker)
	go c.consume(rabbitmq.DeleteSystemEventExchange, rabbitmq.DeleteSystemEventQueueName, c.deleteSystemEventWorker)
	go c.consume(rabbitmq.UpdateSystemEventExchange, rabbitmq.UpdateSystemEventQueueName, c.updateSystemEventWorker)
}

func catchPanic() {
	if r := recover(); r != nil {
		log.Printf("Recovered from panic in goroutine: %v", r)
	}
}
