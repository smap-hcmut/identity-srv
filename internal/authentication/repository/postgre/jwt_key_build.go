package postgre

import (
	"smap-api/internal/model"
	"smap-api/internal/sqlboiler"

	"github.com/aarondl/null/v8"
)

func (r *implRepository) toModelJWTKey(dbKey *sqlboiler.JWTKey) *model.JWTKey {
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

func (r *implRepository) toDBJWTKey(key *model.JWTKey) *sqlboiler.JWTKey {
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

	return dbKey
}
