package http

import (
	"fmt"
	"time"

	"smap-api/internal/model"
)

// --- Request DTOs ---

type getAuditLogsReq struct {
	UserID string `form:"user_id"`
	Action string `form:"action"`
	From   string `form:"from"`
	To     string `form:"to"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

// --- Response DTOs ---

type auditLogResp struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IPAddress    string            `json:"ip_address"`
	UserAgent    string            `json:"user_agent"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
}

type auditLogsResp struct {
	Logs       []auditLogResp `json:"logs"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// --- Response mapper ---

func (h handler) newAuditLogsResp(logs []model.AuditLog, totalCount, page, limit int) auditLogsResp {
	items := make([]auditLogResp, len(logs))
	for i, log := range logs {
		metadata := make(map[string]string)
		for k, v := range log.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				metadata[k] = fmt.Sprint(v)
			}
		}

		items[i] = auditLogResp{
			ID:           log.ID,
			UserID:       derefString(log.UserID),
			Action:       log.Action,
			ResourceType: derefString(log.ResourceType),
			ResourceID:   derefString(log.ResourceID),
			Metadata:     metadata,
			IPAddress:    derefString(log.IPAddress),
			UserAgent:    derefString(log.UserAgent),
			CreatedAt:    log.CreatedAt,
			ExpiresAt:    log.ExpiresAt,
		}
	}

	return auditLogsResp{
		Logs:       items,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}
}

// --- Helpers ---

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
