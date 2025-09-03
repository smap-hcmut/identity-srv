package usecase

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/session"
	"go.mongodb.org/mongo-driver/mongo"
)

func (uc implUsecase) Create(ctx context.Context, sc models.Scope, ip session.CreateSessionInput) (session.CreateSessionOutput, error) {
	u, err := uc.userUC.Detail(ctx, sc, ip.UserID)
	if err != nil {
		uc.l.Errorf(ctx, "session.usecase.Create.userRepo.Detail: %v", err)
		return session.CreateSessionOutput{}, err
	}

	if !u.User.IsVerified {
		uc.l.Warnf(ctx, "session.usecase.Create.userRepo.Detail: %v", err)
		return session.CreateSessionOutput{}, session.ErrUserNotVerified
	}

	r, err := uc.repo.Create(ctx, sc, session.CreateSessionOptions{
		UserID:       ip.UserID,
		AccessToken:  ip.AccessToken,
		RefreshToken: ip.RefreshToken,
		UserAgent:    ip.UserAgent,
		IPAddress:    ip.IPAddress,
		DeviceName:   ip.DeviceName,
		ExpiresAt:    ip.ExpiresAt,
	})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "session.usecase.Create.repo.Create: %v", err)
			return session.CreateSessionOutput{}, session.ErrSessionNotFound
		}
		uc.l.Errorf(ctx, "session.usecase.Create.repo.Create: %v", err)
		return session.CreateSessionOutput{}, err
	}

	return session.CreateSessionOutput{
		Session: r,
	}, nil
}
