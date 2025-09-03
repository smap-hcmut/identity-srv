package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	UserID       primitive.ObjectID `bson:"user_id"              json:"user_id"`
	AccessToken  string             `bson:"access_token"         json:"access_token"` 
	RefreshToken string             `bson:"refresh_token"        json:"refresh_token"` 
	UserAgent    string             `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	IPAddress    string             `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	DeviceName   string             `bson:"device_name,omitempty" json:"device_name,omitempty"`
	ExpiresAt    time.Time          `bson:"expires_at"           json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
