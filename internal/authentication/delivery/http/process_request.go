package http

import (
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/auth"
)

// --- Scope extraction ---

func (h handler) getScope(c *gin.Context) (model.Scope, error) {
	payload, ok := auth.GetPayloadFromContext(c.Request.Context())
	if !ok {
		return model.Scope{}, authentication.ErrScopeNotFound
	}
	
	// Ensure UserID is populated from Subject if empty
	userID := payload.UserID
	if userID == "" && payload.Subject != "" {
		userID = payload.Subject
	}

	if userID == "" {
		return model.Scope{}, authentication.ErrScopeNotFound
	}

	return model.Scope{
		UserID:   userID,
		Username: payload.Username,
		Role:     payload.Role,
		JTI:      payload.Id,
	}, nil
}

// --- Process request functions ---

func (h handler) processCallbackRequest(c *gin.Context) (authentication.OAuthCallbackInput, error) {
	// Validate CSRF state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		return authentication.OAuthCallbackInput{}, errInvalidState
	}
	h.clearStateCookie(c)

	// Validate code
	code := c.Query("code")
	if code == "" {
		return authentication.OAuthCallbackInput{}, errMissingCode
	}

	return authentication.OAuthCallbackInput{
		Code:       code,
		RememberMe: c.Query("remember_me") == "true",
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}, nil
}

func (h handler) processLoginRequest(c *gin.Context) authentication.OAuthLoginInput {
	return authentication.OAuthLoginInput{
		RedirectURL: c.Query("redirect"),
	}
}

func (h handler) processValidateTokenRequest(c *gin.Context) (string, error) {
	var req validateTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", errWrongBody
	}
	return req.Token, nil
}

func (h handler) processRevokeTokenRequest(c *gin.Context) (revokeTokenReq, error) {
	var req revokeTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return revokeTokenReq{}, errWrongBody
	}

	if req.JTI == "" && req.UserID == "" {
		return revokeTokenReq{}, errMissingJTIOrUserID
	}
	if req.JTI != "" && req.UserID != "" {
		return revokeTokenReq{}, errConflictJTIAndUserID
	}

	return req, nil
}

func (h handler) processGetUserRequest(c *gin.Context) (string, error) {
	userID := c.Param("id")
	if userID == "" {
		return "", errMissingUserID
	}
	return userID, nil
}

// --- Cookie helpers (HTTP transport concern) ---

func (h handler) setAuthCookie(c *gin.Context, token string) {
	auth.GinSetAuthCookie(c, token, h.cookieConfig.Domain)
}

func (h handler) expireAuthCookie(c *gin.Context) {
	c.SetCookie(
		h.cookieConfig.Name,
		"",
		-1,
		"/",
		h.cookieConfig.Domain,
		true,
		true,
	)
}

func (h handler) setStateCookie(c *gin.Context, state string) {
	c.SetCookie("oauth_state", state, 300, "/", h.cookieConfig.Domain, true, true)
}

func (h handler) clearStateCookie(c *gin.Context) {
	c.SetCookie("oauth_state", "", -1, "/", h.cookieConfig.Domain, true, true)
}

func (h handler) setRedirectCookie(c *gin.Context, url string) {
	c.SetCookie("oauth_redirect", url, 300, "/", h.cookieConfig.Domain, true, true)
}

func (h handler) clearRedirectCookie(c *gin.Context) {
	c.SetCookie("oauth_redirect", "", -1, "/", h.cookieConfig.Domain, true, true)
}
