package model

const (
	ScopeTypeAccess = "access"
	SMAPAPI         = "smap-api"
)

type Scope struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}
