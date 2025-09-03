package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"       json:"id"`
	UserID    primitive.ObjectID `bson:"user_id"             json:"user_id"`
	Token     string             `bson:"token"               json:"token"` 
	ExpiresAt time.Time          `bson:"expires_at"          json:"expires_at"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
