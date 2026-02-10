package authentication

import (
	"smap-api/internal/model"
)

// GetCurrentUser
type GetCurrentUserOutput struct {
	User model.User
}

// OAuth user operations Input structs
type CreateOrUpdateUserInput struct {
	Email     string
	Name      string
	AvatarURL string
}

type UpdateUserRoleInput struct {
	UserID string
	Role   string
}
