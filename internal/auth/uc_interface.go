package auth

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

//go:generate mockery --name UseCase
type UseCase interface {
	Producer
	Register(ctx context.Context, sc models.Scope, ip RegisterInput) (RegisterOutput, error)
	SendOTP(ctx context.Context, sc models.Scope, ip SendOTPInput) error
	VerifyOTP(ctx context.Context, sc models.Scope, ip VerifyOTPInput) error
	Login(ctx context.Context, sc models.Scope, ip LoginInput) (LoginOutput, error)
	SocialLogin(ctx context.Context, sc models.Scope, ip SocialLoginInput) (SocialLoginOutput, error)
	SocialCallback(ctx context.Context, sc models.Scope, ip SocialCallbackInput) (SocialCallbackOutput, error)
}

type Producer interface {
	PubSendEmailMsg(ctx context.Context, sc models.Scope, ip PubSendEmailMsgInput) error
}
