package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	sharedauth "github.com/smap-hcmut/shared-libs/go/auth"
	sharedencrypter "github.com/smap-hcmut/shared-libs/go/encrypter"
	sharedlog "github.com/smap-hcmut/shared-libs/go/log"
	sharedredis "github.com/smap-hcmut/shared-libs/go/redis"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identity-srv/config"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"
	"identity-srv/internal/user"
	"identity-srv/pkg/oauth"
)

type mockDeps struct {
	userUC   *user.MockUseCase
	scope    *sharedauth.MockManager
	jwt      *sharedauth.MockManager
	encrypt  *sharedencrypter.MockEncrypter
	redis    *sharedredis.MockIRedis
	provider *oauth.MockProvider
}

func initUseCase(t *testing.T) (*ImplUsecase, mockDeps) {
	t.Helper()

	l := sharedlog.NewZapLogger(sharedlog.ZapConfig{
		Level:        sharedlog.LevelFatal,
		Mode:         sharedlog.ModeProduction,
		Encoding:     sharedlog.EncodingJSON,
		ColorEnabled: false,
	})
	scope := sharedauth.NewMockManager(t)
	encrypt := sharedencrypter.NewMockEncrypter(t)
	userUC := user.NewMockUseCase(t)
	jwtManager := sharedauth.NewMockManager(t)
	redisClient := sharedredis.NewMockIRedis(t)
	provider := oauth.NewMockProvider(t)

	uc := New(l, scope, encrypt, userUC)
	uc.SetJWTManager(jwtManager)
	uc.SetSessionManager(NewSessionManager(redisClient, time.Hour, l))
	uc.SetBlacklistManager(NewBlacklistManager(redisClient))
	uc.SetOAuthProvider(provider)
	uc.SetRoleMapper(&RoleMapper{
		userRoles: map[string]string{
			"admin@example.com": model.RoleAdmin,
		},
		defaultRole: model.RoleViewer,
	})
	uc.SetRedirectValidator(NewRedirectValidator([]string{"https://app.example.com"}))
	uc.SetAccessControl([]string{"example.com"}, []string{"blocked@example.com"})
	uc.clock = func() time.Time { return time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC) }

	return uc, mockDeps{
		userUC:   userUC,
		scope:    scope,
		jwt:      jwtManager,
		encrypt:  encrypt,
		redis:    redisClient,
		provider: provider,
	}
}

func anyPositiveDuration() interface{} {
	return mock.MatchedBy(func(ttl time.Duration) bool {
		return ttl > 0
	})
}

func TestNew(t *testing.T) {
	l := sharedlog.NewZapLogger(sharedlog.ZapConfig{
		Level:        sharedlog.LevelFatal,
		Mode:         sharedlog.ModeProduction,
		Encoding:     sharedlog.EncodingJSON,
		ColorEnabled: false,
	})
	scope := sharedauth.NewMockManager(t)
	encrypt := sharedencrypter.NewMockEncrypter(t)
	userUC := user.NewMockUseCase(t)

	tcs := map[string]struct {
		input struct {
			l       sharedlog.Logger
			scope   sharedauth.Manager
			encrypt sharedencrypter.Encrypter
			userUC  user.UseCase
		}
		output bool
	}{
		"success": {
			input: struct {
				l       sharedlog.Logger
				scope   sharedauth.Manager
				encrypt sharedencrypter.Encrypter
				userUC  user.UseCase
			}{
				l:       l,
				scope:   scope,
				encrypt: encrypt,
				userUC:  userUC,
			},
			output: true,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc := New(tc.input.l, tc.input.scope, tc.input.encrypt, tc.input.userUC)

			require.Equal(t, tc.output, uc != nil)
			require.Equal(t, tc.input.l, uc.l)
			require.Equal(t, tc.input.scope, uc.scope)
			require.Equal(t, tc.input.encrypt, uc.encrypt)
			require.Equal(t, tc.input.userUC, uc.userUC)
			require.NotNil(t, uc.clock)
		})
	}
}

func TestFactoriesAndSetters(t *testing.T) {
	l := sharedlog.NewZapLogger(sharedlog.ZapConfig{
		Level:        sharedlog.LevelFatal,
		Mode:         sharedlog.ModeProduction,
		Encoding:     sharedlog.EncodingJSON,
		ColorEnabled: false,
	})
	redisClient := sharedredis.NewMockIRedis(t)
	scope := sharedauth.NewMockManager(t)
	encrypt := sharedencrypter.NewMockEncrypter(t)
	userUC := user.NewMockUseCase(t)
	jwtManager := sharedauth.NewMockManager(t)
	provider := oauth.NewMockProvider(t)

	tcs := map[string]struct {
		input struct {
			ttl            time.Duration
			cfg            *config.Config
			allowedURLs    []string
			allowedDomains []string
			blockedEmails   []string
		}
		output struct {
			defaultRole string
			userRoles   map[string]string
		}
	}{
		"success": {
			input: struct {
				ttl            time.Duration
				cfg            *config.Config
				allowedURLs    []string
				allowedDomains []string
				blockedEmails   []string
			}{
				ttl: 30 * time.Minute,
				cfg: &config.Config{
					AccessControl: config.AccessControlConfig{
						UserRoles: map[string]string{
							"admin@example.com": model.RoleAdmin,
						},
						DefaultRole: model.RoleViewer,
					},
				},
				allowedURLs:    []string{"https://app.example.com"},
				allowedDomains: []string{"example.com"},
				blockedEmails:   []string{"blocked@example.com"},
			},
			output: struct {
				defaultRole string
				userRoles   map[string]string
			}{
				defaultRole: model.RoleViewer,
				userRoles: map[string]string{
					"admin@example.com": model.RoleAdmin,
				},
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc := New(l, scope, encrypt, userUC)
			sessionManager := NewSessionManager(redisClient, tc.input.ttl, l)
			blacklistManager := NewBlacklistManager(redisClient)
			roleMapper := NewRoleMapper(tc.input.cfg)
			redirectValidator := NewRedirectValidator(tc.input.allowedURLs)

			uc.SetSessionManager(sessionManager)
			uc.SetBlacklistManager(blacklistManager)
			uc.SetJWTManager(jwtManager)
			uc.SetRoleMapper(roleMapper)
			uc.SetOAuthProvider(provider)
			uc.SetRedirectValidator(redirectValidator)
			uc.SetAccessControl(tc.input.allowedDomains, tc.input.blockedEmails)

			require.Equal(t, sessionManager, uc.sessionManager)
			require.Equal(t, blacklistManager, uc.blacklistManager)
			require.Equal(t, jwtManager, uc.jwtManager)
			require.Equal(t, tc.output.defaultRole, uc.roleMapper.defaultRole)
			require.Equal(t, tc.output.userRoles, uc.roleMapper.userRoles)
			require.Equal(t, provider, uc.oauthProvider)
			require.Equal(t, redirectValidator, uc.redirectValidator)
			require.Equal(t, tc.input.allowedDomains, uc.allowedDomains)
			require.Equal(t, tc.input.blockedEmails, uc.blockedEmails)
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("user error")
	expectedUser := model.User{ID: "user-id", Email: "user@example.com"}

	type mockUserDetail struct {
		isCalled bool
		input    string
		output   model.User
		err      error
	}

	type mockUser struct {
		detail mockUserDetail
	}

	type mock struct {
		user mockUser
	}

	tcs := map[string]struct {
		input  model.Scope
		mock   mock
		output *model.User
		err    error
	}{
		"success": {
			input: model.Scope{UserID: "user-id"},
			mock: mock{user: mockUser{detail: mockUserDetail{
				isCalled: true,
				input:    "user-id",
				output:   expectedUser,
			}}},
			output: &expectedUser,
		},
		"err user Detail": {
			input: model.Scope{UserID: "user-id"},
			mock: mock{user: mockUser{detail: mockUserDetail{
				isCalled: true,
				input:    "user-id",
				err:      expectedErr,
			}}},
			err: authentication.ErrUserNotFound,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.user.detail.isCalled {
				deps.userUC.EXPECT().Detail(ctx, tc.mock.user.detail.input).
					Return(tc.mock.user.detail.output, tc.mock.user.detail.err)
			}

			res, err := uc.GetCurrentUser(ctx, tc.input)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("user error")
	expectedUser := model.User{ID: "user-id", Email: "user@example.com"}

	type mockUserDetail struct {
		isCalled bool
		input    string
		output   model.User
		err      error
	}

	type mockUser struct {
		detail mockUserDetail
	}

	type mock struct {
		user mockUser
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output *model.User
		err    error
	}{
		"success": {
			input: "user-id",
			mock: mock{user: mockUser{detail: mockUserDetail{
				isCalled: true,
				input:    "user-id",
				output:   expectedUser,
			}}},
			output: &expectedUser,
		},
		"err user Detail": {
			input: "user-id",
			mock: mock{user: mockUser{detail: mockUserDetail{
				isCalled: true,
				input:    "user-id",
				err:      expectedErr,
			}}},
			err: authentication.ErrUserNotFound,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.user.detail.isCalled {
				deps.userUC.EXPECT().Detail(ctx, tc.mock.user.detail.input).
					Return(tc.mock.user.detail.output, tc.mock.user.detail.err)
			}

			res, err := uc.GetUserByID(ctx, tc.input)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestLogout(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("redis error")

	type mockRedisDelete struct {
		isCalled bool
		input    []string
		err      error
	}

	type mockRedis struct {
		delete mockRedisDelete
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			scope             model.Scope
			useSessionManager bool
		}
		mock mock
		err  error
	}{
		"success without session manager": {
			input: struct {
				scope             model.Scope
				useSessionManager bool
			}{
				scope: model.Scope{JTI: "jti"},
			},
		},
		"success": {
			input: struct {
				scope             model.Scope
				useSessionManager bool
			}{
				scope:             model.Scope{JTI: "jti"},
				useSessionManager: true,
			},
			mock: mock{redis: mockRedis{delete: mockRedisDelete{
				isCalled: true,
				input:    []string{"session:jti"},
			}}},
		},
		"err session DeleteSession": {
			input: struct {
				scope             model.Scope
				useSessionManager bool
			}{
				scope:             model.Scope{JTI: "jti"},
				useSessionManager: true,
			},
			mock: mock{redis: mockRedis{delete: mockRedisDelete{
				isCalled: true,
				input:    []string{"session:jti"},
				err:      expectedErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			if !tc.input.useSessionManager {
				uc.SetSessionManager(nil)
			}

			if tc.mock.redis.delete.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.delete.input[0]).
					Return(tc.mock.redis.delete.err)
			}

			err := uc.Logout(ctx, tc.input.scope)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("jwt error")
	redisErr := errors.New("redis error")
	expiresAt := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	type mockJWTVerify struct {
		isCalled bool
		input    string
		output   sharedauth.Payload
		err      error
	}

	type mockRedisExists struct {
		isCalled bool
		input    string
		output   bool
		err      error
	}

	type mockJWT struct {
		verify mockJWTVerify
	}

	type mockRedis struct {
		exists mockRedisExists
	}

	type mock struct {
		jwt   mockJWT
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			token               string
			useJWTManager       bool
			useBlacklistManager bool
		}
		mock   mock
		output *authentication.TokenValidationResult
		err    error
	}{
		"err jwt manager missing": {
			input: struct {
				token               string
				useJWTManager       bool
				useBlacklistManager bool
			}{
				token:               "token",
				useBlacklistManager: true,
			},
			err: errors.New("jwt manager not configured"),
		},
		"invalid token": {
			input: struct {
				token               string
				useJWTManager       bool
				useBlacklistManager bool
			}{
				token:               "token",
				useJWTManager:       true,
				useBlacklistManager: true,
			},
			mock: mock{jwt: mockJWT{verify: mockJWTVerify{
				isCalled: true,
				input:    "token",
				err:      expectedErr,
			}}},
			output: &authentication.TokenValidationResult{Valid: false},
		},
		"err blacklist check": {
			input: struct {
				token               string
				useJWTManager       bool
				useBlacklistManager bool
			}{
				token:               "token",
				useJWTManager:       true,
				useBlacklistManager: true,
			},
			mock: mock{
				jwt: mockJWT{verify: mockJWTVerify{
					isCalled: true,
					input:    "token",
					output: sharedauth.Payload{
						UserID:   "user-id",
						Username: "user@example.com",
						Role:     model.RoleAdmin,
						StandardClaims: jwt.StandardClaims{
							Id:        "jti",
							ExpiresAt: expiresAt.Unix(),
						},
					},
				}},
				redis: mockRedis{exists: mockRedisExists{
					isCalled: true,
					input:    "blacklist:jti",
					err:      redisErr,
				}},
			},
			err: authentication.ErrInternalSystem,
		},
		"blacklisted token": {
			input: struct {
				token               string
				useJWTManager       bool
				useBlacklistManager bool
			}{
				token:               "token",
				useJWTManager:       true,
				useBlacklistManager: true,
			},
			mock: mock{
				jwt: mockJWT{verify: mockJWTVerify{
					isCalled: true,
					input:    "token",
					output: sharedauth.Payload{
						StandardClaims: jwt.StandardClaims{Id: "jti"},
					},
				}},
				redis: mockRedis{exists: mockRedisExists{
					isCalled: true,
					input:    "blacklist:jti",
					output:   true,
				}},
			},
			output: &authentication.TokenValidationResult{Valid: false},
		},
		"success with blacklist manager": {
			input: struct {
				token               string
				useJWTManager       bool
				useBlacklistManager bool
			}{
				token:               "token",
				useJWTManager:       true,
				useBlacklistManager: true,
			},
			mock: mock{
				jwt: mockJWT{verify: mockJWTVerify{
					isCalled: true,
					input:    "token",
					output: sharedauth.Payload{
						UserID:   "user-id",
						Username: "user@example.com",
						Role:     model.RoleAdmin,
						StandardClaims: jwt.StandardClaims{
							Id:        "jti",
							ExpiresAt: expiresAt.Unix(),
						},
					},
				}},
				redis: mockRedis{exists: mockRedisExists{
					isCalled: true,
					input:    "blacklist:jti",
					output:   false,
				}},
			},
			output: &authentication.TokenValidationResult{
				Valid:     true,
				UserID:    "user-id",
				Email:     "user@example.com",
				Role:      model.RoleAdmin,
				Groups:    []string{},
				ExpiresAt: time.Unix(expiresAt.Unix(), 0),
			},
		},
		"success without blacklist manager": {
			input: struct {
				token               string
				useJWTManager       bool
				useBlacklistManager bool
			}{
				token:         "token",
				useJWTManager: true,
			},
			mock: mock{
				jwt: mockJWT{verify: mockJWTVerify{
					isCalled: true,
					input:    "token",
					output: sharedauth.Payload{
						UserID:   "user-id",
						Username: "user@example.com",
						Role:     model.RoleViewer,
						StandardClaims: jwt.StandardClaims{
							Id:        "jti",
							ExpiresAt: expiresAt.Unix(),
						},
					},
				}},
			},
			output: &authentication.TokenValidationResult{
				Valid:     true,
				UserID:    "user-id",
				Email:     "user@example.com",
				Role:      model.RoleViewer,
				Groups:    []string{},
				ExpiresAt: time.Unix(expiresAt.Unix(), 0),
			},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			if !tc.input.useJWTManager {
				uc.SetJWTManager(nil)
			}
			if !tc.input.useBlacklistManager {
				uc.SetBlacklistManager(nil)
			}

			if tc.mock.jwt.verify.isCalled {
				deps.jwt.EXPECT().Verify(tc.mock.jwt.verify.input).
					Return(tc.mock.jwt.verify.output, tc.mock.jwt.verify.err)
			}
			if tc.mock.redis.exists.isCalled {
				deps.redis.EXPECT().Exists(ctx, tc.mock.redis.exists.input).
					Return(tc.mock.redis.exists.output, tc.mock.redis.exists.err)
			}

			res, err := uc.ValidateToken(ctx, tc.input.token)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestRevokeToken(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")
	expiresAt := time.Now().Add(time.Hour)
	sessionData := SessionData{UserID: "user-id", JTI: "jti", ExpiresAt: expiresAt}
	sessionJSON, err := json.Marshal(sessionData)
	require.NoError(t, err)

	type mockRedisGet struct {
		isCalled bool
		input    string
		output   string
		err      error
	}

	type mockRedisSet struct {
		isCalled bool
		input    string
		value    interface{}
		err      error
	}

	type mockRedisDelete struct {
		isCalled bool
		input    []string
		err      error
	}

	type mockRedis struct {
		get    mockRedisGet
		set    mockRedisSet
		delete mockRedisDelete
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			jti                 string
			useSessionManager   bool
			useBlacklistManager bool
		}
		mock mock
		err  error
	}{
		"err configuration missing": {
			input: struct {
				jti                 string
				useSessionManager   bool
				useBlacklistManager bool
			}{
				jti: "jti",
			},
			err: authentication.ErrConfigurationMissing,
		},
		"err GetSession": {
			input: struct {
				jti                 string
				useSessionManager   bool
				useBlacklistManager bool
			}{
				jti:                 "jti",
				useSessionManager:   true,
				useBlacklistManager: true,
			},
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "session:jti",
				err:      redisErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
		"err AddToken": {
			input: struct {
				jti                 string
				useSessionManager   bool
				useBlacklistManager bool
			}{
				jti:                 "jti",
				useSessionManager:   true,
				useBlacklistManager: true,
			},
			mock: mock{redis: mockRedis{
				get: mockRedisGet{
					isCalled: true,
					input:    "session:jti",
					output:   string(sessionJSON),
				},
				set: mockRedisSet{
					isCalled: true,
					input:    "blacklist:jti",
					value:    "1",
					err:      redisErr,
				},
			}},
			err: authentication.ErrInternalSystem,
		},
		"err DeleteSession": {
			input: struct {
				jti                 string
				useSessionManager   bool
				useBlacklistManager bool
			}{
				jti:                 "jti",
				useSessionManager:   true,
				useBlacklistManager: true,
			},
			mock: mock{redis: mockRedis{
				get: mockRedisGet{
					isCalled: true,
					input:    "session:jti",
					output:   string(sessionJSON),
				},
				set: mockRedisSet{
					isCalled: true,
					input:    "blacklist:jti",
					value:    "1",
				},
				delete: mockRedisDelete{
					isCalled: true,
					input:    []string{"session:jti"},
					err:      redisErr,
				},
			}},
			err: authentication.ErrInternalSystem,
		},
		"success": {
			input: struct {
				jti                 string
				useSessionManager   bool
				useBlacklistManager bool
			}{
				jti:                 "jti",
				useSessionManager:   true,
				useBlacklistManager: true,
			},
			mock: mock{redis: mockRedis{
				get: mockRedisGet{
					isCalled: true,
					input:    "session:jti",
					output:   string(sessionJSON),
				},
				set: mockRedisSet{
					isCalled: true,
					input:    "blacklist:jti",
					value:    "1",
				},
				delete: mockRedisDelete{
					isCalled: true,
					input:    []string{"session:jti"},
				},
			}},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			if !tc.input.useSessionManager {
				uc.SetSessionManager(nil)
			}
			if !tc.input.useBlacklistManager {
				uc.SetBlacklistManager(nil)
			}

			if tc.mock.redis.get.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.get.input).
					Return(tc.mock.redis.get.output, tc.mock.redis.get.err)
			}
			if tc.mock.redis.set.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.set.input, tc.mock.redis.set.value, anyPositiveDuration()).
					Return(tc.mock.redis.set.err)
			}
			if tc.mock.redis.delete.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.delete.input[0]).
					Return(tc.mock.redis.delete.err)
			}

			err := uc.RevokeToken(ctx, tc.input.jti)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
