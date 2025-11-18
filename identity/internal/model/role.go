package model

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

var (
	ErrInvalidRole = errors.New("invalid role")
)

// EncryptRole encrypts a role string using SHA256 and base64 encoding
// This prevents direct string comparison and obfuscates the role value
func EncryptRole(role string) (string, error) {
	// Validate role
	if role != RoleUser && role != RoleAdmin {
		return "", ErrInvalidRole
	}

	// Hash the role with a salt
	hash := sha256.Sum256([]byte(role + ":smap:role"))
	encrypted := base64.StdEncoding.EncodeToString(hash[:])

	return encrypted, nil
}

// VerifyRole verifies if a roleHash matches the given plaintext role
func VerifyRole(roleHash, plainRole string) bool {
	// Validate plainRole
	if plainRole != RoleUser && plainRole != RoleAdmin {
		return false
	}

	// Encrypt the plainRole and compare
	encrypted, err := EncryptRole(plainRole)
	if err != nil {
		return false
	}

	return roleHash == encrypted
}

// IsAdmin checks if a user has admin role
func (u *User) IsAdmin() bool {
	if u.RoleHash == nil {
		return false
	}
	return VerifyRole(*u.RoleHash, RoleAdmin)
}

// IsUser checks if a user has user role
func (u *User) IsUser() bool {
	if u.RoleHash == nil {
		return false
	}
	return VerifyRole(*u.RoleHash, RoleUser)
}

// GetRole returns the decrypted role string
// Returns empty string if role cannot be determined
func (u *User) GetRole() string {
	if u.RoleHash == nil {
		return ""
	}

	if VerifyRole(*u.RoleHash, RoleAdmin) {
		return RoleAdmin
	}
	if VerifyRole(*u.RoleHash, RoleUser) {
		return RoleUser
	}

	return ""
}

// SetRole sets the user's role with encryption
func (u *User) SetRole(role string) error {
	encrypted, err := EncryptRole(role)
	if err != nil {
		return fmt.Errorf("failed to encrypt role: %w", err)
	}
	u.RoleHash = &encrypted
	return nil
}
