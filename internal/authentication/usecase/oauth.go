package usecase

import (
	"context"
	"smap-api/internal/audit"
	"smap-api/internal/authentication"

	"golang.org/x/oauth2"
)

// InitiateOAuthLogin generates the OAuth authorization URL and state
func (u *ImplUsecase) InitiateOAuthLogin(ctx context.Context, input authentication.OAuthLoginInput) (*authentication.OAuthLoginOutput, error) {
	if u.oauthProvider == nil {
		return nil, authentication.ErrInvalidProvider
	}

	state := u.generateState()
	authURL := u.oauthProvider.GetAuthCodeURL(state, oauth2.AccessTypeOffline)

	return &authentication.OAuthLoginOutput{
		AuthURL: authURL,
		State:   state,
	}, nil
}

// ProcessOAuthCallback handles the entire OAuth callback business logic:
// exchange code → get user info → validate domain → create/update user →
// fetch groups → map role → generate JWT → create session → audit
func (u *ImplUsecase) ProcessOAuthCallback(ctx context.Context, input authentication.OAuthCallbackInput) (*authentication.OAuthCallbackOutput, error) {
	// 1. Exchange code for token via OAuth provider
	token, err := u.oauthProvider.ExchangeCode(ctx, input.Code)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.ProcessOAuthCallback.ExchangeCode: %v", err)
		return nil, err
	}

	// 2. Get user info from provider
	userInfo, err := u.oauthProvider.GetUserInfo(ctx, token)
	if err != nil {
		u.l.Errorf(ctx, "authentication.usecase.ProcessOAuthCallback.GetUserInfo: %v", err)
		return nil, err
	}

	// 3. Validate domain (business rule)
	if !u.isAllowedDomain(userInfo.Email) {
		u.recordFailedAttempt(ctx, input.IPAddress)
		u.PublishAuditEvent(ctx, audit.AuditEvent{
			UserID:       userInfo.Email,
			Action:       audit.ActionLoginFailed,
			ResourceType: "authentication",
			Metadata: map[string]string{
				"provider": u.oauthProvider.GetProviderName(),
				"reason":   "domain_not_allowed",
				"domain":   u.extractDomain(userInfo.Email),
			},
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
		})
		return nil, authentication.ErrDomainNotAllowed
	}

	// 4. Check blocklist (business rule)
	if u.isBlockedEmail(userInfo.Email) {
		u.recordFailedAttempt(ctx, input.IPAddress)
		u.PublishAuditEvent(ctx, audit.AuditEvent{
			UserID:       userInfo.Email,
			Action:       audit.ActionLoginFailed,
			ResourceType: "authentication",
			Metadata: map[string]string{
				"provider": u.oauthProvider.GetProviderName(),
				"reason":   "account_blocked",
			},
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
		})
		return nil, authentication.ErrAccountBlocked
	}

	// 5. Create or update user
	usr, err := u.createOrUpdateUser(ctx, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		return nil, err
	}

	// 6. Fetch groups and map to role
	u.l.Infof(ctx, "Fetching groups for %s", userInfo.Email)
	groups, err := u.getUserGroups(ctx, userInfo.Email)
	if err != nil {
		u.l.Warnf(ctx, "authentication.usecase.ProcessOAuthCallback.GetUserGroups: %v", err)
		groups = []string{}
	}

	u.l.Infof(ctx, "Mapping groups to role")
	role := u.mapGroupsToRole(groups)
	u.l.Infof(ctx, "Role mapped: %s", role)

	// 7. Update user role
	u.l.Infof(ctx, "Setting user role in memory")
	usr.SetRole(role)
	u.l.Infof(ctx, "Updating user role in DB")
	if err := u.updateUserRole(ctx, usr.ID, role); err != nil {
		u.l.Warnf(ctx, "authentication.usecase.ProcessOAuthCallback.UpdateUserRole: %v", err)
	}

	// 8. Generate JWT token
	u.l.Infof(ctx, "Generating JWT token")
	jwtToken, jti, err := u.generateToken(ctx, usr, role, groups)
	if err != nil {
		return nil, err
	}

	// 9. Create session
	if err := u.createSession(ctx, usr.ID, jti, input.RememberMe); err != nil {
		return nil, err
	}

	// 10. Clear failed attempts on success
	u.clearFailedAttempts(ctx, input.IPAddress)

	// 11. Publish audit event
	u.PublishAuditEvent(ctx, audit.AuditEvent{
		UserID:       usr.ID,
		Action:       audit.ActionLogin,
		ResourceType: "authentication",
		Metadata: map[string]string{
			"provider": u.oauthProvider.GetProviderName(),
			"role":     role,
		},
		IPAddress: input.IPAddress,
		UserAgent: input.UserAgent,
	})

	return &authentication.OAuthCallbackOutput{
		Token: jwtToken,
	}, nil
}
