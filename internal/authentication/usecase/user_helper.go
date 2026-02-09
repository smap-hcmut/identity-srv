package usecase

import (
	"context"
	"database/sql"
	"smap-api/internal/model"
	"smap-api/internal/sqlboiler"
	postgresPkg "smap-api/pkg/postgre"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// CreateOrUpdateUserDirect creates or updates a user directly using SQLBoiler
// This bypasses the user usecase/repository which has legacy code issues
func (u *implUsecase) CreateOrUpdateUserDirect(ctx context.Context, db *sql.DB, email, name, avatarURL string) (*model.User, error) {
	// Try to find existing user by email
	existingUser, err := sqlboiler.Users(
		sqlboiler.UserWhere.Email.EQ(email),
	).One(ctx, db)

	if err == nil {
		// User exists - update
		existingUser.Name = null.StringFrom(name)
		existingUser.AvatarURL = null.StringFrom(avatarURL)
		existingUser.LastLoginAt = null.TimeFrom(time.Now())
		existingUser.UpdatedAt = time.Now()

		_, updateErr := existingUser.Update(ctx, db, boil.Infer())
		if updateErr != nil {
			u.l.Errorf(ctx, "Failed to update user: %v", updateErr)
			return nil, updateErr
		}

		return model.NewUserFromDB(existingUser), nil
	}

	if err != sql.ErrNoRows {
		u.l.Errorf(ctx, "Failed to query user: %v", err)
		return nil, err
	}

	// User doesn't exist - create new
	newUser := &sqlboiler.User{
		ID:          postgresPkg.NewUUID(),
		Email:       email,
		Name:        null.StringFrom(name),
		AvatarURL:   null.StringFrom(avatarURL),
		IsActive:    null.BoolFrom(true),
		LastLoginAt: null.TimeFrom(time.Now()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set default role
	roleHash, err := model.EncryptRole(model.RoleViewer)
	if err != nil {
		u.l.Errorf(ctx, "Failed to encrypt role: %v", err)
		return nil, err
	}
	newUser.RoleHash = roleHash

	if err := newUser.Insert(ctx, db, boil.Infer()); err != nil {
		u.l.Errorf(ctx, "Failed to insert user: %v", err)
		return nil, err
	}

	return model.NewUserFromDB(newUser), nil
}

// UpdateUserRoleDirect updates user role directly using SQLBoiler
func (u *implUsecase) UpdateUserRoleDirect(ctx context.Context, db *sql.DB, userID, role string) error {
	user, err := sqlboiler.Users(
		sqlboiler.UserWhere.ID.EQ(userID),
	).One(ctx, db)

	if err != nil {
		if err == sql.ErrNoRows {
			u.l.Errorf(ctx, "User not found: %s", userID)
			return err
		}
		u.l.Errorf(ctx, "Failed to query user: %v", err)
		return err
	}

	// Encrypt and set role
	roleHash, err := model.EncryptRole(role)
	if err != nil {
		u.l.Errorf(ctx, "Failed to encrypt role: %v", err)
		return err
	}

	user.RoleHash = roleHash
	user.UpdatedAt = time.Now()

	_, updateErr := user.Update(ctx, db, boil.Whitelist(
		sqlboiler.UserColumns.RoleHash,
		sqlboiler.UserColumns.UpdatedAt,
	))
	if updateErr != nil {
		u.l.Errorf(ctx, "Failed to update user role: %v", updateErr)
		return updateErr
	}

	u.l.Infof(ctx, "Updated user %s role to %s", userID, role)
	return nil
}

// GetUserByIDDirect gets user by ID directly using SQLBoiler
func (u *implUsecase) GetUserByIDDirect(ctx context.Context, db *sql.DB, userID string) (*model.User, error) {
	user, err := sqlboiler.Users(
		sqlboiler.UserWhere.ID.EQ(userID),
	).One(ctx, db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		u.l.Errorf(ctx, "Failed to query user: %v", err)
		return nil, err
	}

	return model.NewUserFromDB(user), nil
}
