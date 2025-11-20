package usecase

import (
	"smap-api/internal/authentication"
	"smap-api/internal/authentication/delivery/rabbitmq"
)

// Specialize data type for Rabbitmqs
func (uc implUsecase) toPublishSendEmailMsg(ip authentication.PublishSendEmailMsgInput) rabbitmq.PublishSendEmailMsg {
	atchs := make([]rabbitmq.Attachment, 0)
	for _, a := range ip.Attachments {
		atchs = append(atchs, rabbitmq.Attachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Data:        a.Data,
		})
	}
	return rabbitmq.PublishSendEmailMsg{
		Subject:     ip.Subject,
		Recipient:   ip.Recipient,
		Body:        ip.Body,
		ReplyTo:     ip.ReplyTo,
		CcAddresses: ip.CcAddresses,
		Attachments: atchs,
	}
}
