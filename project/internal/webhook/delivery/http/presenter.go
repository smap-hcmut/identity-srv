package http

import (
	"smap-project/internal/webhook"
)

// CallbackReq represents the HTTP request for webhook callback
type CallbackReq struct {
	JobID    string                  `json:"job_id" binding:"required"`
	Status   string                  `json:"status" binding:"required,oneof=success failed"`
	Platform string                  `json:"platform" binding:"required,oneof=youtube tiktok"`
	Payload  webhook.CallbackPayload `json:"payload"`
	// UserID is optional for backward compatibility during migration.
	// The project service will look up UserID from JobID using Redis.
	UserID string `json:"user_id,omitempty"`
}

func (r CallbackReq) toInput() webhook.CallbackRequest {
	return webhook.CallbackRequest{
		JobID:    r.JobID,
		Status:   r.Status,
		Platform: r.Platform,
		Payload:  r.Payload,
		UserID:   r.UserID,
	}
}

// CallbackResp represents the HTTP response for webhook callback
