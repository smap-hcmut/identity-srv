package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	sharedauth "github.com/smap-hcmut/shared-libs/go/auth"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"
	"identity-srv/internal/user"
	"identity-srv/pkg/oauth"
)

func sharedAuthPayload(claims jwt.StandardClaims) sharedauth.Payload {
	return sharedauth.Payload{
		UserID:   "user-id",
		Username: "admin@example.com",
		Role:     model.RoleAdmin,
		StandardClaims: jwt.StandardClaims{
			Id:        claims.Id,
			ExpiresAt: claims.ExpiresAt,
		},
	}
}

func TestInitiateOAuthLogin(t *testing.T) {
	ctx := context.Background()

	type mockProviderAuthCodeURL struct {
		isCalled bool
		input    string
		output   string
	}

	type mockProvider struct {
		authCodeURL mockProviderAuthCodeURL
	}

	type mock struct {
		provider mockProvider
	}

	tcs := map[string]struct {
		input struct {
			oauthInput  authentication.OAuthLoginInput
			useProvider bool
		}
		mock   mock
		output *authentication.OAuthLoginOutput
		err    error
	}{
		"err invalid provider": {
			input: struct {
				oauthInput  authentication.OAuthLoginInput
				useProvider bool
			}{
				oauthInput: authentication.OAuthLoginInput{State: "state"},
			},
			err: authentication.ErrInvalidProvider,
		},
		"success": {
			input: struct {
				oauthInput  authentication.OAuthLoginInput
				useProvider bool
			}{
				oauthInput:  authentication.OAuthLoginInput{State: "state"},
				useProvider: true,
			},
			mock: mock{provider: mockProvider{authCodeURL: mockProviderAuthCodeURL{
				isCalled: true,
				input:    "state",
				output:   "https://auth.example.com",
			}}},
			output: &authentication.OAuthLoginOutput{
				AuthURL: "https://auth.example.com",
				State:   "state",
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			if !tc.input.useProvider {
				uc.SetOAuthProvider(nil)
			}

			if tc.mock.provider.authCodeURL.isCalled {
				deps.provider.EXPECT().GetAuthCodeURL(tc.mock.provider.authCodeURL.input, oauth2.AccessTypeOffline).
					Return(tc.mock.provider.authCodeURL.output)
			}

			res, err := uc.InitiateOAuthLogin(ctx, tc.input.oauthInput)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestProcessOAuthCallback(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("expected error")
	token := &oauth2.Token{AccessToken: "access-token"}
	expectedUser := model.User{ID: "user-id", Email: "admin@example.com"}
	_ = expectedUser.SetRole(model.RoleViewer)

	type mockProviderExchangeCode struct {
		isCalled bool
		input    string
		output   *oauth2.Token
		err      error
	}

	type mockProviderGetUserInfo struct {
		isCalled bool
		input    *oauth2.Token
		output   *oauth.UserInfo
		err      error
	}

	type mockUserCreate struct {
		isCalled bool
		input    user.CreateInput
		output   model.User
		err      error
	}

	type mockUserUpdate struct {
		isCalled bool
		input    user.UpdateInput
		err      error
	}

	type mockJWTCreateToken struct {
		isCalled bool
		output   string
		err      error
	}

	type mockJWTVerify struct {
		isCalled bool
		input    string
		output   jwt.StandardClaims
		err      error
	}

	type mockRedisSet struct {
		isCalled bool
		input    string
		err      error
	}

	type mockRedisGet struct {
		isCalled bool
		input    string
		err      error
	}

	type mockProvider struct {
		exchangeCode mockProviderExchangeCode
		getUserInfo  mockProviderGetUserInfo
	}

	type mockUser struct {
		create mockUserCreate
		update mockUserUpdate
	}

	type mockJWT struct {
		createToken mockJWTCreateToken
		verify      mockJWTVerify
	}

	type mockRedis struct {
		setSession     mockRedisSet
		getSessions    mockRedisGet
		setUserSession mockRedisSet
	}

	type mock struct {
		provider mockProvider
		user     mockUser
		jwt      mockJWT
		redis    mockRedis
	}

	successMock := mock{
		provider: mockProvider{
			exchangeCode: mockProviderExchangeCode{isCalled: true, input: "code", output: token},
			getUserInfo: mockProviderGetUserInfo{
				isCalled: true,
				input:    token,
				output:   &oauth.UserInfo{Email: "admin@example.com", Name: "Admin", Picture: "avatar"},
			},
		},
		user: mockUser{
			create: mockUserCreate{
				isCalled: true,
				input: user.CreateInput{
					Email:     "admin@example.com",
					Name:      "Admin",
					AvatarURL: "avatar",
				},
				output: expectedUser,
			},
			update: mockUserUpdate{
				isCalled: true,
				input:    user.UpdateInput{UserID: "user-id", Role: model.RoleAdmin},
			},
		},
		jwt: mockJWT{
			createToken: mockJWTCreateToken{isCalled: true, output: "jwt-token"},
			verify: mockJWTVerify{
				isCalled: true,
				input:    "jwt-token",
				output:   jwt.StandardClaims{Id: "jti"},
			},
		},
		redis: mockRedis{
			setSession:     mockRedisSet{isCalled: true, input: "session:jti"},
			getSessions:    mockRedisGet{isCalled: true, input: "user_sessions:user-id", err: errors.New("missing")},
			setUserSession: mockRedisSet{isCalled: true, input: "user_sessions:user-id"},
		},
	}

	tcs := map[string]struct {
		input  authentication.OAuthCallbackInput
		mock   mock
		output *authentication.OAuthCallbackOutput
		err    error
	}{
		"success": {
			input:  authentication.OAuthCallbackInput{Code: "code"},
			mock:   successMock,
			output: &authentication.OAuthCallbackOutput{Token: "jwt-token"},
		},
		"success update role error ignored": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: func() mock {
				m := successMock
				m.user.update.err = expectedErr
				return m
			}(),
			output: &authentication.OAuthCallbackOutput{Token: "jwt-token"},
		},
		"err exchange code": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: mock{provider: mockProvider{exchangeCode: mockProviderExchangeCode{
				isCalled: true,
				input:    "code",
				err:      expectedErr,
			}}},
			err: expectedErr,
		},
		"err get user info": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: mock{provider: mockProvider{
				exchangeCode: mockProviderExchangeCode{isCalled: true, input: "code", output: token},
				getUserInfo:  mockProviderGetUserInfo{isCalled: true, input: token, err: expectedErr},
			}},
			err: expectedErr,
		},
		"err domain not allowed": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: mock{provider: mockProvider{
				exchangeCode: mockProviderExchangeCode{isCalled: true, input: "code", output: token},
				getUserInfo: mockProviderGetUserInfo{
					isCalled: true,
					input:    token,
					output:   &oauth.UserInfo{Email: "user@other.com"},
				},
			}},
			err: authentication.ErrDomainNotAllowed,
		},
		"err account blocked": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: mock{provider: mockProvider{
				exchangeCode: mockProviderExchangeCode{isCalled: true, input: "code", output: token},
				getUserInfo: mockProviderGetUserInfo{
					isCalled: true,
					input:    token,
					output:   &oauth.UserInfo{Email: "blocked@example.com"},
				},
			}},
			err: authentication.ErrAccountBlocked,
		},
		"err create user": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: mock{
				provider: mockProvider{
					exchangeCode: mockProviderExchangeCode{isCalled: true, input: "code", output: token},
					getUserInfo: mockProviderGetUserInfo{
						isCalled: true,
						input:    token,
						output:   &oauth.UserInfo{Email: "admin@example.com", Name: "Admin", Picture: "avatar"},
					},
				},
				user: mockUser{create: mockUserCreate{
					isCalled: true,
					input:    user.CreateInput{Email: "admin@example.com", Name: "Admin", AvatarURL: "avatar"},
					err:      expectedErr,
				}},
			},
			err: authentication.ErrUserCreation,
		},
		"err create token": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: func() mock {
				m := successMock
				m.jwt.createToken.err = expectedErr
				m.jwt.createToken.output = ""
				m.jwt.verify.isCalled = false
				m.redis.setSession.isCalled = false
				m.redis.getSessions.isCalled = false
				m.redis.setUserSession.isCalled = false
				return m
			}(),
			err: expectedErr,
		},
		"err verify token": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: func() mock {
				m := successMock
				m.jwt.verify.err = expectedErr
				m.redis.setSession.isCalled = false
				m.redis.getSessions.isCalled = false
				m.redis.setUserSession.isCalled = false
				return m
			}(),
			err: expectedErr,
		},
		"err create session": {
			input: authentication.OAuthCallbackInput{Code: "code"},
			mock: func() mock {
				m := successMock
				m.redis.setSession.err = expectedErr
				m.redis.getSessions.isCalled = false
				m.redis.setUserSession.isCalled = false
				return m
			}(),
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.provider.exchangeCode.isCalled {
				deps.provider.EXPECT().ExchangeCode(ctx, tc.mock.provider.exchangeCode.input).
					Return(tc.mock.provider.exchangeCode.output, tc.mock.provider.exchangeCode.err)
			}
			if tc.mock.provider.getUserInfo.isCalled {
				deps.provider.EXPECT().GetUserInfo(ctx, tc.mock.provider.getUserInfo.input).
					Return(tc.mock.provider.getUserInfo.output, tc.mock.provider.getUserInfo.err)
			}
			if tc.mock.user.create.isCalled {
				deps.userUC.EXPECT().Create(ctx, tc.mock.user.create.input).
					Return(tc.mock.user.create.output, tc.mock.user.create.err)
			}
			if tc.mock.user.update.isCalled {
				deps.userUC.EXPECT().Update(ctx, tc.mock.user.update.input).
					Return(tc.mock.user.update.err)
			}
			if tc.mock.jwt.createToken.isCalled {
				deps.jwt.EXPECT().CreateToken(testifymock.Anything).
					Return(tc.mock.jwt.createToken.output, tc.mock.jwt.createToken.err)
			}
			if tc.mock.jwt.verify.isCalled {
				deps.jwt.EXPECT().Verify(tc.mock.jwt.verify.input).
					Return(sharedAuthPayload(tc.mock.jwt.verify.output), tc.mock.jwt.verify.err)
			}
			if tc.mock.redis.setSession.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.setSession.input, testifymock.Anything, testifymock.Anything).
					Return(tc.mock.redis.setSession.err)
			}
			if tc.mock.redis.getSessions.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.getSessions.input).
					Return("", tc.mock.redis.getSessions.err)
			}
			if tc.mock.redis.setUserSession.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.setUserSession.input, testifymock.Anything, 7*24*time.Hour).
					Return(tc.mock.redis.setUserSession.err)
			}

			res, err := uc.ProcessOAuthCallback(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestValidateRedirectURL(t *testing.T) {
	tcs := map[string]struct {
		input struct {
			allowedURLs []string
			redirectURL string
		}
		err error
	}{
		"success empty": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{redirectURL: ""},
		},
		"success relative": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{redirectURL: "/dashboard"},
		},
		"success exact absolute": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{
				allowedURLs: []string{"https://app.example.com"},
				redirectURL: "https://app.example.com/callback",
			},
		},
		"success wildcard": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{
				allowedURLs: []string{"*.example.com"},
				redirectURL: "https://sub.example.com/callback",
			},
		},
		"success skips invalid allowed url": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{
				allowedURLs: []string{":", "https://app.example.com"},
				redirectURL: "https://app.example.com/callback",
			},
		},
		"err invalid redirect url": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{
				redirectURL: ":",
			},
			err: authentication.ErrInvalidRedirectURL,
		},
		"err not allowed": {
			input: struct {
				allowedURLs []string
				redirectURL string
			}{
				allowedURLs: []string{"https://app.example.com"},
				redirectURL: "https://evil.example.net",
			},
			err: authentication.ErrRedirectURLNotAllowed,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			validator := NewRedirectValidator(tc.input.allowedURLs)

			err := validator.ValidateRedirectURL(tc.input.redirectURL)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRoleMapperMapEmailToRole(t *testing.T) {
	mapper := &RoleMapper{
		userRoles:   map[string]string{"admin@example.com": model.RoleAdmin},
		defaultRole: model.RoleViewer,
	}

	tcs := map[string]struct {
		input  string
		output string
	}{
		"success mapped": {
			input:  "admin@example.com",
			output: model.RoleAdmin,
		},
		"success default": {
			input:  "user@example.com",
			output: model.RoleViewer,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			res := mapper.MapEmailToRole(tc.input)

			require.Equal(t, tc.output, res)
		})
	}
}

func TestGetUserRoles(t *testing.T) {
	roles := map[string]string{"admin@example.com": model.RoleAdmin}
	mapper := &RoleMapper{userRoles: roles, defaultRole: model.RoleViewer}

	tcs := map[string]struct {
		output map[string]string
	}{
		"success": {
			output: roles,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			res := mapper.GetUserRoles()

			require.Equal(t, tc.output, res)
		})
	}
}

func TestGetDefaultRole(t *testing.T) {
	mapper := &RoleMapper{defaultRole: model.RoleViewer}

	tcs := map[string]struct {
		output string
	}{
		"success": {
			output: model.RoleViewer,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			res := mapper.GetDefaultRole()

			require.Equal(t, tc.output, res)
		})
	}
}
