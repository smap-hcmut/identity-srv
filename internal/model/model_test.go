package model

import (
	"errors"
	"testing"
	"time"

	"identity-srv/internal/sqlboiler"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scopeGetter struct {
	userID   string
	username string
	role     string
	jti      string
}

func (s scopeGetter) GetUserID() string   { return s.userID }
func (s scopeGetter) GetUsername() string { return s.username }
func (s scopeGetter) GetRole() string     { return s.role }
func (s scopeGetter) GetJTI() string      { return s.jti }

func TestNewUser(t *testing.T) {
	type input struct {
		email string
		name  string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output User
		err    error
	}{
		"success": {
			input: input{email: "user@example.com", name: "User Name"},
			output: User{
				Email:    "user@example.com",
				Name:     strPtr("User Name"),
				IsActive: true,
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			user := NewUser(tc.input.email, tc.input.name)

			require.NotNil(t, user)
			assert.Equal(t, tc.output.Email, user.Email)
			assert.Equal(t, tc.output.Name, user.Name)
			assert.Equal(t, tc.output.IsActive, user.IsActive)
			assert.True(t, user.IsViewer())
			assert.NoError(t, tc.err)
		})
	}
}

func TestUpdateLastLogin(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  User
		mock   mock
		output bool
		err    error
	}{
		"success": {
			output: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			before := time.Now()
			user := tc.input

			user.UpdateLastLogin()

			require.NotNil(t, user.LastLoginAt)
			assert.Equal(t, tc.output, !user.LastLoginAt.Before(before))
			assert.NoError(t, tc.err)
		})
	}
}

func TestNewUserFromDB(t *testing.T) {
	now := time.Now()
	type input struct {
		user *sqlboiler.User
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output *User
		err    error
	}{
		"nil_db_user": {
			input: input{user: nil},
		},
		"success_with_nullable_fields": {
			input: input{user: &sqlboiler.User{
				ID:          "user-1",
				Email:       "user@example.com",
				Name:        null.StringFrom("User Name"),
				AvatarURL:   null.StringFrom("https://example.com/avatar.png"),
				RoleHash:    "role-hash",
				IsActive:    null.BoolFrom(false),
				LastLoginAt: null.TimeFrom(now),
				CreatedAt:   now,
				UpdatedAt:   now,
			}},
			output: &User{
				ID:          "user-1",
				Email:       "user@example.com",
				Name:        strPtr("User Name"),
				AvatarURL:   strPtr("https://example.com/avatar.png"),
				RoleHash:    strPtr("role-hash"),
				IsActive:    false,
				LastLoginAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		"success_defaults_active_when_null": {
			input: input{user: &sqlboiler.User{
				ID:        "user-2",
				Email:     "viewer@example.com",
				RoleHash:  "role-hash",
				CreatedAt: now,
				UpdatedAt: now,
			}},
			output: &User{
				ID:        "user-2",
				Email:     "viewer@example.com",
				RoleHash:  strPtr("role-hash"),
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := NewUserFromDB(tc.input.user)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestToDBUser(t *testing.T) {
	now := time.Now()
	type mock struct{}

	tcs := map[string]struct {
		input  User
		mock   mock
		output sqlboiler.User
		err    error
	}{
		"success_with_nullable_fields": {
			input: User{
				ID:          "user-1",
				Email:       "user@example.com",
				Name:        strPtr("User Name"),
				AvatarURL:   strPtr("https://example.com/avatar.png"),
				RoleHash:    strPtr("role-hash"),
				IsActive:    true,
				LastLoginAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			output: sqlboiler.User{
				ID:          "user-1",
				Email:       "user@example.com",
				Name:        null.StringFrom("User Name"),
				AvatarURL:   null.StringFrom("https://example.com/avatar.png"),
				RoleHash:    "role-hash",
				IsActive:    null.BoolFrom(true),
				LastLoginAt: null.TimeFrom(now),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		"success_without_nullable_fields": {
			input: User{
				ID:        "user-2",
				Email:     "viewer@example.com",
				IsActive:  false,
				CreatedAt: now,
				UpdatedAt: now,
			},
			output: sqlboiler.User{
				ID:        "user-2",
				Email:     "viewer@example.com",
				IsActive:  null.BoolFrom(false),
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.ToDBUser()

			assert.Equal(t, &tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestEncryptRole(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output string
		err    error
	}{
		"success_admin": {
			input: RoleAdmin,
		},
		"success_analyst": {
			input: RoleAnalyst,
		},
		"success_viewer": {
			input: RoleViewer,
		},
		"invalid_role": {
			input: "OWNER",
			err:   ErrInvalidRole,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output, err := EncryptRole(tc.input)

			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				assert.Equal(t, tc.output, output)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, output)
		})
	}
}

func TestVerifyRole(t *testing.T) {
	adminHash, err := EncryptRole(RoleAdmin)
	require.NoError(t, err)
	type input struct {
		roleHash  string
		plainRole string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output bool
		err    error
	}{
		"success_match": {
			input:  input{roleHash: adminHash, plainRole: RoleAdmin},
			output: true,
		},
		"role_hash_mismatch": {
			input: input{roleHash: adminHash, plainRole: RoleViewer},
		},
		"invalid_plain_role": {
			input: input{roleHash: adminHash, plainRole: "OWNER"},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := VerifyRole(tc.input.roleHash, tc.input.plainRole)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestUserIsAdmin(t *testing.T) {
	adminHash := mustEncryptRole(t, RoleAdmin)
	type mock struct{}

	tcs := map[string]struct {
		input  User
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  User{RoleHash: &adminHash},
			output: true,
		},
		"nil_role_hash": {},
		"different_role": {
			input: User{RoleHash: strPtr(mustEncryptRole(t, RoleViewer))},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.IsAdmin()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestIsAnalyst(t *testing.T) {
	analystHash := mustEncryptRole(t, RoleAnalyst)
	type mock struct{}

	tcs := map[string]struct {
		input  User
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  User{RoleHash: &analystHash},
			output: true,
		},
		"nil_role_hash": {},
		"different_role": {
			input: User{RoleHash: strPtr(mustEncryptRole(t, RoleViewer))},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.IsAnalyst()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestIsViewer(t *testing.T) {
	viewerHash := mustEncryptRole(t, RoleViewer)
	type mock struct{}

	tcs := map[string]struct {
		input  User
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  User{RoleHash: &viewerHash},
			output: true,
		},
		"nil_role_hash": {},
		"different_role": {
			input: User{RoleHash: strPtr(mustEncryptRole(t, RoleAdmin))},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.IsViewer()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestGetRole(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  User
		mock   mock
		output string
		err    error
	}{
		"admin": {
			input:  User{RoleHash: strPtr(mustEncryptRole(t, RoleAdmin))},
			output: RoleAdmin,
		},
		"analyst": {
			input:  User{RoleHash: strPtr(mustEncryptRole(t, RoleAnalyst))},
			output: RoleAnalyst,
		},
		"viewer": {
			input:  User{RoleHash: strPtr(mustEncryptRole(t, RoleViewer))},
			output: RoleViewer,
		},
		"nil_role_hash": {},
		"unknown_role_hash": {
			input: User{RoleHash: strPtr("unknown")},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.GetRole()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestSetRole(t *testing.T) {
	type input struct {
		role string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  input{role: RoleAdmin},
			output: RoleAdmin,
		},
		"invalid_role": {
			input: input{role: "OWNER"},
			err:   ErrInvalidRole,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			user := User{}

			err := user.SetRole(tc.input.role)

			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				assert.Nil(t, user.RoleHash)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output, user.GetRole())
		})
	}
}

func TestHasRole(t *testing.T) {
	type input struct {
		user User
		role string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  input{user: User{RoleHash: strPtr(mustEncryptRole(t, RoleAdmin))}, role: RoleAdmin},
			output: true,
		},
		"nil_role_hash": {
			input: input{role: RoleAdmin},
		},
		"different_role": {
			input: input{user: User{RoleHash: strPtr(mustEncryptRole(t, RoleAdmin))}, role: RoleViewer},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.user.HasRole(tc.input.role)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestHasAnyRole(t *testing.T) {
	type input struct {
		user  User
		roles []string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  input{user: User{RoleHash: strPtr(mustEncryptRole(t, RoleAnalyst))}, roles: []string{RoleAdmin, RoleAnalyst}},
			output: true,
		},
		"nil_role_hash": {
			input: input{roles: []string{RoleAdmin}},
		},
		"empty_roles": {
			input: input{user: User{RoleHash: strPtr(mustEncryptRole(t, RoleAnalyst))}},
		},
		"no_match": {
			input: input{user: User{RoleHash: strPtr(mustEncryptRole(t, RoleAnalyst))}, roles: []string{RoleAdmin, RoleViewer}},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.user.HasAnyRole(tc.input.roles...)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestScopeIsAdmin(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Scope
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  Scope{Role: RoleAdmin},
			output: true,
		},
		"different_role": {
			input: Scope{Role: RoleViewer},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.IsAdmin()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestScopeIsAnalyst(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Scope
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  Scope{Role: RoleAnalyst},
			output: true,
		},
		"different_role": {
			input: Scope{Role: RoleViewer},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.IsAnalyst()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestScopeIsViewer(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Scope
		mock   mock
		output bool
		err    error
	}{
		"success_true": {
			input:  Scope{Role: RoleViewer},
			output: true,
		},
		"different_role": {
			input: Scope{Role: RoleAdmin},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.IsViewer()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestToScope(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  interface{}
		mock   mock
		output Scope
		err    error
	}{
		"success_from_getters": {
			input: scopeGetter{
				userID:   "user-1",
				username: "user@example.com",
				role:     RoleAdmin,
				jti:      "jti-1",
			},
			output: Scope{
				UserID:   "user-1",
				Username: "user@example.com",
				Role:     RoleAdmin,
				JTI:      "jti-1",
			},
		},
		"success_from_struct": {
			input: struct {
				UserID   string
				Username string
				Role     string
				JTI      string
			}{
				UserID:   "user-2",
				Username: "viewer@example.com",
				Role:     RoleViewer,
				JTI:      "jti-2",
			},
			output: Scope{
				UserID:   "user-2",
				Username: "viewer@example.com",
				Role:     RoleViewer,
				JTI:      "jti-2",
			},
		},
		"unsupported_input": {
			input: "unsupported",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := ToScope(tc.input)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestNewAuditLog(t *testing.T) {
	type input struct {
		userID string
		action string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output AuditLog
		err    error
	}{
		"success": {
			input: input{userID: "user-1", action: ActionLogin},
			output: AuditLog{
				UserID: strPtr("user-1"),
				Action: ActionLogin,
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			before := time.Now()
			output := NewAuditLog(tc.input.userID, tc.input.action)

			require.NotNil(t, output)
			assert.Equal(t, tc.output.UserID, output.UserID)
			assert.Equal(t, tc.output.Action, output.Action)
			assert.False(t, output.CreatedAt.Before(before))
			assert.Equal(t, output.CreatedAt.AddDate(0, 0, 90), output.ExpiresAt)
			assert.NoError(t, tc.err)
		})
	}
}

func TestWithResource(t *testing.T) {
	type input struct {
		resourceType string
		resourceID   string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output AuditLog
		err    error
	}{
		"success": {
			input: input{resourceType: "project", resourceID: "project-1"},
			output: AuditLog{
				ResourceType: strPtr("project"),
				ResourceID:   strPtr("project-1"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			auditLog := &AuditLog{}

			output := auditLog.WithResource(tc.input.resourceType, tc.input.resourceID)

			assert.Same(t, auditLog, output)
			assert.Equal(t, tc.output.ResourceType, output.ResourceType)
			assert.Equal(t, tc.output.ResourceID, output.ResourceID)
			assert.NoError(t, tc.err)
		})
	}
}

func TestWithMetadata(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  map[string]interface{}
		mock   mock
		output AuditLog
		err    error
	}{
		"success": {
			input: map[string]interface{}{"provider": "google"},
			output: AuditLog{
				Metadata: map[string]interface{}{"provider": "google"},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			auditLog := &AuditLog{}

			output := auditLog.WithMetadata(tc.input)

			assert.Same(t, auditLog, output)
			assert.Equal(t, tc.output.Metadata, output.Metadata)
			assert.NoError(t, tc.err)
		})
	}
}

func TestWithIPAddress(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output AuditLog
		err    error
	}{
		"success": {
			input: "127.0.0.1",
			output: AuditLog{
				IPAddress: strPtr("127.0.0.1"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			auditLog := &AuditLog{}

			output := auditLog.WithIPAddress(tc.input)

			assert.Same(t, auditLog, output)
			assert.Equal(t, tc.output.IPAddress, output.IPAddress)
			assert.NoError(t, tc.err)
		})
	}
}

func TestWithUserAgent(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output AuditLog
		err    error
	}{
		"success": {
			input: "Mozilla/5.0",
			output: AuditLog{
				UserAgent: strPtr("Mozilla/5.0"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			auditLog := &AuditLog{}

			output := auditLog.WithUserAgent(tc.input)

			assert.Same(t, auditLog, output)
			assert.Equal(t, tc.output.UserAgent, output.UserAgent)
			assert.NoError(t, tc.err)
		})
	}
}

func strPtr(value string) *string {
	return &value
}

func mustEncryptRole(t *testing.T, role string) string {
	t.Helper()
	hash, err := EncryptRole(role)
	require.NoError(t, err)
	return hash
}

func assertErrorIs(t *testing.T, actual error, expected error) {
	t.Helper()
	if expected == nil {
		assert.NoError(t, actual)
		return
	}
	assert.True(t, errors.Is(actual, expected))
}
