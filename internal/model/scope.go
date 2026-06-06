package model

const (
	ScopeTypeAccess = "access"
	SMAPAPI         = "identity-srv"
)

type Scope struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"` // ADMIN, ANALYST, or VIEWER
	JTI      string `json:"jti"`
}
