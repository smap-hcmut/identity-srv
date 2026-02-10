package http

import (
	"slices"
	"smap-api/internal/audit"
	"smap-api/pkg/response"
	"smap-api/pkg/scope"

	"github.com/gin-gonic/gin"
)

// @Summary Logout
// @Description Logout by expiring the authentication cookie. Requires authentication via cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp "Success - Authentication cookie expired"
// @Header 200 {string} Set-Cookie "smap_auth_token=; Path=/; Domain=.tantai.dev; HttpOnly; Secure; Max-Age=-1"
// @Failure 401 {object} response.Resp "Unauthorized - Missing or invalid authentication cookie"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/logout [POST]
// @Security CookieAuth
func (h handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// Get scope from context (set by Auth middleware)
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "authentication.http.Logout.GetScopeFromContext: scope not found")
		response.Unauthorized(c)
		return
	}

	// Call logout usecase (for any cleanup logic)
	err := h.uc.Logout(ctx, sc)
	if err != nil {
		h.l.Errorf(ctx, "authentication.http.Logout.Logout: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// Publish audit log event for logout
	h.uc.PublishAuditEvent(ctx, audit.AuditEvent{
		UserID:       sc.UserID,
		Action:       audit.ActionLogout,
		ResourceType: "authentication",
		Metadata: map[string]string{
			"role": sc.Role,
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	// Expire authentication cookie by setting MaxAge to -1
	c.SetCookie(
		h.cookieConfig.Name,
		"",
		-1, // MaxAge: -1 expires the cookie immediately
		"/",
		h.cookieConfig.Domain,
		h.cookieConfig.Secure,
		true, // HttpOnly
	)

	response.OK(c, nil)
}

// @Summary Get Current User
// @Description Get current authenticated user information. Requires authentication via cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp{data=getMeResp} "Success - Returns current user information"
// @Failure 401 {object} response.Resp "Unauthorized - Missing or invalid authentication cookie"
// @Failure 500 {object} response.Resp "Internal Server Error"
// @Router /authentication/me [GET]
// @Security CookieAuth
func (h handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	// Get scope from context (set by Auth middleware)
	sc, ok := scope.GetScopeFromContext(ctx)
	if !ok {
		h.l.Errorf(ctx, "authentication.http.GetMe.GetScopeFromContext: scope not found")
		response.Unauthorized(c)
		return
	}

	// Call GetCurrentUser usecase
	o, err := h.uc.GetCurrentUser(ctx, sc)
	if err != nil {
		err = h.mapErrorCode(err)
		if !slices.Contains(NotFound, err) {
			h.l.Errorf(ctx, "authentication.http.GetMe.GetCurrentUser: %v", err)
			response.Error(c, err, h.discord)
			return
		} else {
			h.l.Warnf(ctx, "authentication.http.GetMe.GetCurrentUser: %v", err)
			response.Error(c, err, h.discord)
			return
		}
	}

	response.OK(c, h.newGetMeResp(o))
}

// addSameSiteAttribute manually adds SameSite attribute to the last Set-Cookie header
// This is a workaround since Gin's SetCookie doesn't support SameSite parameter
func (h handler) addSameSiteAttribute(c *gin.Context, sameSite string) {
	cookies := c.Writer.Header()["Set-Cookie"]
	if len(cookies) > 0 {
		// Get the last cookie (the one we just set)
		lastCookie := cookies[len(cookies)-1]
		// Add SameSite attribute
		lastCookie += "; SameSite=" + sameSite
		// Update the header
		cookies[len(cookies)-1] = lastCookie
		c.Writer.Header()["Set-Cookie"] = cookies
	}
}
