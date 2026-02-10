package http

import (
	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ValidateToken validates a JWT token (internal service endpoint)
// @Summary Validate Token (Internal)
// @Description Fallback token validation endpoint for services. Requires X-Service-Key header.
// @Tags Internal
// @Accept json
// @Produce json
// @Param X-Service-Key header string true "Service authentication key"
// @Param body body validateTokenReq true "Token to validate"
// @Success 200 {object} response.Resp{data=validateTokenResp} "Token validation result"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /internal/validate [POST]
func (h handler) ValidateToken(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request
	token, err := h.processValidateTokenRequest(c)
	if err != nil {
		response.Error(c, err, h.discord)
		return
	}

	// 2. Call UseCase
	result, err := h.uc.ValidateToken(ctx, token)
	if err != nil {
		h.l.Errorf(ctx, "uc.ValidateToken: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	response.OK(c, h.newValidateTokenResp(result))
}

// RevokeToken revokes a specific token or all user tokens (internal service endpoint)
// @Summary Revoke Token (Internal)
// @Description Revoke specific token or all user tokens. Requires X-Service-Key + ADMIN role.
// @Tags Internal
// @Accept json
// @Produce json
// @Param X-Service-Key header string true "Service authentication key"
// @Param body body revokeTokenReq true "Token revocation request"
// @Success 200 {object} response.Resp "Token(s) revoked"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 403 {object} response.Resp "Forbidden"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /internal/revoke-token [POST]
func (h handler) RevokeToken(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request
	req, err := h.processRevokeTokenRequest(c)
	if err != nil {
		response.Error(c, err, h.discord)
		return
	}

	// 2. Call UseCase
	if req.JTI != "" {
		err = h.uc.RevokeToken(ctx, req.JTI)
	} else {
		err = h.uc.RevokeAllUserTokens(ctx, req.UserID)
	}
	if err != nil {
		h.l.Errorf(ctx, "uc.RevokeToken: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	response.OK(c, gin.H{"message": revokeResultMsg(req)})
}

// GetUserByID gets user information by ID (internal service endpoint)
// @Summary Get User by ID (Internal)
// @Description Get user information by ID. Requires X-Service-Key header.
// @Tags Internal
// @Accept json
// @Produce json
// @Param X-Service-Key header string true "Service authentication key"
// @Param id path string true "User ID"
// @Success 200 {object} response.Resp{data=getUserResp} "User information"
// @Failure 400 {object} response.Resp "Bad Request"
// @Failure 401 {object} response.Resp "Unauthorized"
// @Failure 404 {object} response.Resp "Not Found"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /internal/users/{id} [GET]
func (h handler) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request
	userID, err := h.processGetUserRequest(c)
	if err != nil {
		response.Error(c, err, h.discord)
		return
	}

	// 2. Call UseCase
	user, err := h.uc.GetUserByID(ctx, userID)
	if err != nil {
		h.l.Errorf(ctx, "uc.GetUserByID: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	response.OK(c, h.newGetUserResp(user))
}
