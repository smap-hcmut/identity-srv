package usecase

import (
	"context"

	"smap-api/internal/authentication"
	"smap-api/internal/model"
)

func (u implUsecase) PublishSendEmail(ctx context.Context, sc model.Scope, ip authentication.PublishSendEmailMsgInput) error {
	err := u.prod.PublishSendEmail(ctx, u.toPublishSendEmailMsg(ip))
	if err != nil {
		u.l.Error(ctx, "authentication.usecase.producer.PublishSendEmail: %v", err)
		return err
	}
	return nil
}
