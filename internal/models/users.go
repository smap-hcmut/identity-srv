package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"         json:"id"`
	Email        string             `bson:"email"                 json:"email"`
	PasswordHash string             `bson:"password_hash,omitempty" json:"password_hash,omitempty"`
	FullName     string             `bson:"full_name,omitempty"   json:"full_name,omitempty"`
	NameCode     string             `bson:"name_code,omitempty"   json:"name_code,omitempty"`
	IsVerified   bool               `bson:"is_verified,omitempty" json:"is_verified,omitempty"`
	AvatarURL    string             `bson:"avatar_url,omitempty"  json:"avatar_url,omitempty"`
	Provider     string             `bson:"provider,omitempty"    json:"provider,omitempty"`
	ProviderID   string             `bson:"provider_id,omitempty" json:"provider_id,omitempty"`
	OTP          string             `bson:"otp,omitempty"         json:"otp,omitempty"`
	OTPExpiredAt time.Time          `bson:"otp_expired_at,omitempty" json:"otp_expired_at,omitempty"`
	RoleID       primitive.ObjectID `bson:"role_id"               json:"role_id"`
	CreatedAt    time.Time          `bson:"created_at,omitempty"  json:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty"  json:"updated_at,omitempty"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty"  json:"deleted_at,omitempty"`
}
