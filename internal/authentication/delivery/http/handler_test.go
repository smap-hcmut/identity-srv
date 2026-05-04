package http

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"identity-srv/config"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	sharedauth "github.com/smap-hcmut/shared-libs/go/auth"
	sharedlog "github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDeps struct {
	uc *authentication.MockUseCase
}

func initHandler(t *testing.T, environment string) (handler, mockDeps) {
	t.Helper()

	l := sharedlog.NewZapLogger(sharedlog.ZapConfig{
		Level:        sharedlog.LevelFatal,
		Mode:         sharedlog.ModeProduction,
		Encoding:     sharedlog.EncodingJSON,
		ColorEnabled: false,
	})
	uc := authentication.NewMockUseCase(t)
	cfg := &config.Config{
		Environment: config.EnvironmentConfig{Name: environment},
		Cookie: config.CookieConfig{
			Name:   "smap_auth_token",
			MaxAge: 3600,
			Domain: ".example.com",
		},
		JWT: config.JWTConfig{SecretKey: "test-secret"},
	}

	return handler{
		l:            l,
		uc:           uc,
		cookieConfig: cfg.Cookie,
		config:       cfg,
		stateSecret:  cfg.JWT.SecretKey,
	}, mockDeps{uc: uc}
}

func TestLogout(t *testing.T) {
	type mockUCLogout struct {
		isCalled bool
		input    model.Scope
		err      error
	}

	tcs := map[string]struct {
		input    sharedauth.Payload
		mock     mockUCLogout
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success": {
			input: sharedauth.Payload{UserID: "user-1", Username: "user@example.com", Role: model.RoleAdmin, StandardClaims: jwtClaims("jti-1")},
			mock: mockUCLogout{
				isCalled: true,
				input:    model.Scope{UserID: "user-1", Username: "user@example.com", Role: model.RoleAdmin, JTI: "jti-1"},
			},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success"}`,
		},
		"unauthorized": {
			wantCode: 401,
			wantBody: `{"error_code":401,"message":"Authentication required"}`,
		},
		"usecase_error": {
			input: sharedauth.Payload{UserID: "user-1"},
			mock: mockUCLogout{
				isCalled: true,
				input:    model.Scope{UserID: "user-1"},
				err:      authentication.ErrInternalSystem,
			},
			wantCode: 400,
			wantBody: `{"error_code":20022,"message":"Internal system error"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, string(model.EnvironmentDevelopment))
			w, c, engine := newAuthRequest("POST", "/logout", nil, tc.input)
			engine.POST("/logout", h.Logout)
			if tc.mock.isCalled {
				deps.uc.EXPECT().Logout(c.Request.Context(), tc.mock.input).Return(tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
			if tc.mock.err == nil && tc.mock.isCalled {
				require.Contains(t, w.Header().Values("Set-Cookie")[0], "Max-Age=0")
			}
		})
	}
}

func TestGetMe(t *testing.T) {
	name := "User Name"
	type mockUCGetCurrentUser struct {
		isCalled bool
		input    model.Scope
		output   *model.User
		err      error
	}

	tcs := map[string]struct {
		input    sharedauth.Payload
		mock     mockUCGetCurrentUser
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success": {
			input: sharedauth.Payload{StandardClaims: jwt.StandardClaims{Subject: "user-1"}, Username: "user@example.com", Role: model.RoleViewer},
			mock: mockUCGetCurrentUser{
				isCalled: true,
				input:    model.Scope{UserID: "user-1", Username: "user@example.com", Role: model.RoleViewer},
				output:   &model.User{ID: "user-1", Email: "user@example.com", Name: &name},
			},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"id":"user-1","email":"user@example.com","full_name":"User Name"}}`,
		},
		"unauthorized": {
			wantCode: 401,
			wantBody: `{"error_code":401,"message":"Authentication required"}`,
		},
		"usecase_error": {
			input: sharedauth.Payload{UserID: "user-1"},
			mock: mockUCGetCurrentUser{
				isCalled: true,
				input:    model.Scope{UserID: "user-1"},
				err:      authentication.ErrUserNotFound,
			},
			wantCode: 400,
			wantBody: `{"error_code":20002,"message":"User not found"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, string(model.EnvironmentDevelopment))
			w, c, engine := newAuthRequest("GET", "/me", nil, tc.input)
			engine.GET("/me", h.GetMe)
			if tc.mock.isCalled {
				deps.uc.EXPECT().GetCurrentUser(c.Request.Context(), tc.mock.input).Return(tc.mock.output, tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestOAuthLogin(t *testing.T) {
	type mockUCInitiateOAuthLogin struct {
		isCalled bool
		input    authentication.OAuthLoginInput
		output   *authentication.OAuthLoginOutput
		err      error
	}

	tcs := map[string]struct {
		input    string
		mock     mockUCInitiateOAuthLogin
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success": {
			input: "/login?redirect=https://app.example.com/dashboard",
			mock: mockUCInitiateOAuthLogin{
				isCalled: true,
				input:    authentication.OAuthLoginInput{RedirectURL: "https://app.example.com/dashboard"},
				output:   &authentication.OAuthLoginOutput{AuthURL: "https://oauth.example.com/auth"},
			},
			wantCode: 307,
			output:   "https://oauth.example.com/auth",
		},
		"usecase_error": {
			input: "/login?redirect=://bad",
			mock: mockUCInitiateOAuthLogin{
				isCalled: true,
				input:    authentication.OAuthLoginInput{RedirectURL: "://bad"},
				err:      authentication.ErrInvalidRedirectURL,
			},
			wantCode: 400,
			wantBody: `{"error_code":20021,"message":"Invalid redirect URL"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, string(model.EnvironmentDevelopment))
			w, c, engine := newAuthRequest("GET", tc.input, nil, sharedauth.Payload{})
			engine.GET("/login", h.OAuthLogin)
			if tc.mock.isCalled {
				deps.uc.EXPECT().InitiateOAuthLogin(c.Request.Context(), mock.MatchedBy(func(input authentication.OAuthLoginInput) bool {
					return input.RedirectURL == tc.mock.input.RedirectURL && input.State != ""
				})).Return(tc.mock.output, tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			if tc.output != "" {
				require.Equal(t, tc.output, w.Header().Get("Location"))
			}
			if tc.wantBody != "" {
				require.JSONEq(t, tc.wantBody, w.Body.String())
			}
		})
	}
}

func TestOAuthCallback(t *testing.T) {
	type mockUCProcessOAuthCallback struct {
		isCalled bool
		input    authentication.OAuthCallbackInput
		output   *authentication.OAuthCallbackOutput
		err      error
	}

	tcs := map[string]struct {
		input       string
		environment string
		redirect    string
		mock        mockUCProcessOAuthCallback
		output      string
		err         error
		wantCode    int
		wantBody    string
	}{
		"success_development": {
			input:       "/callback?code=code-1&remember_me=true",
			environment: string(model.EnvironmentDevelopment),
			mock: mockUCProcessOAuthCallback{
				isCalled: true,
				input:    authentication.OAuthCallbackInput{Code: "code-1", RememberMe: true},
				output:   &authentication.OAuthCallbackOutput{Token: "jwt-token"},
			},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"token":"jwt-token"}}`,
		},
		"success_production_redirect": {
			input:       "/callback?code=code-1",
			environment: string(model.EnvironmentProduction),
			redirect:    "https://app.example.com/dashboard",
			mock: mockUCProcessOAuthCallback{
				isCalled: true,
				input:    authentication.OAuthCallbackInput{Code: "code-1"},
				output:   &authentication.OAuthCallbackOutput{Token: "jwt-token"},
			},
			wantCode: 307,
			output:   "https://app.example.com/dashboard?token=jwt-token",
		},
		"missing_code": {
			input:       "/callback",
			environment: string(model.EnvironmentDevelopment),
			wantCode:    400,
			wantBody:    `{"error_code":20015,"message":"Missing authorization code"}`,
		},
		"usecase_error": {
			input:       "/callback?code=code-1",
			environment: string(model.EnvironmentDevelopment),
			mock: mockUCProcessOAuthCallback{
				isCalled: true,
				input:    authentication.OAuthCallbackInput{Code: "code-1"},
				err:      authentication.ErrDomainNotAllowed,
			},
			wantCode: 400,
			wantBody: `{"error_code":20012,"message":"Domain not allowed"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, tc.environment)
			state, err := h.generateSignedState(tc.redirect)
			require.NoError(t, err)
			separator := "?"
			if bytes.Contains([]byte(tc.input), []byte("?")) {
				separator = "&"
			}
			w, c, engine := newAuthRequest("GET", tc.input+separator+"state="+state, nil, sharedauth.Payload{})
			engine.GET("/callback", h.OAuthCallback)
			if tc.mock.isCalled {
				deps.uc.EXPECT().ProcessOAuthCallback(c.Request.Context(), mock.MatchedBy(func(input authentication.OAuthCallbackInput) bool {
					return input.Code == tc.mock.input.Code && input.RememberMe == tc.mock.input.RememberMe
				})).Return(tc.mock.output, tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			if tc.output != "" {
				require.Equal(t, tc.output, w.Header().Get("Location"))
			}
			if tc.wantBody != "" {
				require.JSONEq(t, tc.wantBody, w.Body.String())
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	expiresAt := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	type mockUCValidateToken struct {
		isCalled bool
		input    string
		output   *authentication.TokenValidationResult
		err      error
	}

	tcs := map[string]struct {
		input    string
		mock     mockUCValidateToken
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success_valid": {
			input: `{"token":"jwt-token"}`,
			mock: mockUCValidateToken{
				isCalled: true,
				input:    "jwt-token",
				output:   &authentication.TokenValidationResult{Valid: true, UserID: "user-1", Email: "user@example.com", Role: model.RoleAdmin, Groups: []string{"admins"}, ExpiresAt: expiresAt},
			},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"valid":true,"user_id":"user-1","email":"user@example.com","role":"ADMIN","groups":["admins"],"expires_at":"2026-05-04T00:00:00Z"}}`,
		},
		"wrong_body": {
			input:    `{}`,
			wantCode: 400,
			wantBody: `{"error_code":20001,"message":"Wrong body"}`,
		},
		"usecase_error": {
			input: `{"token":"jwt-token"}`,
			mock: mockUCValidateToken{
				isCalled: true,
				input:    "jwt-token",
				err:      authentication.ErrInvalidEmail,
			},
			wantCode: 400,
			wantBody: `{"error_code":20011,"message":"Invalid email"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, string(model.EnvironmentDevelopment))
			w, c, engine := newAuthRequest("POST", "/internal/validate", bytes.NewBufferString(tc.input), sharedauth.Payload{})
			engine.POST("/internal/validate", h.ValidateToken)
			if tc.mock.isCalled {
				deps.uc.EXPECT().ValidateToken(c.Request.Context(), tc.mock.input).Return(tc.mock.output, tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestRevokeToken(t *testing.T) {
	type mockUCRevoke struct {
		isCalled bool
		jti      string
		userID   string
		err      error
	}

	tcs := map[string]struct {
		input    string
		mock     mockUCRevoke
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success_jti": {
			input:    `{"jti":"jti-1"}`,
			mock:     mockUCRevoke{isCalled: true, jti: "jti-1"},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"message":"Token revoked successfully"}}`,
		},
		"success_user_id": {
			input:    `{"user_id":"user-1"}`,
			mock:     mockUCRevoke{isCalled: true, userID: "user-1"},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"message":"All tokens for user user-1 revoked successfully"}}`,
		},
		"missing_jti_and_user_id": {
			input:    `{}`,
			wantCode: 400,
			wantBody: `{"error_code":20017,"message":"Must provide either jti or user_id"}`,
		},
		"conflict_jti_and_user_id": {
			input:    `{"jti":"jti-1","user_id":"user-1"}`,
			wantCode: 400,
			wantBody: `{"error_code":20018,"message":"Cannot provide both jti and user_id"}`,
		},
		"usecase_error": {
			input:    `{"jti":"jti-1"}`,
			mock:     mockUCRevoke{isCalled: true, jti: "jti-1", err: authentication.ErrUserNotFound},
			wantCode: 400,
			wantBody: `{"error_code":20002,"message":"User not found"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, string(model.EnvironmentDevelopment))
			w, c, engine := newAuthRequest("POST", "/internal/revoke-token", bytes.NewBufferString(tc.input), sharedauth.Payload{})
			engine.POST("/internal/revoke-token", h.RevokeToken)
			if tc.mock.isCalled && tc.mock.jti != "" {
				deps.uc.EXPECT().RevokeToken(c.Request.Context(), tc.mock.jti).Return(tc.mock.err)
			}
			if tc.mock.isCalled && tc.mock.userID != "" {
				deps.uc.EXPECT().RevokeAllUserTokens(c.Request.Context(), tc.mock.userID).Return(tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestGetUserByID(t *testing.T) {
	now := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	name := "User Name"
	avatar := "avatar.png"
	roleHash, err := model.EncryptRole(model.RoleAdmin)
	require.NoError(t, err)
	type mockUCGetUserByID struct {
		isCalled bool
		input    string
		output   *model.User
		err      error
	}

	tcs := map[string]struct {
		input    string
		mock     mockUCGetUserByID
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success": {
			input: "/internal/users/user-1",
			mock: mockUCGetUserByID{
				isCalled: true,
				input:    "user-1",
				output:   &model.User{ID: "user-1", Email: "user@example.com", Name: &name, AvatarURL: &avatar, RoleHash: &roleHash, IsActive: true, CreatedAt: now, UpdatedAt: now},
			},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"id":"user-1","email":"user@example.com","name":"User Name","avatar_url":"avatar.png","role":"ADMIN","is_active":true,"created_at":"2026-05-04T00:00:00Z","updated_at":"2026-05-04T00:00:00Z"}}`,
		},
		"usecase_error": {
			input: "/internal/users/user-1",
			mock: mockUCGetUserByID{
				isCalled: true,
				input:    "user-1",
				err:      authentication.ErrUserNotFound,
			},
			wantCode: 400,
			wantBody: `{"error_code":20002,"message":"User not found"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t, string(model.EnvironmentDevelopment))
			w, c, engine := newAuthRequest("GET", tc.input, nil, sharedauth.Payload{})
			engine.GET("/internal/users/:id", h.GetUserByID)
			if tc.mock.isCalled {
				deps.uc.EXPECT().GetUserByID(c.Request.Context(), tc.mock.input).Return(tc.mock.output, tc.mock.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestNew(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  *config.Config
		mock   mock
		output bool
		err    error
	}{
		"success": {
			input: &config.Config{
				Cookie: config.CookieConfig{Name: "smap_auth_token"},
				JWT:    config.JWTConfig{SecretKey: "test-secret"},
			},
			output: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h := New(nil, nil, nil, tc.input)

			require.Equal(t, tc.output, h != nil)
			require.NoError(t, tc.err)
		})
	}
}

func TestRegisterRoutes(t *testing.T) {
	type mockData struct{}

	tcs := map[string]struct {
		input  string
		mock   mockData
		output []string
		err    error
	}{
		"success": {
			input: "/authentication",
			output: []string{
				"GET /authentication/login",
				"GET /authentication/callback",
				"POST /authentication/logout",
				"GET /authentication/me",
				"POST /authentication/internal/validate",
				"POST /authentication/internal/revoke-token",
				"GET /authentication/internal/users/:id",
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			h, _ := initHandler(t, string(model.EnvironmentDevelopment))
			mw := middleware.New(middleware.Config{})

			h.RegisterRoutes(engine.Group(tc.input), mw)

			routes := make([]string, 0, len(engine.Routes()))
			for _, route := range engine.Routes() {
				routes = append(routes, route.Method+" "+route.Path)
			}
			require.ElementsMatch(t, tc.output, routes)
			require.NoError(t, tc.err)
		})
	}
}

func TestMapError(t *testing.T) {
	type mockData struct{}

	tcs := map[string]struct {
		input  error
		mock   mockData
		output string
		err    error
	}{
		"user_not_found":            {input: authentication.ErrUserNotFound, output: "User not found"},
		"username_existed":          {input: authentication.ErrUsernameExisted, output: "Username existed"},
		"wrong_password":            {input: authentication.ErrWrongPassword, output: "Wrong password"},
		"wrong_otp":                 {input: authentication.ErrWrongOTP, output: "Wrong OTP"},
		"otp_expired":               {input: authentication.ErrOTPExpired, output: "OTP expired"},
		"too_many_attempts":         {input: authentication.ErrTooManyAttempts, output: "Too many attempts"},
		"user_not_verified":         {input: authentication.ErrUserNotVerified, output: "User not verified"},
		"invalid_provider":          {input: authentication.ErrInvalidProvider, output: "Invalid provider"},
		"invalid_email":             {input: authentication.ErrInvalidEmail, output: "Invalid email"},
		"user_verified":             {input: authentication.ErrUserVerified, output: "User verified"},
		"domain_not_allowed":        {input: authentication.ErrDomainNotAllowed, output: "Domain not allowed"},
		"account_blocked":           {input: authentication.ErrAccountBlocked, output: "Account blocked"},
		"scope_not_found":           {input: authentication.ErrScopeNotFound, output: "Scope not found"},
		"configuration_missing":     {input: authentication.ErrConfigurationMissing, output: "Server configuration missing"},
		"invalid_redirect_url":      {input: authentication.ErrInvalidRedirectURL, output: "Invalid redirect URL"},
		"redirect_url_not_allowed":  {input: authentication.ErrRedirectURLNotAllowed, output: "Invalid redirect URL"},
		"internal_system":           {input: authentication.ErrInternalSystem, output: "Internal system error"},
		"user_creation":             {input: authentication.ErrUserCreation, output: "Failed to create or update user"},
		"unknown_error_passthrough": {input: errors.New("unknown"), output: "unknown"},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, _ := initHandler(t, string(model.EnvironmentDevelopment))

			output := h.mapError(tc.input)

			require.EqualError(t, output, tc.output)
			require.NoError(t, tc.err)
		})
	}
}

func newAuthRequest(method string, target string, body *bytes.Buffer, payload sharedauth.Payload) (*httptest.ResponseRecorder, *gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	c.Request = httptest.NewRequest(method, target, body)
	c.Request.Header.Set("Content-Type", "application/json")
	if payload.UserID != "" || payload.Subject != "" {
		c.Request = c.Request.WithContext(sharedauth.SetPayloadToContext(c.Request.Context(), payload))
	}
	return w, c, engine
}

func jwtClaims(id string) jwt.StandardClaims {
	return jwt.StandardClaims{Id: id}
}
