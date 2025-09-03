package mongo

import (
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/session"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildSessionModel(sc models.Scope, opt session.CreateSessionOptions) (models.Session, error) {
	now := repo.clock()
	userID, err := primitive.ObjectIDFromHex(opt.UserID)
	if err != nil {
		return models.Session{}, err
	}

	session := models.Session{
		ID:           repo.db.NewObjectID(),
		UserID:       userID,
		AccessToken:  opt.AccessToken,
		RefreshToken: opt.RefreshToken,
		UserAgent:    opt.UserAgent,
		IPAddress:    opt.IPAddress,
		DeviceName:   opt.DeviceName,
		ExpiresAt:    opt.ExpiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return session, nil
}