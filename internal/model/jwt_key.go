package model

import (
	"time"
)

// JWT key status constants
const (
	KeyStatusActive   = "active"   // Currently signing new tokens
	KeyStatusRotating = "rotating" // Grace period - old tokens still valid
	KeyStatusRetired  = "retired"  // No longer used
)

// JWTKey represents a JWT signing key pair in the domain layer.
// Supports key rotation for enhanced security.
type JWTKey struct {
	KID        string     `json:"kid"` // Key ID
	PrivateKey string     `json:"private_key"`
	PublicKey  string     `json:"public_key"`
	Status     string     `json:"status"` // active, rotating, retired
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RetiredAt  *time.Time `json:"retired_at,omitempty"`
}

// NewJWTKey creates a new JWT key pair
func NewJWTKey(kid, privateKey, publicKey string) *JWTKey {
	return &JWTKey{
		KID:        kid,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Status:     KeyStatusActive,
		CreatedAt:  time.Now(),
	}
}

// IsActive checks if the key is currently active
func (k *JWTKey) IsActive() bool {
	return k.Status == KeyStatusActive
}

// IsRotating checks if the key is in rotation grace period
func (k *JWTKey) IsRotating() bool {
	return k.Status == KeyStatusRotating
}

// IsRetired checks if the key is retired
func (k *JWTKey) IsRetired() bool {
	return k.Status == KeyStatusRetired
}

// MarkRotating marks the key as rotating (grace period)
func (k *JWTKey) MarkRotating() {
	k.Status = KeyStatusRotating
}

// MarkRetired marks the key as retired
func (k *JWTKey) MarkRetired() {
	k.Status = KeyStatusRetired
	now := time.Now()
	k.RetiredAt = &now
}

// SetExpiration sets the expiration time for the key
func (k *JWTKey) SetExpiration(expiresAt time.Time) {
	k.ExpiresAt = &expiresAt
}
