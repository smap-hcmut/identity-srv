package authentication

import (
	"smap-api/internal/model"
)

// Login
type LoginInput struct {
	Email      string
	Password   string
	Remember   bool
	UserAgent  string
	IPAddress  string
	DeviceName string
}

type LoginOutput struct {
	User  model.User
	Token TokenOutput
}

type TokenOutput struct {
	AccessToken string
	TokenType   string
}

// GetCurrentUser
type GetCurrentUserOutput struct {
	User model.User
}
