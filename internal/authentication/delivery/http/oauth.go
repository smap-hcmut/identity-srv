package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"smap-api/internal/audit"
	"smap-api/internal/model"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthConfig holds OAuth2 configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

// InitOAuth2Config initializes OAuth2 configuration
func (h *handler) InitOAuth2Config(cfg OAuthConfig) {
	h.oauth2Config = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       cfg.Scopes,
		Endpoint:     google.Endpoint,
	}
}

// OAuthLogin redirects user to Google OAuth2 authorization page
// @Summary Login with Google OAuth2
// @Description Redirects to Google OAuth2 authorization page
// @Tags Authentication
// @Produce json
// @Success 302 {string} string "Redirect to Google OAuth"
// @Router /authentication/login [get]
func (h *handler) OAuthLogin(c *gin.Context) {
	// Generate random state for CSRF protection
	state := generateRandomState()

	// Store state in session/cookie for validation
	c.SetCookie(
		"oauth_state",
		state,
		300, // 5 minutes
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true, // HttpOnly
	)

	// Redirect to Google OAuth2 authorization page
	authURL := h.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// OAuthCallback handles OAuth2 callback from Google
// @Summary OAuth2 callback handler
// @Description Handles OAuth2 callback, exchanges code for token, creates user session
// @Tags Authentication
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string true "State parameter for CSRF protection"
// @Success 302 {string} string "Redirect to dashboard"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Domain not allowed or account blocked"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /authentication/callback [get]
func (h *handler) OAuthCallback(c *gin.Context) {
	ctx := context.Background()

	// Validate state parameter (CSRF protection)
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		h.l.Error(ctx, "Invalid OAuth state parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_STATE",
				"message": "Invalid state parameter. Please try again.",
			},
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", h.config.Cookie.Domain, h.config.Cookie.Secure, true)

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		h.l.Error(ctx, "Missing authorization code")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_CODE",
				"message": "Authorization code is missing.",
			},
		})
		return
	}

	// Exchange code for token
	token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		h.l.Errorf(ctx, "Failed to exchange code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "OAUTH_EXCHANGE_FAILED",
				"message": "Failed to exchange authorization code. Please try again.",
			},
		})
		return
	}

	// Get user info from Google
	userInfo, err := h.getUserInfoFromGoogle(ctx, token.AccessToken)
	if err != nil {
		h.l.Errorf(ctx, "Failed to get user info from Google: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "USER_INFO_FAILED",
				"message": "Failed to retrieve user information from Google.",
			},
		})
		return
	}

	// Validate domain
	if !h.isAllowedDomain(userInfo.Email) {
		h.l.Warnf(ctx, "Domain not allowed: %s", userInfo.Email)

		// Publish audit event for failed login
		h.uc.PublishAuditEvent(ctx, audit.AuditEvent{
			UserID:       userInfo.Email, // Use email since user not created yet
			Action:       audit.ActionLoginFailed,
			ResourceType: "authentication",
			Metadata: map[string]string{
				"provider": "google",
				"reason":   "domain_not_allowed",
				"domain":   extractDomain(userInfo.Email),
			},
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})

		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "DOMAIN_NOT_ALLOWED",
				"message": "Your email domain is not allowed to access this system. Please contact admin.",
				"details": gin.H{
					"email":  userInfo.Email,
					"domain": extractDomain(userInfo.Email),
				},
			},
		})
		return
	}

	// Check blocklist
	if h.isBlockedEmail(userInfo.Email) {
		h.l.Warnf(ctx, "Account blocked: %s", userInfo.Email)

		// Publish audit event for failed login
		h.uc.PublishAuditEvent(ctx, audit.AuditEvent{
			UserID:       userInfo.Email,
			Action:       audit.ActionLoginFailed,
			ResourceType: "authentication",
			Metadata: map[string]string{
				"provider": "google",
				"reason":   "account_blocked",
			},
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})

		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "ACCOUNT_BLOCKED",
				"message": "Your account has been blocked. Please contact admin for assistance.",
			},
		})
		return
	}

	// Create or update user
	user, err := h.uc.CreateOrUpdateUser(ctx, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		h.l.Errorf(ctx, "Failed to create/update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "USER_CREATE_FAILED",
				"message": "Failed to create user account.",
			},
		})
		return
	}

	// Fetch Google Groups (Task 2.2)
	groups, err := h.groupsManager.GetUserGroups(ctx, userInfo.Email)
	if err != nil {
		h.l.Warnf(ctx, "Failed to fetch user groups (using default role): %v", err)
		groups = []string{} // Use empty groups if fetch fails
	}

	// Map groups to role (Task 2.3)
	role := h.roleMapper.MapGroupsToRole(groups)

	// Update user role in database
	user.SetRole(role)
	if err := h.uc.UpdateUserRole(ctx, user.ID, role); err != nil {
		h.l.Warnf(ctx, "Failed to update user role (continuing with role from groups): %v", err)
	}

	// Generate JWT token with role and groups (Task 1.4, 2.4)
	jwtToken, jti, err := h.generateJWTWithRoleAndGroups(user, role, groups)
	if err != nil {
		h.l.Errorf(ctx, "Failed to generate JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "JWT_GENERATION_FAILED",
				"message": "Failed to generate authentication token.",
			},
		})
		return
	}

	// Create session in Redis (Task 1.6)
	rememberMe := c.Query("remember_me") == "true"
	if err := h.createSession(ctx, user.ID, jti, rememberMe); err != nil {
		h.l.Errorf(ctx, "Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "SESSION_CREATE_FAILED",
				"message": "Failed to create session.",
			},
		})
		return
	}

	// Set JWT as HttpOnly cookie (Task 1.7)
	h.setAuthCookie(c, jwtToken)

	// Publish audit log event (Task 2.6)
	h.uc.PublishAuditEvent(ctx, audit.AuditEvent{
		UserID:       user.ID,
		Action:       audit.ActionLogin,
		ResourceType: "authentication",
		Metadata: map[string]string{
			"provider": "google",
			"role":     role,
			"groups":   fmt.Sprintf("%v", groups),
		},
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	// Redirect to dashboard
	c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

// extractDomain extracts domain from email address
func extractDomain(email string) string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return email[i+1:]
		}
	}
	return ""
}

// generateRandomState generates a random state for CSRF protection
func generateRandomState() string {
	// TODO: Implement secure random state generation using crypto/rand
	// For now, use a simple UUID
	return fmt.Sprintf("state-%d", time.Now().UnixNano())
}

// getUserInfoFromGoogle fetches user information from Google UserInfo API
func (h *handler) getUserInfoFromGoogle(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Call Google UserInfo API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call UserInfo API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("UserInfo API returned status %d", resp.StatusCode)
	}

	// Parse response
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode UserInfo response: %w", err)
	}

	return &userInfo, nil
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Helper methods for handler

func (h *handler) isAllowedDomain(email string) bool {
	domain := extractDomain(email)
	for _, allowedDomain := range h.config.AccessControl.AllowedDomains {
		if domain == allowedDomain {
			return true
		}
	}
	return false
}

func (h *handler) isBlockedEmail(email string) bool {
	for _, blockedEmail := range h.config.AccessControl.BlockedEmails {
		if email == blockedEmail {
			return true
		}
	}
	return false
}

func (h *handler) generateJWT(user *model.User) (string, string, error) {
	// Get user role (decrypted from RoleHash)
	role := user.GetRole()

	// TODO: Fetch Google Groups in Task 2.2
	groups := []string{}

	// Generate JWT token with RS256
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, role, groups)
	if err != nil {
		return "", "", err
	}

	// Extract JTI from token for session management
	claims, err := h.jwtManager.VerifyToken(token)
	if err != nil {
		return "", "", err
	}

	return token, claims.ID, nil
}

func (h *handler) generateJWTWithRoleAndGroups(user *model.User, role string, groups []string) (string, string, error) {
	// Generate JWT token with RS256, including role and groups
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, role, groups)
	if err != nil {
		return "", "", err
	}

	// Extract JTI from token for session management
	claims, err := h.jwtManager.VerifyToken(token)
	if err != nil {
		return "", "", err
	}

	return token, claims.ID, nil
}

func (h *handler) createSession(ctx context.Context, userID, jti string, rememberMe bool) error {
	return h.sessionManager.CreateSession(ctx, userID, jti, rememberMe)
}

func (h *handler) setAuthCookie(c *gin.Context, token string) {
	c.SetCookie(
		h.config.Cookie.Name,
		token,
		h.config.Cookie.MaxAge,
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true, // HttpOnly
	)

	// Add SameSite attribute
	h.addSameSiteAttribute(c, h.config.Cookie.SameSite)
}
