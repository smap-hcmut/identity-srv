package authentication

import (
	"context"

	"smap-api/internal/model"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Producer
	Register(ctx context.Context, sc model.Scope, ip RegisterInput) (RegisterOutput, error)
	SendOTP(ctx context.Context, sc model.Scope, ip SendOTPInput) error
	VerifyOTP(ctx context.Context, sc model.Scope, ip VerifyOTPInput) error
	Login(ctx context.Context, sc model.Scope, ip LoginInput) (LoginOutput, error)
}

type Producer interface {
	PublishSendEmail(ctx context.Context, sc model.Scope, ip PublishSendEmailMsgInput) error
}
