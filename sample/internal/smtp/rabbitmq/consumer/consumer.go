package consumer

import (
	"context"
	"encoding/json"

	"smap-api/internal/smtp"
	"smap-api/internal/smtp/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (c Consumer) Consume() {
	go c.consume(rabbitmq.SendEmailExc, rabbitmq.SEND_EMAIL_QUEUE_NAME, c.sendEmailWorker)
}

func (c Consumer) sendEmailWorker(d amqp.Delivery) {
	ctx := context.Background()
	c.l.Info(ctx, "smtp.delivery.rabbitmq.consumer.sendEmailWorker: %v", d)

	var email rabbitmq.EmailData
	err := json.Unmarshal(d.Body, &email)
	if err != nil {
		c.l.Errorf(ctx, "smtp.delivery.rabbitmq.consumer.sendEmailWorker.json.Unmarshal: %v", err)
		d.Ack(false)
		return
	}

	attachments := make([]smtp.Attachment, 0)
	for _, attachment := range email.Attachments {
		attachments = append(attachments, smtp.Attachment{
			Filename:    attachment.Filename,
			ContentType: attachment.ContentType,
			Data:        attachment.Data,
		})
	}

	err = c.uc.SendEmail(ctx, smtp.EmailData{
		Subject:     email.Subject,
		Recipient:   email.Recipient,
		Body:        email.Body,
		ReplyTo:     email.ReplyTo,
		CcAddresses: email.CcAddresses,
		Attachments: attachments,
	})
	if err != nil {
		c.l.Errorf(ctx, "smtp.delivery.rabbitmq.consumer.sendEmailWorker.uc.SendEmail: %v", err)
		return
	}

	d.Ack(false)
}
