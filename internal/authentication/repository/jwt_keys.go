package repository

import (
	"context"
	"database/sql"
	"smap-api/internal/model"
	"smap-api/internal/sqlboiler"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type JWTKeysRepository struct {
	db *sql.DB
}

func NewJWTKeysRepository(db *sql.DB) *JWTKeysRepository {
	return &JWTKeysRepository{db: db}
}

func (r *JWTKeysRepository) SaveKey(ctx context.Context, key *model.JWTKey) error {
	dbKey := &sqlboiler.JWTKey{
		Kid:        key.KID,
		PrivateKey: key.PrivateKey,
		PublicKey:  key.PublicKey,
		Status:     key.Status,
		CreatedAt:  key.CreatedAt,
	}

	if key.ExpiresAt != nil {
		dbKey.ExpiresAt = null.TimeFrom(*key.ExpiresAt)
	}

	if key.RetiredAt != nil {
		dbKey.RetiredAt = null.TimeFrom(*key.RetiredAt)
	}

	return dbKey.Insert(ctx, r.db, boil.Infer())
}

func (r *JWTKeysRepository) GetActiveKey(ctx context.Context) (*model.JWTKey, error) {
	dbKey, err := sqlboiler.JWTKeys(
		qm.Where("status = ?", model.KeyStatusActive),
		qm.OrderBy("created_at DESC"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return toModelJWTKey(dbKey), nil
}

func (r *JWTKeysRepository) GetActiveAndRotatingKeys(ctx context.Context) ([]*model.JWTKey, error) {
	dbKeys, err := sqlboiler.JWTKeys(
		qm.WhereIn("status IN ?", model.KeyStatusActive, model.KeyStatusRotating),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		return nil, err
	}

	keys := make([]*model.JWTKey, len(dbKeys))
	for i, dbKey := range dbKeys {
		keys[i] = toModelJWTKey(dbKey)
	}

	return keys, nil
}

func (r *JWTKeysRepository) UpdateKeyStatus(ctx context.Context, kid, status string) error {
	dbKey, err := sqlboiler.JWTKeys(qm.Where("kid = ?", kid)).One(ctx, r.db)
	if err != nil {
		return err
	}

	dbKey.Status = status
	_, err = dbKey.Update(ctx, r.db, boil.Whitelist("status"))
	return err
}

func (r *JWTKeysRepository) GetRotatingKeys(ctx context.Context) ([]*model.JWTKey, error) {
	dbKeys, err := sqlboiler.JWTKeys(
		qm.Where("status = ?", model.KeyStatusRotating),
	).All(ctx, r.db)

	if err != nil {
		return nil, err
	}

	keys := make([]*model.JWTKey, len(dbKeys))
	for i, dbKey := range dbKeys {
		keys[i] = toModelJWTKey(dbKey)
	}

	return keys, nil
}

func toModelJWTKey(dbKey *sqlboiler.JWTKey) *model.JWTKey {
	key := &model.JWTKey{
		KID:        dbKey.Kid,
		PrivateKey: dbKey.PrivateKey,
		PublicKey:  dbKey.PublicKey,
		Status:     dbKey.Status,
		CreatedAt:  dbKey.CreatedAt,
	}

	if dbKey.ExpiresAt.Valid {
		t := dbKey.ExpiresAt.Time
		key.ExpiresAt = &t
	}

	if dbKey.RetiredAt.Valid {
		t := dbKey.RetiredAt.Time
		key.RetiredAt = &t
	}

	return key
}
