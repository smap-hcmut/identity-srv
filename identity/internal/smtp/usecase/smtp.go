package smtp

import (
	"context"
	"io"

	"smap-api/internal/smtp"

	gomail "github.com/go-mail/mail/v2"
)

func (s implService) SendEmail(ctx context.Context, data smtp.EmailData) error {
	message := gomail.NewMessage()
	cfg := s.cfg

	message.SetHeader("From", cfg.From)
	message.SetHeader("To", data.Recipient)
	message.SetHeader("Subject", data.Subject)
	message.SetBody("text/html", data.Body)
	ccAddresses := make([]string, 0)
	ccAddresses = append(ccAddresses, data.CcAddresses...)
	if len(ccAddresses) > 0 {
		message.SetHeader("Cc", ccAddresses...)
	}

	if len(data.Attachments) > 0 {
		for _, attachment := range data.Attachments {
			message.Attach(attachment.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Data)
				if err != nil {
					s.l.Errorf(ctx, "smtp.usecase.SendEmail.Attach.Write: %v", err)
					return err
				}
				return nil
			}))
		}
	}

	smtp := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	if err := smtp.DialAndSend(message); err != nil {
		s.l.Errorf(ctx, "smtp.usecase.SendEmail.DialAndSend: %v", err)
		return err
	}

	return nil
}
