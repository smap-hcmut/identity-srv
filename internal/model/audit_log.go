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

// NewAuditLog creates a new audit log entry
func NewAuditLog(userID, action string) *AuditLog {
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 90) // 90 days retention

	return &AuditLog{
		UserID:    &userID,
		Action:    action,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
}

// WithResource adds resource information to the audit log
func (a *AuditLog) WithResource(resourceType, resourceID string) *AuditLog {
	a.ResourceType = &resourceType
	a.ResourceID = &resourceID
	return a
}

// WithMetadata adds metadata to the audit log
func (a *AuditLog) WithMetadata(metadata map[string]interface{}) *AuditLog {
	a.Metadata = metadata
	return a
}

// WithIPAddress adds IP address to the audit log
func (a *AuditLog) WithIPAddress(ipAddress string) *AuditLog {
	a.IPAddress = &ipAddress
	return a
}

// WithUserAgent adds user agent to the audit log
func (a *AuditLog) WithUserAgent(userAgent string) *AuditLog {
	a.UserAgent = &userAgent
	return a
}
