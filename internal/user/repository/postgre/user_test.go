package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"identity-srv/internal/model"
	"identity-srv/internal/user/repository"

	"github.com/DATA-DOG/go-sqlmock"
	sharedlog "github.com/smap-hcmut/shared-libs/go/log"
	"github.com/stretchr/testify/require"
)

func initRepo(t *testing.T) (*implRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)

	l := sharedlog.NewZapLogger(sharedlog.ZapConfig{
		Level:        sharedlog.LevelFatal,
		Mode:         sharedlog.ModeProduction,
		Encoding:     sharedlog.EncodingJSON,
		ColorEnabled: false,
	})

	return &implRepository{
			l:     l,
			db:    db,
			clock: time.Now,
		},
		mockDB,
		func() {
			mockDB.ExpectClose()
			require.NoError(t, db.Close())
			require.NoError(t, mockDB.ExpectationsWereMet())
		}
}

func TestUpsert(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	viewerHash, err := model.EncryptRole(model.RoleViewer)
	require.NoError(t, err)

	type mockSelect struct {
		output *model.User
		err    error
	}
	type mockUpdate struct {
		isCalled bool
		err      error
	}
	type mockInsert struct {
		isCalled bool
		err      error
	}
	type mock struct {
		selectUser mockSelect
		updateUser mockUpdate
		insertUser mockInsert
	}

	tcs := map[string]struct {
		input  repository.UpsertOptions
		mock   mock
		output model.User
		err    error
	}{
		"success_update_existing_user": {
			input: repository.UpsertOptions{
				Email:     "user@example.com",
				Name:      "Updated User",
				AvatarURL: "avatar.png",
			},
			mock: mock{
				selectUser: mockSelect{output: &model.User{
					ID:        "user-1",
					Email:     "user@example.com",
					RoleHash:  &viewerHash,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}},
				updateUser: mockUpdate{isCalled: true},
			},
			output: model.User{
				ID:        "user-1",
				Email:     "user@example.com",
				Name:      strPtr("Updated User"),
				AvatarURL: strPtr("avatar.png"),
				RoleHash:  &viewerHash,
				IsActive:  true,
				CreatedAt: now,
			},
		},
		"success_insert_new_user": {
			input: repository.UpsertOptions{
				Email:     "new@example.com",
				Name:      "New User",
				AvatarURL: "avatar.png",
			},
			mock: mock{
				selectUser: mockSelect{err: sql.ErrNoRows},
				insertUser: mockInsert{isCalled: true},
			},
			output: model.User{
				Email:     "new@example.com",
				Name:      strPtr("New User"),
				AvatarURL: strPtr("avatar.png"),
				RoleHash:  &viewerHash,
				IsActive:  true,
			},
		},
		"select_error": {
			input: repository.UpsertOptions{Email: "user@example.com"},
			mock:  mock{selectUser: mockSelect{err: errors.New("select error")}},
			err:   errors.New("select error"),
		},
		"update_error": {
			input: repository.UpsertOptions{
				Email:     "user@example.com",
				Name:      "Updated User",
				AvatarURL: "avatar.png",
			},
			mock: mock{
				selectUser: mockSelect{output: &model.User{
					ID:        "user-1",
					Email:     "user@example.com",
					RoleHash:  &viewerHash,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}},
				updateUser: mockUpdate{isCalled: true, err: errors.New("update error")},
			},
			err: errors.New("update error"),
		},
		"insert_error": {
			input: repository.UpsertOptions{
				Email:     "new@example.com",
				Name:      "New User",
				AvatarURL: "avatar.png",
			},
			mock: mock{
				selectUser: mockSelect{err: sql.ErrNoRows},
				insertUser: mockInsert{isCalled: true, err: errors.New("insert error")},
			},
			err: errors.New("sqlboiler: unable to insert into users"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repoImpl, mockDB, cleanup := initRepo(t)
			defer cleanup()
			expectSelectUserByEmail(mockDB, tc.input.Email, tc.mock.selectUser.output, tc.mock.selectUser.err)
			if tc.mock.updateUser.isCalled {
				expectUpdateUser(mockDB, tc.mock.updateUser.err)
			}
			if tc.mock.insertUser.isCalled {
				expectInsertUser(mockDB, tc.mock.insertUser.err)
			}

			output, err := repoImpl.Upsert(ctx, tc.input)

			if tc.err != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.output.Email, output.Email)
			require.Equal(t, tc.output.Name, output.Name)
			require.Equal(t, tc.output.AvatarURL, output.AvatarURL)
			require.Equal(t, tc.output.RoleHash, output.RoleHash)
			require.Equal(t, tc.output.IsActive, output.IsActive)
		})
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	viewerHash, err := model.EncryptRole(model.RoleViewer)
	require.NoError(t, err)
	adminHash, err := model.EncryptRole(model.RoleAdmin)
	require.NoError(t, err)

	type mockSelect struct {
		output *model.User
		err    error
	}
	type mockUpdate struct {
		isCalled bool
		err      error
	}
	type mock struct {
		selectUser mockSelect
		updateUser mockUpdate
	}

	tcs := map[string]struct {
		input  repository.UpdateOptions
		mock   mock
		output error
		err    error
	}{
		"success": {
			input: repository.UpdateOptions{UserID: "user-1", Role: model.RoleAdmin},
			mock: mock{
				selectUser: mockSelect{output: &model.User{
					ID:        "user-1",
					Email:     "user@example.com",
					RoleHash:  &viewerHash,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}},
				updateUser: mockUpdate{isCalled: true},
			},
		},
		"not_found": {
			input: repository.UpdateOptions{UserID: "user-1", Role: model.RoleAdmin},
			mock:  mock{selectUser: mockSelect{err: sql.ErrNoRows}},
			err:   sql.ErrNoRows,
		},
		"select_error": {
			input: repository.UpdateOptions{UserID: "user-1", Role: model.RoleAdmin},
			mock:  mock{selectUser: mockSelect{err: errors.New("select error")}},
			err:   errors.New("select error"),
		},
		"invalid_role": {
			input: repository.UpdateOptions{UserID: "user-1", Role: "OWNER"},
			mock: mock{selectUser: mockSelect{output: &model.User{
				ID:        "user-1",
				Email:     "user@example.com",
				RoleHash:  &viewerHash,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			}}},
			err: model.ErrInvalidRole,
		},
		"update_error": {
			input: repository.UpdateOptions{UserID: "user-1", Role: model.RoleAdmin},
			mock: mock{
				selectUser: mockSelect{output: &model.User{
					ID:        "user-1",
					Email:     "user@example.com",
					RoleHash:  &viewerHash,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}},
				updateUser: mockUpdate{isCalled: true, err: errors.New("update error")},
			},
			err: errors.New("sqlboiler: unable to update users row"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repoImpl, mockDB, cleanup := initRepo(t)
			defer cleanup()
			expectSelectUserByID(mockDB, tc.input.UserID, tc.mock.selectUser.output, tc.mock.selectUser.err)
			if tc.mock.updateUser.isCalled {
				expectUpdateUserRole(mockDB, adminHash, tc.input.UserID, tc.mock.updateUser.err)
			}

			err := repoImpl.Update(ctx, tc.input)

			if tc.err != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDetail(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	viewerHash, err := model.EncryptRole(model.RoleViewer)
	require.NoError(t, err)

	type mockSelect struct {
		output *model.User
		err    error
	}
	type mock struct {
		selectUser mockSelect
	}

	tcs := map[string]struct {
		input  repository.DetailOptions
		mock   mock
		output model.User
		err    error
	}{
		"success": {
			input: repository.DetailOptions{UserID: "user-1"},
			mock: mock{selectUser: mockSelect{output: &model.User{
				ID:        "user-1",
				Email:     "user@example.com",
				RoleHash:  &viewerHash,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			}}},
			output: model.User{
				ID:        "user-1",
				Email:     "user@example.com",
				RoleHash:  &viewerHash,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		"not_found": {
			input: repository.DetailOptions{UserID: "user-1"},
			mock:  mock{selectUser: mockSelect{err: sql.ErrNoRows}},
			err:   sql.ErrNoRows,
		},
		"select_error": {
			input: repository.DetailOptions{UserID: "user-1"},
			mock:  mock{selectUser: mockSelect{err: errors.New("select error")}},
			err:   errors.New("select error"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repoImpl, mockDB, cleanup := initRepo(t)
			defer cleanup()
			expectSelectUserByID(mockDB, tc.input.UserID, tc.mock.selectUser.output, tc.mock.selectUser.err)

			output, err := repoImpl.Detail(ctx, tc.input)

			if tc.err != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.output, output)
		})
	}
}

func expectSelectUserByEmail(mockDB sqlmock.Sqlmock, email string, output *model.User, err error) {
	expectSelectUser(mockDB, `(?s)SELECT .*FROM "identity"\."users".*"email".*LIMIT 1`, email, output, err)
}

func expectSelectUserByID(mockDB sqlmock.Sqlmock, userID string, output *model.User, err error) {
	expectSelectUser(mockDB, `(?s)SELECT .*FROM "identity"\."users".*"id".*LIMIT 1`, userID, output, err)
}

func expectSelectUser(mockDB sqlmock.Sqlmock, query string, arg string, output *model.User, err error) {
	expect := mockDB.ExpectQuery(query).WithArgs(arg)
	if err != nil {
		expect.WillReturnError(err)
		return
	}
	expect.WillReturnRows(userRows(output))
}

func expectUpdateUser(mockDB sqlmock.Sqlmock, err error) {
	expect := mockDB.ExpectExec(`UPDATE "identity"\."users" SET`)
	if err != nil {
		expect.WillReturnError(err)
		return
	}
	expect.WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectUpdateUserRole(mockDB sqlmock.Sqlmock, roleHash string, userID string, err error) {
	expect := mockDB.ExpectExec(`UPDATE "identity"\."users" SET`).
		WithArgs(roleHash, sqlmock.AnyArg(), userID)
	if err != nil {
		expect.WillReturnError(err)
		return
	}
	expect.WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectInsertUser(mockDB sqlmock.Sqlmock, err error) {
	expect := mockDB.ExpectExec(`INSERT INTO "identity"\."users"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		)
	if err != nil {
		expect.WillReturnError(err)
		return
	}
	expect.WillReturnResult(sqlmock.NewResult(0, 1))
}

func userRows(user *model.User) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"id",
		"email",
		"name",
		"avatar_url",
		"role_hash",
		"is_active",
		"last_login_at",
		"created_at",
		"updated_at",
	})
	rows.AddRow(
		user.ID,
		user.Email,
		nilString(user.Name),
		nilString(user.AvatarURL),
		nilString(user.RoleHash),
		user.IsActive,
		nil,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return rows
}

func nilString(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func strPtr(value string) *string {
	return &value
}
