package authentication

import (
	"smap-api/internal/model"
	"time"
)

// TokenValidationResult contains the result of token validation
type TokenValidationResult struct {
	Valid     bool
	UserID    string
	Email     string
	Role      string
	Groups    []string
	ExpiresAt time.Time
}

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

// ProcessOAuthCallback Input/Output

// OAuthCallbackInput contains the data extracted from the HTTP request by the handler
type OAuthCallbackInput struct {
	Code       string // Authorization code from OAuth provider
	RememberMe bool   // Whether to create a long-lived session
	IPAddress  string // Client IP address (for audit)
	UserAgent  string // Client user agent (for audit)
}

// OAuthCallbackOutput contains the result of the OAuth callback processing
type OAuthCallbackOutput struct {
	Token string // JWT token to set as cookie
}

// OAuthLoginInput contains the data for initiating OAuth login
type OAuthLoginInput struct {
	RedirectURL string // URL to redirect to after login
}

// OAuthLoginOutput contains the result of initiating OAuth login
type OAuthLoginOutput struct {
	AuthURL string // URL to redirect user to OAuth provider
	State   string // CSRF state token to store in cookie
}
