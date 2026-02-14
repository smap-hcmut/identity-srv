package postgre

import (
	"context"
	"database/sql"
	"smap-api/internal/model"
	"smap-api/internal/sqlboiler"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func (r *implRepository) SaveKey(ctx context.Context, key *model.JWTKey) error {
	dbKey := r.toDBJWTKey(key)
	return dbKey.Insert(ctx, r.db, boil.Infer())
}

func (r *implRepository) GetActiveKey(ctx context.Context) (*model.JWTKey, error) {
	mods := r.buildGetActiveKeyQuery(ctx)
	dbKey, err := sqlboiler.JWTKeys(mods...).One(ctx, r.db)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.toModelJWTKey(dbKey), nil
}

func (r *implRepository) GetActiveAndRotatingKeys(ctx context.Context) ([]*model.JWTKey, error) {
	mods := r.buildGetActiveAndRotatingKeysQuery(ctx)
	dbKeys, err := sqlboiler.JWTKeys(mods...).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	keys := make([]*model.JWTKey, len(dbKeys))
	for i, dbKey := range dbKeys {
		keys[i] = r.toModelJWTKey(dbKey)
	}
	return keys, nil
}

func (r *implRepository) UpdateKeyStatus(ctx context.Context, kid, status string) error {
	mods := r.buildUpdateKeyStatusQuery(ctx, kid)
	dbKey, err := sqlboiler.JWTKeys(mods...).One(ctx, r.db)
	if err != nil {
		return err
	}

	dbKey.Status = status
	_, err = dbKey.Update(ctx, r.db, boil.Whitelist("status"))
	return err
}

func (r *implRepository) GetRotatingKeys(ctx context.Context) ([]*model.JWTKey, error) {
	mods := r.buildGetRotatingKeysQuery(ctx)
	dbKeys, err := sqlboiler.JWTKeys(mods...).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	keys := make([]*model.JWTKey, len(dbKeys))
	for i, dbKey := range dbKeys {
		keys[i] = r.toModelJWTKey(dbKey)
	}
	return keys, nil
}
