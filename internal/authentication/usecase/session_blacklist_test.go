package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identity-srv/internal/authentication"
)

func jsonBytesMatchedBy(t *testing.T, assert func(map[string]interface{})) interface{} {
	t.Helper()

	return testifymock.MatchedBy(func(v interface{}) bool {
		raw, ok := v.([]byte)
		if !ok {
			return false
		}

		var data map[string]interface{}
		if err := json.Unmarshal(raw, &data); err != nil {
			return false
		}

		assert(data)
		return true
	})
}

func TestCreateSession(t *testing.T) {
	// Coverage note: json.Marshal(SessionData) and json.Marshal([]string) error branches
	// are not practically triggerable because both values are fully marshalable concrete types.
	ctx := context.Background()
	redisErr := errors.New("redis error")
	existingJTIs, err := json.Marshal([]string{"old-jti"})
	require.NoError(t, err)

	type mockRedisSet struct {
		isCalled bool
		input    string
		value    interface{}
		err      error
	}

	type mockRedisGet struct {
		isCalled bool
		input    string
		output   string
		err      error
	}

	type mockRedisExists struct {
		isCalled bool
		input    string
		output   bool
		err      error
	}

	type mockRedis struct {
		setSession     mockRedisSet
		getSessions    mockRedisGet
		existsSession  mockRedisExists
		setUserSession mockRedisSet
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			userID     string
			jti        string
			rememberMe bool
		}
		mock mock
		err  error
	}{
		"success": {
			input: struct {
				userID     string
				jti        string
				rememberMe bool
			}{
				userID: "user-id",
				jti:    "jti",
			},
			mock: mock{redis: mockRedis{
				setSession: mockRedisSet{
					isCalled: true,
					input:    "session:jti",
				},
				getSessions: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					err:      redisErr,
				},
				setUserSession: mockRedisSet{
					isCalled: true,
					input:    "user_sessions:user-id",
				},
			}},
		},
		"success remember me filters expired existing sessions": {
			input: struct {
				userID     string
				jti        string
				rememberMe bool
			}{
				userID:     "user-id",
				jti:        "new-jti",
				rememberMe: true,
			},
			mock: mock{redis: mockRedis{
				setSession: mockRedisSet{
					isCalled: true,
					input:    "session:new-jti",
				},
				getSessions: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					output:   string(existingJTIs),
				},
				existsSession: mockRedisExists{
					isCalled: true,
					input:    "session:old-jti",
					output:   true,
				},
				setUserSession: mockRedisSet{
					isCalled: true,
					input:    "user_sessions:user-id",
				},
			}},
		},
		"success ignores invalid existing jtis": {
			input: struct {
				userID     string
				jti        string
				rememberMe bool
			}{
				userID: "user-id",
				jti:    "jti",
			},
			mock: mock{redis: mockRedis{
				setSession: mockRedisSet{
					isCalled: true,
					input:    "session:jti",
				},
				getSessions: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					output:   "{invalid",
				},
				setUserSession: mockRedisSet{
					isCalled: true,
					input:    "user_sessions:user-id",
				},
			}},
		},
		"err store session": {
			input: struct {
				userID     string
				jti        string
				rememberMe bool
			}{
				userID: "user-id",
				jti:    "jti",
			},
			mock: mock{redis: mockRedis{
				setSession: mockRedisSet{
					isCalled: true,
					input:    "session:jti",
					err:      redisErr,
				},
			}},
			err: authentication.ErrInternalSystem,
		},
		"err store user session mapping": {
			input: struct {
				userID     string
				jti        string
				rememberMe bool
			}{
				userID: "user-id",
				jti:    "jti",
			},
			mock: mock{redis: mockRedis{
				setSession: mockRedisSet{
					isCalled: true,
					input:    "session:jti",
				},
				getSessions: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					err:      redisErr,
				},
				setUserSession: mockRedisSet{
					isCalled: true,
					input:    "user_sessions:user-id",
					err:      redisErr,
				},
			}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			sm := uc.sessionManager

			if tc.mock.redis.setSession.isCalled {
				deps.redis.EXPECT().Set(
					ctx,
					tc.mock.redis.setSession.input,
					testifymock.Anything,
					testifymock.MatchedBy(func(ttl time.Duration) bool {
						if tc.input.rememberMe {
							return ttl == 7*24*time.Hour
						}
						return ttl == time.Hour
					}),
				).Return(tc.mock.redis.setSession.err)
			}
			if tc.mock.redis.getSessions.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.getSessions.input).
					Return(tc.mock.redis.getSessions.output, tc.mock.redis.getSessions.err)
			}
			if tc.mock.redis.existsSession.isCalled {
				deps.redis.EXPECT().Exists(ctx, tc.mock.redis.existsSession.input).
					Return(tc.mock.redis.existsSession.output, tc.mock.redis.existsSession.err)
			}
			if tc.mock.redis.setUserSession.isCalled {
				deps.redis.EXPECT().Set(
					ctx,
					tc.mock.redis.setUserSession.input,
					testifymock.Anything,
					7*24*time.Hour,
				).Return(tc.mock.redis.setUserSession.err)
			}

			err := sm.CreateSession(ctx, tc.input.userID, tc.input.jti, tc.input.rememberMe)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetSession(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")
	session := SessionData{UserID: "user-id", JTI: "jti", ExpiresAt: time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)}
	sessionJSON, err := json.Marshal(session)
	require.NoError(t, err)

	type mockRedisGet struct {
		isCalled bool
		input    string
		output   string
		err      error
	}

	type mockRedis struct {
		get mockRedisGet
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output *SessionData
		err    error
	}{
		"success": {
			input: "jti",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "session:jti",
				output:   string(sessionJSON),
			}}},
			output: &session,
		},
		"err redis Get": {
			input: "jti",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "session:jti",
				err:      redisErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
		"err invalid session json": {
			input: "jti",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "session:jti",
				output:   "{invalid",
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			sm := uc.sessionManager

			if tc.mock.redis.get.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.get.input).
					Return(tc.mock.redis.get.output, tc.mock.redis.get.err)
			}

			res, err := sm.GetSession(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

	type mockRedisDelete struct {
		isCalled bool
		input    string
		err      error
	}

	type mockRedis struct {
		delete mockRedisDelete
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input string
		mock  mock
		err   error
	}{
		"success": {
			input: "jti",
			mock: mock{redis: mockRedis{delete: mockRedisDelete{
				isCalled: true,
				input:    "session:jti",
			}}},
		},
		"err redis Delete": {
			input: "jti",
			mock: mock{redis: mockRedis{delete: mockRedisDelete{
				isCalled: true,
				input:    "session:jti",
				err:      redisErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			sm := uc.sessionManager

			if tc.mock.redis.delete.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.delete.input).
					Return(tc.mock.redis.delete.err)
			}

			err := sm.DeleteSession(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetAllUserSessions(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")
	jtisJSON, err := json.Marshal([]string{"jti-1", "jti-2"})
	require.NoError(t, err)

	type mockRedisGet struct {
		isCalled bool
		input    string
		output   string
		err      error
	}

	type mockRedis struct {
		get mockRedisGet
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output []string
		err    error
	}{
		"success": {
			input: "user-id",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "user_sessions:user-id",
				output:   string(jtisJSON),
			}}},
			output: []string{"jti-1", "jti-2"},
		},
		"success no sessions": {
			input: "user-id",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "user_sessions:user-id",
				err:      redisErr,
			}}},
			output: []string{},
		},
		"err invalid jtis json": {
			input: "user-id",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "user_sessions:user-id",
				output:   "{invalid",
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			sm := uc.sessionManager

			if tc.mock.redis.get.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.get.input).
					Return(tc.mock.redis.get.output, tc.mock.redis.get.err)
			}

			res, err := sm.GetAllUserSessions(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestDeleteUserSessions(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")
	jtisJSON, err := json.Marshal([]string{"jti-1", "jti-2"})
	require.NoError(t, err)

	type mockRedisGet struct {
		isCalled bool
		input    string
		output   string
		err      error
	}

	type mockRedisDelete struct {
		isCalled bool
		input    string
		err      error
	}

	type mockRedis struct {
		get           mockRedisGet
		deleteSession []mockRedisDelete
		deleteMapping mockRedisDelete
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input string
		mock  mock
		err   error
	}{
		"success": {
			input: "user-id",
			mock: mock{redis: mockRedis{
				get: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					output:   string(jtisJSON),
				},
				deleteSession: []mockRedisDelete{
					{isCalled: true, input: "session:jti-1"},
					{isCalled: true, input: "session:jti-2"},
				},
				deleteMapping: mockRedisDelete{
					isCalled: true,
					input:    "user_sessions:user-id",
				},
			}},
		},
		"success continues when deleting one session fails": {
			input: "user-id",
			mock: mock{redis: mockRedis{
				get: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					output:   string(jtisJSON),
				},
				deleteSession: []mockRedisDelete{
					{isCalled: true, input: "session:jti-1", err: redisErr},
					{isCalled: true, input: "session:jti-2"},
				},
				deleteMapping: mockRedisDelete{
					isCalled: true,
					input:    "user_sessions:user-id",
				},
			}},
		},
		"err get user sessions": {
			input: "user-id",
			mock: mock{redis: mockRedis{get: mockRedisGet{
				isCalled: true,
				input:    "user_sessions:user-id",
				output:   "{invalid",
			}}},
			err: authentication.ErrInternalSystem,
		},
		"err delete user sessions mapping": {
			input: "user-id",
			mock: mock{redis: mockRedis{
				get: mockRedisGet{
					isCalled: true,
					input:    "user_sessions:user-id",
					output:   string(jtisJSON),
				},
				deleteSession: []mockRedisDelete{
					{isCalled: true, input: "session:jti-1"},
					{isCalled: true, input: "session:jti-2"},
				},
				deleteMapping: mockRedisDelete{
					isCalled: true,
					input:    "user_sessions:user-id",
					err:      redisErr,
				},
			}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			sm := uc.sessionManager

			if tc.mock.redis.get.isCalled {
				deps.redis.EXPECT().Get(ctx, tc.mock.redis.get.input).
					Return(tc.mock.redis.get.output, tc.mock.redis.get.err)
			}
			for _, deleteSession := range tc.mock.redis.deleteSession {
				if deleteSession.isCalled {
					deps.redis.EXPECT().Delete(ctx, deleteSession.input).
						Return(deleteSession.err)
				}
			}
			if tc.mock.redis.deleteMapping.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.deleteMapping.input).
					Return(tc.mock.redis.deleteMapping.err)
			}

			err := sm.DeleteUserSessions(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSessionExists(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

	type mockRedisExists struct {
		isCalled bool
		input    string
		output   bool
		err      error
	}

	type mockRedis struct {
		exists mockRedisExists
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output bool
		err    error
	}{
		"success exists": {
			input: "jti",
			mock: mock{redis: mockRedis{exists: mockRedisExists{
				isCalled: true,
				input:    "session:jti",
				output:   true,
			}}},
			output: true,
		},
		"err redis Exists": {
			input: "jti",
			mock: mock{redis: mockRedis{exists: mockRedisExists{
				isCalled: true,
				input:    "session:jti",
				err:      redisErr,
			}}},
			err: redisErr,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)
			sm := uc.sessionManager

			if tc.mock.redis.exists.isCalled {
				deps.redis.EXPECT().Exists(ctx, tc.mock.redis.exists.input).
					Return(tc.mock.redis.exists.output, tc.mock.redis.exists.err)
			}

			res, err := sm.SessionExists(ctx, tc.input)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestAddToken(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

	type mockRedisSet struct {
		isCalled bool
		input    string
		value    interface{}
		err      error
	}

	type mockRedis struct {
		set mockRedisSet
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			jti       string
			expiresAt time.Time
		}
		mock mock
		err  error
	}{
		"success expired token no-op": {
			input: struct {
				jti       string
				expiresAt time.Time
			}{
				jti:       "jti",
				expiresAt: time.Now().Add(-time.Hour),
			},
		},
		"success": {
			input: struct {
				jti       string
				expiresAt time.Time
			}{
				jti:       "jti",
				expiresAt: time.Now().Add(time.Hour),
			},
			mock: mock{redis: mockRedis{set: mockRedisSet{
				isCalled: true,
				input:    "blacklist:jti",
				value:    "1",
			}}},
		},
		"err redis Set": {
			input: struct {
				jti       string
				expiresAt time.Time
			}{
				jti:       "jti",
				expiresAt: time.Now().Add(time.Hour),
			},
			mock: mock{redis: mockRedis{set: mockRedisSet{
				isCalled: true,
				input:    "blacklist:jti",
				value:    "1",
				err:      redisErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			_, deps := initUseCase(t)
			bm := NewBlacklistManager(deps.redis)

			if tc.mock.redis.set.isCalled {
				deps.redis.EXPECT().Set(ctx, tc.mock.redis.set.input, tc.mock.redis.set.value, anyPositiveDuration()).
					Return(tc.mock.redis.set.err)
			}

			err := bm.AddToken(ctx, tc.input.jti, tc.input.expiresAt)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAddAllUserTokens(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

	type mockRedisSet struct {
		isCalled bool
		input    string
		value    interface{}
		err      error
	}

	type mockRedis struct {
		set []mockRedisSet
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input struct {
			jtis      []string
			expiresAt time.Time
		}
		mock mock
		err  error
	}{
		"success expired tokens no-op": {
			input: struct {
				jtis      []string
				expiresAt time.Time
			}{
				jtis:      []string{"jti"},
				expiresAt: time.Now().Add(-time.Hour),
			},
		},
		"success": {
			input: struct {
				jtis      []string
				expiresAt time.Time
			}{
				jtis:      []string{"jti-1", "jti-2"},
				expiresAt: time.Now().Add(time.Hour),
			},
			mock: mock{redis: mockRedis{set: []mockRedisSet{
				{isCalled: true, input: "blacklist:jti-1", value: "1"},
				{isCalled: true, input: "blacklist:jti-2", value: "1"},
			}}},
		},
		"err redis Set": {
			input: struct {
				jtis      []string
				expiresAt time.Time
			}{
				jtis:      []string{"jti-1", "jti-2"},
				expiresAt: time.Now().Add(time.Hour),
			},
			mock: mock{redis: mockRedis{set: []mockRedisSet{
				{isCalled: true, input: "blacklist:jti-1", value: "1"},
				{isCalled: true, input: "blacklist:jti-2", value: "1", err: redisErr},
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			_, deps := initUseCase(t)
			bm := NewBlacklistManager(deps.redis)

			for _, set := range tc.mock.redis.set {
				if set.isCalled {
					deps.redis.EXPECT().Set(ctx, set.input, set.value, anyPositiveDuration()).
						Return(set.err)
				}
			}

			err := bm.AddAllUserTokens(ctx, tc.input.jtis, tc.input.expiresAt)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestIsBlacklisted(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

	type mockRedisExists struct {
		isCalled bool
		input    string
		output   bool
		err      error
	}

	type mockRedis struct {
		exists mockRedisExists
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output bool
		err    error
	}{
		"success blacklisted": {
			input: "jti",
			mock: mock{redis: mockRedis{exists: mockRedisExists{
				isCalled: true,
				input:    "blacklist:jti",
				output:   true,
			}}},
			output: true,
		},
		"err redis Exists": {
			input: "jti",
			mock: mock{redis: mockRedis{exists: mockRedisExists{
				isCalled: true,
				input:    "blacklist:jti",
				err:      redisErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			_, deps := initUseCase(t)
			bm := NewBlacklistManager(deps.redis)

			if tc.mock.redis.exists.isCalled {
				deps.redis.EXPECT().Exists(ctx, tc.mock.redis.exists.input).
					Return(tc.mock.redis.exists.output, tc.mock.redis.exists.err)
			}

			res, err := bm.IsBlacklisted(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestRemoveToken(t *testing.T) {
	ctx := context.Background()
	redisErr := errors.New("redis error")

	type mockRedisDelete struct {
		isCalled bool
		input    string
		err      error
	}

	type mockRedis struct {
		delete mockRedisDelete
	}

	type mock struct {
		redis mockRedis
	}

	tcs := map[string]struct {
		input string
		mock  mock
		err   error
	}{
		"success": {
			input: "jti",
			mock: mock{redis: mockRedis{delete: mockRedisDelete{
				isCalled: true,
				input:    "blacklist:jti",
			}}},
		},
		"err redis Delete": {
			input: "jti",
			mock: mock{redis: mockRedis{delete: mockRedisDelete{
				isCalled: true,
				input:    "blacklist:jti",
				err:      redisErr,
			}}},
			err: authentication.ErrInternalSystem,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			_, deps := initUseCase(t)
			bm := NewBlacklistManager(deps.redis)

			if tc.mock.redis.delete.isCalled {
				deps.redis.EXPECT().Delete(ctx, tc.mock.redis.delete.input).
					Return(tc.mock.redis.delete.err)
			}

			err := bm.RemoveToken(ctx, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
