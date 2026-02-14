package http

import (
	"smap-api/internal/authentication"
	"smap-api/internal/model"
	"smap-api/pkg/scope"

	"github.com/gin-gonic/gin"
)

// --- Scope extraction ---

func (h handler) getScope(c *gin.Context) (model.Scope, error) {
	sc, ok := scope.GetScopeFromContext(c.Request.Context())
	if !ok {
		return model.Scope{}, authentication.ErrScopeNotFound
	}
	return sc, nil
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
	c.SetCookie(
		h.cookieConfig.Name,
		token,
		h.cookieConfig.MaxAge,
		"/",
		h.cookieConfig.Domain,
		h.cookieConfig.Secure,
		true,
	)
	h.addSameSiteAttribute(c, h.cookieConfig.SameSite)
}

func (h handler) expireAuthCookie(c *gin.Context) {
	c.SetCookie(
		h.cookieConfig.Name,
		"",
		-1,
		"/",
		h.cookieConfig.Domain,
		h.cookieConfig.Secure,
		true,
	)
}

func (h handler) setStateCookie(c *gin.Context, state string) {
	c.SetCookie("oauth_state", state, 300, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)
}

func (h handler) clearStateCookie(c *gin.Context) {
	c.SetCookie("oauth_state", "", -1, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)
}

func (h handler) setRedirectCookie(c *gin.Context, url string) {
	c.SetCookie("oauth_redirect", url, 300, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)
}

func (h handler) clearRedirectCookie(c *gin.Context) {
	c.SetCookie("oauth_redirect", "", -1, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)
}

// addSameSiteAttribute manually adds SameSite attribute to the last Set-Cookie header
func (h handler) addSameSiteAttribute(c *gin.Context, sameSite string) {
	if sameSite == "" {
		sameSite = "Lax" // Default
	}
	cookies := c.Writer.Header()["Set-Cookie"]
	if len(cookies) > 0 {
		lastCookie := cookies[len(cookies)-1]
		lastCookie += "; SameSite=" + sameSite
		cookies[len(cookies)-1] = lastCookie
		c.Writer.Header()["Set-Cookie"] = cookies
	}
}
