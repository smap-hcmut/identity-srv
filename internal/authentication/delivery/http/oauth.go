package http

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/response"
)

// OAuthLogin redirects user to OAuth2 authorization page
// @Summary Login with OAuth2
// @Description Redirects to OAuth2 authorization page (Google, Azure AD, or Okta)
// @Tags Authentication
// @Produce json
// @Param redirect query string false "URL to redirect to after login"
// @Success 302 {string} string "Redirect to OAuth provider"
// @Router /authentication/login [get]
func (h handler) OAuthLogin(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request — generates HMAC-signed state embedding the redirect URL
	input, err := h.processLoginRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "processLoginRequest: %v", err)
		response.Error(c, errInternalSystem, h.discord)
		return
	}

	// 2. Call UseCase
	output, err := h.uc.InitiateOAuthLogin(ctx, input)
	if err != nil {
		h.l.Errorf(ctx, "uc.InitiateOAuthLogin: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response — redirect to OAuth provider (no state cookie needed)
	c.Redirect(http.StatusTemporaryRedirect, output.AuthURL)
}

// OAuthCallback handles OAuth2 callback from identity provider
// @Summary OAuth2 callback handler
// @Description Handles OAuth2 callback, exchanges code for token, creates user session
// @Tags Authentication
// @Produce json
// @Param code query string true "Authorization code from provider"
// @Param state query string true "State parameter for CSRF protection"
// @Success 302 {string} string "Redirect to dashboard (production mode)"
// @Success 200 {object} response.Resp{data=oauthCallbackResp} "Token response (development mode)"
// @Failure 400 {object} response.Resp "Invalid request"
// @Failure 403 {object} response.Resp "Domain not allowed or account blocked"
// @Failure 500 {object} response.Resp "Internal server error"
// @Router /authentication/callback [get]
func (h handler) OAuthCallback(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Process Request — validates HMAC state, extracts redirect URL (no cookies)
	input, redirectURL, err := h.processCallbackRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "processCallbackRequest: %v", err)
		response.Error(c, err, h.discord)
		return
	}

	// 2. Call UseCase
	output, err := h.uc.ProcessOAuthCallback(ctx, input)
	if err != nil {
		h.l.Errorf(ctx, "uc.ProcessOAuthCallback: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	// Development mode: Return token in JSON response for easier testing
	if h.isDevelopmentMode() {
		h.l.Infof(ctx, "Development mode: returning token in response body")
		response.OK(c, h.newOAuthCallbackResp(output.Token))
		return
	}

	// Production mode: Set HttpOnly cookie and redirect.
	// The redirect URL was embedded in the HMAC-signed state at login time,
	// so SameSite is set correctly based on the original request origin:
	// - localhost origin → SameSite=None (cross-site fetch from local dev)
	// - production origin → SameSite=Lax
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	h.setAuthCookieForRedirect(c, output.Token, redirectURL)

	// Also pass the token in the redirect URL so the frontend can set its
	// own cookie when it runs on a different domain (e.g., localhost dev).
	// The frontend callback page reads ?token=..., validates it server-side,
	// and sets an HttpOnly cookie on its own domain.
	if parsed, err := url.Parse(redirectURL); err == nil {
		q := parsed.Query()
		q.Set("token", output.Token)
		parsed.RawQuery = q.Encode()
		redirectURL = parsed.String()
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
