package mongo

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/session"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
)

const (
	sessionCollection = "sessions"
)

func (repo implRepository) getSessionCollection() mongo.Collection {
	return repo.db.Collection(sessionCollection)
}

func (repo implRepository) Create(ctx context.Context, sc models.Scope, opt session.CreateSessionOptions) (models.Session, error) {
	col := repo.getSessionCollection()

	session, err := repo.buildSessionModel(sc, opt)
	if err != nil {
		repo.l.Errorf(ctx, "session.mongo.Create.buildSessionModel: %v", err)
		return models.Session{}, err
	}

	_, err = col.InsertOne(ctx, session)
	if err != nil {
		repo.l.Errorf(ctx, "session.mongo.Create.InsertOne: %v", err)
		return models.Session{}, err
	}

	return session, nil
}