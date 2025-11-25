package consumer

import (
	"context"
	"encoding/json"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/nguyentantai21042004/smap-api/internal/dispatcher"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (c Consumer) dispatchWorker(d amqp.Delivery) {
	ctx := context.Background()
	c.l.Info(ctx, "dispatcher.delivery.rabbitmq.consumer.dispatchWorker")

	var req models.CrawlRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		c.l.Warnf(ctx, "dispatcher.consumer.Unmarshal: %v", err)
		d.Ack(false)
		return
	}

	// Derive platform/task_type từ routing key nếu thiếu.
	if req.Platform == "" || req.TaskType == "" {
		setFromRouting(&req, d.RoutingKey)
	}

	tasks, err := c.uc.Dispatch(ctx, req)
	if err != nil {
		if err == dispatcher.ErrInvalidInput || err == dispatcher.ErrUnknownRoute {
			c.l.Warnf(ctx, "dispatcher.consumer.Dispatch invalid: %v", err)
			d.Ack(false)
			return
		}
		c.l.Errorf(ctx, "dispatcher.consumer.Dispatch: %v", err)
		d.Ack(false)
		return
	}

	c.l.Infof(ctx, "dispatcher.consumer.Dispatch published %d task(s)", len(tasks))
	d.Ack(false)
}

func setFromRouting(req *models.CrawlRequest, routing string) {
	parts := strings.Split(routing, ".")
	if len(parts) >= 3 && req.Platform == "" {
		req.Platform = models.Platform(parts[1])
	}
	if len(parts) >= 3 && req.TaskType == "" {
		req.TaskType = models.TaskType(parts[2])
	}
}
