package usecase

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/auth"
	"github.com/nguyentantai21042004/smap-api/internal/models"
)

func (u implUsecase) PubSendEmailMsg(ctx context.Context, sc models.Scope, ip auth.PubSendEmailMsgInput) error {
	err := u.prod.PubSendEmail(ctx, u.toSendEmailMsg(ip))
	if err != nil {
		u.l.Error(ctx, "auth.usecase.producer.PubSendEmailMsg: %v", err)
		return err
	}
	return nil
}
