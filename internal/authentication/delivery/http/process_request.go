package http

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"
	"net/http"
	"strings"
	"time"

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

// --- HMAC-signed OAuth state ---
//
// The OAuth "state" parameter serves as a CSRF token. Instead of storing it in
// a Set-Cookie header (which breaks when the login goes through the localhost
// proxy but the callback hits the production domain directly), we embed both the
// nonce and the post-login redirect URL in a self-contained, HMAC-signed token.
//
// Format: base64url(JSON payload) + "." + base64url(HMAC-SHA256)
// base64url uses the RawURL alphabet (no padding, no "+", no "/"), so "." is a
// safe, unambiguous separator.

type statePayload struct {
	Nonce    string `json:"n"`
	Redirect string `json:"r,omitempty"`
	Exp      int64  `json:"e"` // Unix timestamp (5-minute window)
}

// generateSignedState creates a tamper-proof state token that embeds the
// post-login redirect URL. No cookie is needed.
func (h handler) generateSignedState(redirectURL string) (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("generateSignedState: rand.Read: %w", err)
	}

	payload := statePayload{
		Nonce:    base64.RawURLEncoding.EncodeToString(nonceBytes),
		Redirect: redirectURL,
		Exp:      time.Now().Add(5 * time.Minute).Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("generateSignedState: json.Marshal: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := h.computeStateHMAC(encodedPayload)
	return encodedPayload + "." + sig, nil
}

// verifySignedState validates the HMAC signature and expiry, then returns the
// embedded payload. Returns an error if the signature is invalid or the token
// has expired.
func (h handler) verifySignedState(signedState string) (statePayload, error) {
	dotIdx := strings.LastIndex(signedState, ".")
	if dotIdx < 0 {
		return statePayload{}, fmt.Errorf("invalid state format: missing separator")
	}

	encodedPayload := signedState[:dotIdx]
	sig := signedState[dotIdx+1:]

	expectedSig := h.computeStateHMAC(encodedPayload)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return statePayload{}, fmt.Errorf("invalid state signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return statePayload{}, fmt.Errorf("invalid state encoding: %w", err)
	}

	var p statePayload
	if err := json.Unmarshal(payloadJSON, &p); err != nil {
		return statePayload{}, fmt.Errorf("invalid state payload: %w", err)
	}

	if time.Now().Unix() > p.Exp {
		return statePayload{}, fmt.Errorf("state expired")
	}

	return p, nil
}

func (h handler) computeStateHMAC(data string) string {
	mac := hmac.New(sha256.New, []byte(h.stateSecret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// --- Process request functions ---

// processLoginRequest generates a signed state (embedding the redirect URL) and
// returns the login input for the use case.
func (h handler) processLoginRequest(c *gin.Context) (authentication.OAuthLoginInput, error) {
	redirectURL := c.Query("redirect")
	signedState, err := h.generateSignedState(redirectURL)
	if err != nil {
		return authentication.OAuthLoginInput{}, err
	}
	return authentication.OAuthLoginInput{
		RedirectURL: redirectURL,
		State:       signedState,
	}, nil
}

// processCallbackRequest validates the HMAC-signed state from the query param and
// returns the callback input together with the redirect URL embedded in the state.
// No cookies are read — this works regardless of which origin the callback arrives from.
func (h handler) processCallbackRequest(c *gin.Context) (authentication.OAuthCallbackInput, string, error) {
	state := c.Query("state")
	payload, err := h.verifySignedState(state)
	if err != nil {
		return authentication.OAuthCallbackInput{}, "", errInvalidState
	}

	code := c.Query("code")
	if code == "" {
		return authentication.OAuthCallbackInput{}, "", errMissingCode
	}

	return authentication.OAuthCallbackInput{
		Code:       code,
		RememberMe: c.Query("remember_me") == "true",
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}, payload.Redirect, nil
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

// setAuthCookieForRedirect sets the auth cookie with SameSite determined by the
// redirect destination rather than the Origin header (which is absent in OAuth redirects).
func (h handler) setAuthCookieForRedirect(c *gin.Context, token string, redirectURL string) {
	isLocalhost := strings.HasPrefix(redirectURL, "http://localhost") ||
		strings.HasPrefix(redirectURL, "https://localhost")

	if isLocalhost {
		// Cross-site: frontend on localhost, API on production domain.
		// SameSite=None;Secure required for browser to send cookie on fetch requests.
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     h.cookieConfig.Name,
			Value:    token,
			Path:     "/",
			MaxAge:   h.cookieConfig.MaxAge,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})
	} else {
		auth.GinSetAuthCookie(c, token, h.cookieConfig.Domain)
	}
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
