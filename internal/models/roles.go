package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	Name        string             `bson:"name"                  json:"name"`
	Code        string             `bson:"code"                  json:"code"`
	Alias       string             `bson:"alias"                 json:"alias"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time          `bson:"created_at,omitempty"  json:"created_at,omitempty"`
	UpdatedAt   time.Time          `bson:"updated_at,omitempty"  json:"updated_at,omitempty"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty"  json:"deleted_at,omitempty"`
}
