package http

import (
	"time"

	"smap-api/internal/model"
	"smap-api/internal/user/repository"
	"smap-api/pkg/errors"
	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// validateTokenReq is the request body for POST /internal/validate
type validateTokenReq struct {
	Token string `json:"token" binding:"required"`
}

// validateTokenResp is the response body for POST /internal/validate
type validateTokenResp struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Groups    []string  `json:"groups,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// revokeTokenReq is the request body for POST /internal/revoke-token
type revokeTokenReq struct {
	JTI    string `json:"jti,omitempty"`     // Revoke specific token
	UserID string `json:"user_id,omitempty"` // Revoke all user tokens
}

// getUserResp is the response body for GET /internal/users/:id
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

// @Summary Validate Token (Internal)
// @Description Fallback token validation endpoint for services. Requires X-Service-Key header.
// @Tags Internal
// @Accept json
// @Produce json
// @Param X-Service-Key header string true "Service authentication key (encrypted)"
// @Param validateReq body validateTokenReq true "Token to validate"
// @Success 200 {object} response.Resp{data=validateTokenResp} "Token validation result"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized - Invalid or missing X-Service-Key"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /internal/validate [POST]
func (h handler) ValidateToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req validateTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "authentication.http.ValidateToken.ShouldBindJSON: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Verify JWT token
	claims, err := h.jwtManager.VerifyToken(req.Token)
	if err != nil {
		// Token is invalid
		response.OK(c, validateTokenResp{
			Valid: false,
		})
		return
	}

	// Check if token is blacklisted
	isBlacklisted, err := h.blacklistManager.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.ValidateToken.IsBlacklisted: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	if isBlacklisted {
		// Token is blacklisted
		response.OK(c, validateTokenResp{
			Valid: false,
		})
		return
	}

	// Token is valid
	response.OK(c, validateTokenResp{
		Valid:     true,
		UserID:    claims.Subject,
		Email:     claims.Email,
		Role:      claims.Role,
		Groups:    claims.Groups,
		ExpiresAt: claims.ExpiresAt.Time,
	})
}

// @Summary Revoke Token (Internal)
// @Description Revoke specific token or all user tokens. Requires X-Service-Key header and ADMIN role.
// @Tags Internal
// @Accept json
// @Produce json
// @Param X-Service-Key header string true "Service authentication key (encrypted)"
// @Param revokeReq body revokeTokenReq true "Token revocation request"
// @Success 200 {object} response.Resp "Token(s) revoked successfully"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized - Invalid or missing X-Service-Key"
// @Failure 403 {object} response.Resp "Forbidden - Requires ADMIN role"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /internal/revoke-token [POST]
func (h handler) RevokeToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req revokeTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.l.Errorf(ctx, "authentication.http.RevokeToken.ShouldBindJSON: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Validate request - must provide either JTI or UserID
	if req.JTI == "" && req.UserID == "" {
		response.Error(c, errors.NewHTTPError(400, "Must provide either jti or user_id"), h.discord)
		return
	}

	if req.JTI != "" && req.UserID != "" {
		response.Error(c, errors.NewHTTPError(400, "Cannot provide both jti and user_id"), h.discord)
		return
	}

	// Revoke specific token by JTI
	if req.JTI != "" {
		// Get session to find expiration time
		session, err := h.sessionManager.GetSession(ctx, req.JTI)
		if err != nil {
			h.l.Errorf(ctx, "authentication.http.RevokeToken.GetSession: %v", err)
			response.Error(c, err, h.discord)
			return
		}

		// Add to blacklist
		if err := h.blacklistManager.AddToken(ctx, req.JTI, session.ExpiresAt); err != nil {
			h.l.Errorf(ctx, "authentication.http.RevokeToken.AddToken: %v", err)
			response.Error(c, err, h.discord)
			return
		}

		// Delete session
		if err := h.sessionManager.DeleteSession(ctx, req.JTI); err != nil {
			h.l.Errorf(ctx, "authentication.http.RevokeToken.DeleteSession: %v", err)
			response.Error(c, err, h.discord)
			return
		}

		response.OK(c, gin.H{"message": "Token revoked successfully"})
		return
	}

	// Revoke all user tokens by UserID
	if req.UserID != "" {
		// Get all user sessions
		jtis, err := h.sessionManager.GetAllUserSessions(ctx, req.UserID)
		if err != nil {
			h.l.Errorf(ctx, "authentication.http.RevokeToken.GetAllUserSessions: %v", err)
			response.Error(c, err, h.discord)
			return
		}

		// Add all JTIs to blacklist
		// Use a reasonable expiration time (e.g., max JWT TTL = 8 hours from now)
		expiresAt := time.Now().Add(8 * time.Hour)
		if err := h.blacklistManager.AddAllUserTokens(ctx, jtis, expiresAt); err != nil {
			h.l.Errorf(ctx, "authentication.http.RevokeToken.AddAllUserTokens: %v", err)
			response.Error(c, err, h.discord)
			return
		}

		// Delete all user sessions
		if err := h.sessionManager.DeleteUserSessions(ctx, req.UserID); err != nil {
			h.l.Errorf(ctx, "authentication.http.RevokeToken.DeleteUserSessions: %v", err)
			response.Error(c, err, h.discord)
			return
		}

		response.OK(c, gin.H{"message": "All user tokens revoked successfully"})
		return
	}

	response.Error(c, errors.NewHTTPError(400, "Invalid request"), h.discord)
}

// @Summary Get User by ID (Internal)
// @Description Get user information by ID. Requires X-Service-Key header.
// @Tags Internal
// @Accept json
// @Produce json
// @Param X-Service-Key header string true "Service authentication key (encrypted)"
// @Param id path string true "User ID"
// @Success 200 {object} response.Resp{data=getUserResp} "User information"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized - Invalid or missing X-Service-Key"
// @Failure 404 {object} response.Resp "User not found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /internal/users/{id} [GET]
func (h handler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.NewHTTPError(400, "User ID is required"), h.discord)
		return
	}

	// Get user from database using GetOne with empty scope (internal service call)
	user, err := h.userRepo.GetOne(ctx, model.Scope{}, repository.GetOneOptions{
		ID: userID,
	})
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.GetUserByID.GetOne: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Helper function to safely dereference string pointers
	getName := func(s *string) string {
		if s != nil {
			return *s
		}
		return ""
	}

	response.OK(c, getUserResp{
		ID:        user.ID,
		Email:     user.Email,
		Name:      getName(user.Name),
		AvatarURL: getName(user.AvatarURL),
		Role:      user.GetRole(),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
