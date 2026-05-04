package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identity-srv/internal/authentication"
	"identity-srv/internal/model"
	"identity-srv/internal/user"
)

func TestIsAllowedDomain(t *testing.T) {
	tcs := map[string]struct {
		input struct {
			allowedDomains []string
			email          string
		}
		output bool
	}{
		"success no restrictions": {
			input: struct {
				allowedDomains []string
				email          string
			}{email: "user@any.com"},
			output: true,
		},
		"success allowed": {
			input: struct {
				allowedDomains []string
				email          string
			}{allowedDomains: []string{"example.com"}, email: "user@example.com"},
			output: true,
		},
		"success not allowed": {
			input: struct {
				allowedDomains []string
				email          string
			}{allowedDomains: []string{"example.com"}, email: "user@other.com"},
			output: false,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, _ := initUseCase(t)
			uc.allowedDomains = tc.input.allowedDomains

			res := uc.isAllowedDomain(tc.input.email)

			require.Equal(t, tc.output, res)
		})
	}
}

func TestIsBlockedEmail(t *testing.T) {
	tcs := map[string]struct {
		input struct {
			blockedEmails []string
			email         string
		}
		output bool
	}{
		"success blocked": {
			input: struct {
				blockedEmails []string
				email         string
			}{blockedEmails: []string{"blocked@example.com"}, email: "blocked@example.com"},
			output: true,
		},
		"success not blocked": {
			input: struct {
				blockedEmails []string
				email         string
			}{blockedEmails: []string{"blocked@example.com"}, email: "user@example.com"},
			output: false,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, _ := initUseCase(t)
			uc.blockedEmails = tc.input.blockedEmails

			res := uc.isBlockedEmail(tc.input.email)

			require.Equal(t, tc.output, res)
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tcs := map[string]struct {
		input  string
		output string
	}{
		"success": {
			input:  "user@example.com",
			output: "example.com",
		},
		"success no domain": {
			input:  "invalid",
			output: "",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, _ := initUseCase(t)

			res := uc.extractDomain(tc.input)

			require.Equal(t, tc.output, res)
		})
	}
}

func TestCreateOrUpdateUser(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("user error")
	expectedUser := model.User{ID: "user-id", Email: "user@example.com"}

	type mockUserCreate struct {
		isCalled bool
		input    user.CreateInput
		output   model.User
		err      error
	}

	type mockUser struct {
		create mockUserCreate
	}

	type mockDepsCase struct {
		user mockUser
	}

	tcs := map[string]struct {
		input struct {
			email     string
			name      string
			avatarURL string
		}
		mock   mockDepsCase
		output *model.User
		err    error
	}{
		"success": {
			input: struct {
				email     string
				name      string
				avatarURL string
			}{email: "user@example.com", name: "User", avatarURL: "avatar"},
			mock: mockDepsCase{user: mockUser{create: mockUserCreate{
				isCalled: true,
				input:    user.CreateInput{Email: "user@example.com", Name: "User", AvatarURL: "avatar"},
				output:   expectedUser,
			}}},
			output: &expectedUser,
		},
		"err user Create": {
			input: struct {
				email     string
				name      string
				avatarURL string
			}{email: "user@example.com", name: "User", avatarURL: "avatar"},
			mock: mockDepsCase{user: mockUser{create: mockUserCreate{
				isCalled: true,
				input:    user.CreateInput{Email: "user@example.com", Name: "User", AvatarURL: "avatar"},
				err:      expectedErr,
			}}},
			err: authentication.ErrUserCreation,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.user.create.isCalled {
				deps.userUC.EXPECT().Create(ctx, tc.mock.user.create.input).
					Return(tc.mock.user.create.output, tc.mock.user.create.err)
			}

			res, err := uc.createOrUpdateUser(ctx, tc.input.email, tc.input.name, tc.input.avatarURL)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestUpdateUserRole(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("user error")

	type mockUserUpdate struct {
		isCalled bool
		input    user.UpdateInput
		err      error
	}

	type mockUser struct {
		update mockUserUpdate
	}

	type mockDepsCase struct {
		user mockUser
	}

	tcs := map[string]struct {
		input struct {
			userID string
			role   string
		}
		mock mockDepsCase
		err  error
	}{
		"success": {
			input: struct {
				userID string
				role   string
			}{userID: "user-id", role: model.RoleAdmin},
			mock: mockDepsCase{user: mockUser{update: mockUserUpdate{
				isCalled: true,
				input:    user.UpdateInput{UserID: "user-id", Role: model.RoleAdmin},
			}}},
		},
		"err user Update": {
			input: struct {
				userID string
				role   string
			}{userID: "user-id", role: model.RoleAdmin},
			mock: mockDepsCase{user: mockUser{update: mockUserUpdate{
				isCalled: true,
				input:    user.UpdateInput{UserID: "user-id", Role: model.RoleAdmin},
				err:      expectedErr,
			}}},
			err: expectedErr,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.user.update.isCalled {
				deps.userUC.EXPECT().Update(ctx, tc.mock.user.update.input).
					Return(tc.mock.user.update.err)
			}

			err := uc.updateUserRole(ctx, tc.input.userID, tc.input.role)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestImplUsecaseMapEmailToRole(t *testing.T) {
	tcs := map[string]struct {
		input struct {
			email         string
			useRoleMapper bool
		}
		output string
	}{
		"success default without mapper": {
			input: struct {
				email         string
				useRoleMapper bool
			}{email: "user@example.com"},
			output: model.RoleViewer,
		},
		"success mapped": {
			input: struct {
				email         string
				useRoleMapper bool
			}{email: "admin@example.com", useRoleMapper: true},
			output: model.RoleAdmin,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, _ := initUseCase(t)
			if !tc.input.useRoleMapper {
				uc.SetRoleMapper(nil)
			}

			res := uc.mapEmailToRole(tc.input.email)

			require.Equal(t, tc.output, res)
		})
	}
}

func TestGenerateToken(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("jwt error")
	expectedUser := &model.User{ID: "user-id", Email: "user@example.com"}

	type mockJWTCreateToken struct {
		isCalled bool
		output   string
		err      error
	}

	type sharedAuthPayloadFields struct {
		id string
	}

	type mockJWTVerify struct {
		isCalled bool
		input    string
		output   sharedAuthPayloadFields
		err      error
	}

	type mockJWT struct {
		createToken mockJWTCreateToken
		verify      mockJWTVerify
	}

	type mockDepsCase struct {
		jwt mockJWT
	}

	tcs := map[string]struct {
		input struct {
			user          *model.User
			role          string
			groups        []string
			useJWTManager bool
		}
		mock  mockDepsCase
		token string
		jti   string
		err   error
	}{
		"err jwt manager missing": {
			input: struct {
				user          *model.User
				role          string
				groups        []string
				useJWTManager bool
			}{user: expectedUser, role: model.RoleAdmin},
			err: errors.New("jwt manager not configured"),
		},
		"err create token": {
			input: struct {
				user          *model.User
				role          string
				groups        []string
				useJWTManager bool
			}{user: expectedUser, role: model.RoleAdmin, useJWTManager: true},
			mock: mockDepsCase{jwt: mockJWT{createToken: mockJWTCreateToken{
				isCalled: true,
				err:      expectedErr,
			}}},
			err: expectedErr,
		},
		"err verify token": {
			input: struct {
				user          *model.User
				role          string
				groups        []string
				useJWTManager bool
			}{user: expectedUser, role: model.RoleAdmin, useJWTManager: true},
			mock: mockDepsCase{jwt: mockJWT{
				createToken: mockJWTCreateToken{isCalled: true, output: "token"},
				verify:      mockJWTVerify{isCalled: true, input: "token", err: expectedErr},
			}},
			err: expectedErr,
		},
		"success": {
			input: struct {
				user          *model.User
				role          string
				groups        []string
				useJWTManager bool
			}{user: expectedUser, role: model.RoleAdmin, groups: []string{}, useJWTManager: true},
			mock: mockDepsCase{jwt: mockJWT{
				createToken: mockJWTCreateToken{isCalled: true, output: "token"},
				verify: mockJWTVerify{
					isCalled: true,
					input:    "token",
					output:   sharedAuthPayloadFields{id: "jti"},
				},
			}},
			token: "token",
			jti:   "jti",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			if !tc.input.useJWTManager {
				uc.SetJWTManager(nil)
			}

			if tc.mock.jwt.createToken.isCalled {
				deps.jwt.EXPECT().CreateToken(mock.Anything).
					Return(tc.mock.jwt.createToken.output, tc.mock.jwt.createToken.err)
			}
			if tc.mock.jwt.verify.isCalled {
				deps.jwt.EXPECT().Verify(tc.mock.jwt.verify.input).
					Return(sharedAuthPayload(jwt.StandardClaims{Id: tc.mock.jwt.verify.output.id}), tc.mock.jwt.verify.err)
			}

			token, jti, err := uc.generateToken(ctx, tc.input.user, tc.input.role, tc.input.groups)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.token, token)
			require.Equal(t, tc.jti, jti)
		})
	}
}

func TestCreateSessionHelper(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

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

	type mockRedis struct {
		setSession     mockRedisSet
		getSessions    mockRedisGet
		setUserSession mockRedisSet
	}

	type mockDepsCase struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			userID            string
			jti               string
			rememberMe        bool
			useSessionManager bool
		}
		mock mockDepsCase
		err  error
	}{
		"success without session manager": {
			input: struct {
				userID            string
				jti               string
				rememberMe        bool
				useSessionManager bool
			}{userID: "user-id", jti: "jti"},
		},
		"success": {
			input: struct {
				userID            string
				jti               string
				rememberMe        bool
				useSessionManager bool
			}{userID: "user-id", jti: "jti", useSessionManager: true},
			mock: mockDepsCase{redis: mockRedis{
				setSession:     mockRedisSet{isCalled: true, input: "session:jti"},
				getSessions:    mockRedisGet{isCalled: true, input: "user_sessions:user-id", err: redisErr},
				setUserSession: mockRedisSet{isCalled: true, input: "user_sessions:user-id"},
			}},
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			if !tc.input.useSessionManager {
				uc.SetSessionManager(nil)
			}

			if tc.mock.redis.setSession.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.setSession.input, mock.Anything, mock.Anything).
					Return(tc.mock.redis.setSession.err)
			}
			if tc.mock.redis.getSessions.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.getSessions.input).
					Return("", tc.mock.redis.getSessions.err)
			}
			if tc.mock.redis.setUserSession.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.setUserSession.input, mock.Anything, 7*24*time.Hour).
					Return(tc.mock.redis.setUserSession.err)
			}

			err := uc.createSession(ctx, tc.input.userID, tc.input.jti, tc.input.rememberMe)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRevokeAllUserTokens(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")
	jtisJSON, err := json.Marshal([]string{"jti-1"})
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
		err      error
	}

	type mockRedisDelete struct {
		isCalled bool
		input    string
		err      error
	}

	type mockRedis struct {
		getForBlacklist mockRedisGet
		setBlacklist    mockRedisSet
		getForDelete    mockRedisGet
		deleteSession   mockRedisDelete
		deleteMapping   mockRedisDelete
	}

	type mockDepsCase struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			userID              string
			useSessionManager   bool
			useBlacklistManager bool
		}
		mock mockDepsCase
		err  error
	}{
		"err configuration missing": {
			input: struct {
				userID              string
				useSessionManager   bool
				useBlacklistManager bool
			}{userID: "user-id"},
			err: authentication.ErrConfigurationMissing,
		},
		"err get all sessions": {
			input: struct {
				userID              string
				useSessionManager   bool
				useBlacklistManager bool
			}{userID: "user-id", useSessionManager: true, useBlacklistManager: true},
			mock: mockDepsCase{redis: mockRedis{getForBlacklist: mockRedisGet{
				isCalled: true,
				input:    "user_sessions:user-id",
				output:   "{invalid",
			}}},
			err: authentication.ErrInternalSystem,
		},
		"err add all tokens": {
			input: struct {
				userID              string
				useSessionManager   bool
				useBlacklistManager bool
			}{userID: "user-id", useSessionManager: true, useBlacklistManager: true},
			mock: mockDepsCase{redis: mockRedis{
				getForBlacklist: mockRedisGet{isCalled: true, input: "user_sessions:user-id", output: string(jtisJSON)},
				setBlacklist:    mockRedisSet{isCalled: true, input: "blacklist:jti-1", err: redisErr},
			}},
			err: authentication.ErrInternalSystem,
		},
		"err delete user sessions": {
			input: struct {
				userID              string
				useSessionManager   bool
				useBlacklistManager bool
			}{userID: "user-id", useSessionManager: true, useBlacklistManager: true},
			mock: mockDepsCase{redis: mockRedis{
				getForBlacklist: mockRedisGet{isCalled: true, input: "user_sessions:user-id", output: string(jtisJSON)},
				setBlacklist:    mockRedisSet{isCalled: true, input: "blacklist:jti-1"},
				getForDelete:    mockRedisGet{isCalled: true, input: "user_sessions:user-id", output: string(jtisJSON)},
				deleteSession:   mockRedisDelete{isCalled: true, input: "session:jti-1"},
				deleteMapping:   mockRedisDelete{isCalled: true, input: "user_sessions:user-id", err: redisErr},
			}},
			err: authentication.ErrInternalSystem,
		},
		"success": {
			input: struct {
				userID              string
				useSessionManager   bool
				useBlacklistManager bool
			}{userID: "user-id", useSessionManager: true, useBlacklistManager: true},
			mock: mockDepsCase{redis: mockRedis{
				getForBlacklist: mockRedisGet{isCalled: true, input: "user_sessions:user-id", output: string(jtisJSON)},
				setBlacklist:    mockRedisSet{isCalled: true, input: "blacklist:jti-1"},
				getForDelete:    mockRedisGet{isCalled: true, input: "user_sessions:user-id", output: string(jtisJSON)},
				deleteSession:   mockRedisDelete{isCalled: true, input: "session:jti-1"},
				deleteMapping:   mockRedisDelete{isCalled: true, input: "user_sessions:user-id"},
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

			if tc.mock.redis.getForBlacklist.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.getForBlacklist.input).
					Return(tc.mock.redis.getForBlacklist.output, tc.mock.redis.getForBlacklist.err).Once()
			}
			if tc.mock.redis.setBlacklist.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.setBlacklist.input, "1", anyPositiveDuration()).
					Return(tc.mock.redis.setBlacklist.err)
			}
			if tc.mock.redis.getForDelete.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.getForDelete.input).
					Return(tc.mock.redis.getForDelete.output, tc.mock.redis.getForDelete.err).Once()
			}
			if tc.mock.redis.deleteSession.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.deleteSession.input).
					Return(tc.mock.redis.deleteSession.err)
			}
			if tc.mock.redis.deleteMapping.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.deleteMapping.input).
					Return(tc.mock.redis.deleteMapping.err)
			}

			err := uc.RevokeAllUserTokens(ctx, tc.input.userID)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
