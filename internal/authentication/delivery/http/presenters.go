package http

import (
	"fmt"
	"smap-api/internal/authentication"
	"smap-api/internal/model"
	"time"
)

// --- Request DTOs ---

type validateTokenReq struct {
	Token string `json:"token" binding:"required"`
}

type revokeTokenReq struct {
	JTI    string `json:"jti,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// --- Response DTOs ---

type getMeResp struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName *string `json:"full_name,omitempty"`
}

type validateTokenResp struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Groups    []string  `json:"groups,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type getUserResp struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Response Mappers ---

func (h handler) newGetMeResp(o *model.User) *getMeResp {
	return &getMeResp{
		ID:       o.ID,
		Email:    o.Email,
		FullName: o.Name,
	}
}

func (h handler) newValidateTokenResp(o *authentication.TokenValidationResult) validateTokenResp {
	if !o.Valid {
		return validateTokenResp{Valid: false}
	}
	return validateTokenResp{
		Valid:     true,
		UserID:    o.UserID,
		Email:     o.Email,
		Role:      o.Role,
		Groups:    o.Groups,
		ExpiresAt: o.ExpiresAt,
	}
}

func (h handler) newGetUserResp(o *model.User) getUserResp {
	return getUserResp{
		ID:        o.ID,
		Email:     o.Email,
		Name:      derefString(o.Name),
		AvatarURL: derefString(o.AvatarURL),
		Role:      o.GetRole(),
		IsActive:  o.IsActive,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

// --- Helpers ---

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func revokeResultMsg(req revokeTokenReq) string {
	if req.JTI != "" {
		return "Token revoked successfully"
	}
	return fmt.Sprintf("All tokens for user %s revoked successfully", req.UserID)
}
