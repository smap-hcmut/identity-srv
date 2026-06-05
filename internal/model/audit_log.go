package model

import (
	"time"
)

// Audit action constants
const (
	ActionLogin         = "LOGIN"
	ActionLogout        = "LOGOUT"
	ActionLoginFailed   = "LOGIN_FAILED"
	ActionTokenRevoked  = "TOKEN_REVOKED"
	ActionCreateProject = "CREATE_PROJECT"
	ActionDeleteProject = "DELETE_PROJECT"
	ActionCreateSource  = "CREATE_SOURCE"
	ActionDeleteSource  = "DELETE_SOURCE"
	ActionExportData    = "EXPORT_DATA"
)

// AuditLog represents an audit log entry in the domain layer.
// Audit logs track all security-relevant events for compliance.
type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       *string                `json:"user_id,omitempty"` // Nullable for system events
	Action       string                 `json:"action"`
	ResourceType *string                `json:"resource_type,omitempty"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"` // Auto-delete after 90 days
}
