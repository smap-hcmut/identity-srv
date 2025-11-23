package consumer

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
	"gitlab.com/gma-vietnam/tanca-connect/internal/models"
)

func (c Consumer) createSystemEventWorker(d amqp.Delivery) {
	ctx := context.Background()
	c.l.Info(ctx, "event.delivery.rabbitmq.consumer.createEventWorker")

	var msg rabbitmq.CreateEventMsg
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		c.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.Unmarshal: %v", err)
		d.Ack(false)
		return
	}

	sc := models.Scope{
		ShopID: msg.ShopID,
	}

	err = c.uc.CreateSystemEvent(ctx, sc, event.CreateSystemEventInput{
		Title:             msg.Title,
		AssignIDs:         msg.AssignIDs,
		DepartmentIDs:     msg.DepartmentIDs,
		TimezoneID:        msg.TimezoneID,
		StartTime:         msg.StartTime,
		EndTime:           msg.EndTime,
		AllDay:            msg.AllDay,
		Repeat:            msg.Repeat,
		CategoryID:        msg.CategoryID,
		Notify:            false,
		System:            true,
		ObjectID:          msg.ObjectID,
		NeedParseTimezone: msg.NeedParseTimezone,
	})
	if err != nil {
		c.l.Errorf(ctx, "event.delivery.rabbitmq.consumer.Create: %v", err)
		d.Ack(false)
		return
	}

	d.Ack(false)
}

func (c Consumer) deleteSystemEventWorker(d amqp.Delivery) {
	ctx := context.Background()
	c.l.Info(ctx, "event.delivery.rabbitmq.consumer.deleteSystemEventWorker")

	var msg rabbitmq.DeleteEventMsg
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		c.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.Unmarshal: %v", err)
		d.Ack(false)
		return
	}

	sc := models.Scope{
		ShopID: msg.ShopID,
	}

	err = c.uc.Delete(ctx, sc, event.DeleteInput{
		ID:      msg.EventID,
		EventID: msg.EventID,
	})
	if err != nil {
		c.l.Errorf(ctx, "event.delivery.rabbitmq.consumer.Delete: %v", err)
		d.Ack(false)
		return
	}

	d.Ack(false)
}

func (c Consumer) updateSystemEventWorker(d amqp.Delivery) {
	ctx := context.Background()
	c.l.Info(ctx, "event.delivery.rabbitmq.consumer.updateSystemEventWorker")

	var msg rabbitmq.UpdateSystemEventMsg
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		c.l.Warnf(ctx, "event.delivery.rabbitmq.consumer.Unmarshal: %v", err)
		d.Ack(false)
		return
	}

	sc := models.Scope{
		ShopID: msg.ShopID,
	}

	err = c.uc.UpdateSystemEvent(ctx, sc, event.UpdateSystemEventInput{
		EventID:           msg.EventID,
		Title:             msg.Title,
		AssignIDs:         msg.AssignIDs,
		DepartmentIDs:     msg.DepartmentIDs,
		TimezoneID:        msg.TimezoneID,
		StartTime:         msg.StartTime,
		EndTime:           msg.EndTime,
		AllDay:            msg.AllDay,
		Repeat:            msg.Repeat,
		CategoryID:        msg.CategoryID,
		ObjectID:          msg.ObjectID,
		NeedParseTimezone: msg.NeedParseTimezone,
	})
	if err != nil {
		c.l.Errorf(ctx, "event.delivery.rabbitmq.consumer.Update: %v", err)
		d.Ack(false)
		return
	}

	d.Ack(false)
}
