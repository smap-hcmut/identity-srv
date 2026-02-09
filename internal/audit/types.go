package audit

import "time"

// Action represents the type of audit action
type Action string

const (
	ActionLogin        Action = "LOGIN"
	ActionLogout       Action = "LOGOUT"
	ActionLoginFailed  Action = "LOGIN_FAILED"
	ActionTokenRevoked Action = "TOKEN_REVOKED"
)

// AuditEvent represents an audit event to be published
type AuditEvent struct {
	UserID       string            `json:"user_id"`
	Action       Action            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IPAddress    string            `json:"ip_address"`
	UserAgent    string            `json:"user_agent"`
	Timestamp    time.Time         `json:"timestamp"`
}
