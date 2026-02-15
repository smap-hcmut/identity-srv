package http

import (
	"net/http"

	"identity-srv/pkg/response"

	"github.com/gin-gonic/gin"
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

	// 1. Process Request
	input := h.processLoginRequest(c)

	// 2. Call UseCase
	output, err := h.uc.InitiateOAuthLogin(ctx, input)
	if err != nil {
		h.l.Errorf(ctx, "uc.InitiateOAuthLogin: %v", err)
		response.Error(c, h.mapError(err), h.discord)
		return
	}

	// 3. Response
	h.setStateCookie(c, output.State)
	if input.RedirectURL != "" {
		h.setRedirectCookie(c, input.RedirectURL)
	}
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

	// 1. Process Request (includes CSRF state validation)
	input, err := h.processCallbackRequest(c)
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

	// Production mode: Set HttpOnly cookie and redirect
	h.setAuthCookie(c, output.Token)

	redirectURL, cookieErr := c.Cookie("oauth_redirect")
	if cookieErr != nil || redirectURL == "" {
		redirectURL = "/dashboard"
	}
	h.clearRedirectCookie(c)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
