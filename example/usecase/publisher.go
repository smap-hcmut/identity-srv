package usecase

import (
	"context"

	rabb "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq"
	"gitlab.com/gma-vietnam/tanca-connect/internal/resources/notification"
)

func (uc implUseCase) publishPushNotiMsg(ctx context.Context, n notification.Notification) error {
	msg := rabb.PushNotiMsg{
		ShopScope: rabb.ShopScope{
			ID:     n.ShopScope.ID,
			Suffix: n.ShopScope.Suffix,
		},
		Content:       n.Content,
		Heading:       n.Heading,
		UserIDs:       n.UserIDs,
		CreatedUserID: n.CreatedUserID,
		En: rabb.MultiLangObj{
			Heading: n.En.Heading,
			Content: n.En.Content,
		},
		Ja: rabb.MultiLangObj{
			Heading: n.Ja.Heading,
			Content: n.Ja.Content,
		},
		Data: rabb.NotiData{
			Data:     n.Data.Data,
			Activity: n.Data.Activity,
		},
		Source: n.Source,
	}

	err := uc.producer.PublishPushNotiMsg(ctx, msg)
	if err != nil {
		uc.l.Errorf(ctx, "event.usecase.publishPushNotiMsg: %v", err)
		return err
	}

	return nil
}
