package model

import (
	"smap-api/internal/sqlboiler"
	"time"

	"github.com/aarondl/null/v8"
)

// User role constants
const (
	RoleAdmin   = "ADMIN"
	RoleAnalyst = "ANALYST"
	RoleViewer  = "VIEWER"
)

// User represents a user entity in the domain layer.
// This is a safe type model that doesn't depend on database-specific types.
// Users are created automatically on first OAuth2 login.
type User struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Name        *string    `json:"name,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	RoleHash    *string    `json:"-"` // Encrypted role stored in database
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewUser creates a new User with default values
func NewUser(email, name string) *User {
	user := &User{
		Email:    email,
		Name:     &name,
		IsActive: true,
	}
	// Set default role as VIEWER
	_ = user.SetRole(RoleViewer)
	return user
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// NewUserFromDB converts a SQLBoiler User to domain User
func NewUserFromDB(dbUser *sqlboiler.User) *User {
	if dbUser == nil {
		return nil
	}

	user := &User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		RoleHash:  &dbUser.RoleHash,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}

	// Handle nullable fields
	if dbUser.Name.Valid {
		user.Name = &dbUser.Name.String
	}
	if dbUser.AvatarURL.Valid {
		user.AvatarURL = &dbUser.AvatarURL.String
	}
	if dbUser.IsActive.Valid {
		user.IsActive = dbUser.IsActive.Bool
	} else {
		user.IsActive = true // Default to true if not set
	}
	if dbUser.LastLoginAt.Valid {
		user.LastLoginAt = &dbUser.LastLoginAt.Time
	}

	return user
}

// ToDBUser converts domain User to SQLBoiler User
func (u *User) ToDBUser() *sqlboiler.User {
	dbUser := &sqlboiler.User{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	// Handle RoleHash
	if u.RoleHash != nil {
		dbUser.RoleHash = *u.RoleHash
	}

	// Handle nullable fields
	if u.Name != nil {
		dbUser.Name = null.StringFrom(*u.Name)
	}
	if u.AvatarURL != nil {
		dbUser.AvatarURL = null.StringFrom(*u.AvatarURL)
	}
	dbUser.IsActive = null.BoolFrom(u.IsActive)
	if u.LastLoginAt != nil {
		dbUser.LastLoginAt = null.TimeFrom(*u.LastLoginAt)
	}

	return dbUser
}
