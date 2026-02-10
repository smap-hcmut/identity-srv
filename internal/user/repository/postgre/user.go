package postgres

import (
	"context"
	"database/sql"
	"time"

	"smap-api/internal/model"
	"smap-api/internal/sqlboiler"
	"smap-api/internal/user/repository"
	postgresPkg "smap-api/pkg/postgre"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// Upsert creates or updates a user by email (for OAuth)
func (r *implRepository) Upsert(ctx context.Context, opts repository.UpsertOptions) (model.User, error) {
	// Try to find existing user by email
	existingUser, err := sqlboiler.Users(
		sqlboiler.UserWhere.Email.EQ(opts.Email),
	).One(ctx, r.db)

	if err == nil {
		// User exists - update
		existingUser.Name = null.StringFrom(opts.Name)
		existingUser.AvatarURL = null.StringFrom(opts.AvatarURL)
		existingUser.LastLoginAt = null.TimeFrom(time.Now())
		existingUser.UpdatedAt = time.Now()

		_, updateErr := existingUser.Update(ctx, r.db, boil.Infer())
		if updateErr != nil {
			r.l.Errorf(ctx, "Failed to update user: %v", updateErr)
			return model.User{}, updateErr
		}

		return *model.NewUserFromDB(existingUser), nil
	}

	if err != sql.ErrNoRows {
		r.l.Errorf(ctx, "Failed to query user: %v", err)
		return model.User{}, err
	}

	// User doesn't exist - create new
	newUser := &sqlboiler.User{
		ID:          postgresPkg.NewUUID(),
		Email:       opts.Email,
		Name:        null.StringFrom(opts.Name),
		AvatarURL:   null.StringFrom(opts.AvatarURL),
		IsActive:    null.BoolFrom(true),
		LastLoginAt: null.TimeFrom(time.Now()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set default role
	roleHash, err := model.EncryptRole(model.RoleViewer)
	if err != nil {
		r.l.Errorf(ctx, "Failed to encrypt role: %v", err)
		return model.User{}, err
	}
	newUser.RoleHash = roleHash

	if err := newUser.Insert(ctx, r.db, boil.Infer()); err != nil {
		r.l.Errorf(ctx, "Failed to insert user: %v", err)
		return model.User{}, err
	}

	return *model.NewUserFromDB(newUser), nil
}

// Update updates user (currently only supports role update)
func (r *implRepository) Update(ctx context.Context, opts repository.UpdateOptions) error {
	user, err := sqlboiler.Users(
		sqlboiler.UserWhere.ID.EQ(opts.UserID),
	).One(ctx, r.db)

	if err != nil {
		if err == sql.ErrNoRows {
			r.l.Errorf(ctx, "User not found: %s", opts.UserID)
			return err
		}
		r.l.Errorf(ctx, "Failed to query user: %v", err)
		return err
	}

	// Encrypt and set role
	roleHash, err := model.EncryptRole(opts.Role)
	if err != nil {
		r.l.Errorf(ctx, "Failed to encrypt role: %v", err)
		return err
	}

	user.RoleHash = roleHash
	user.UpdatedAt = time.Now()

	_, updateErr := user.Update(ctx, r.db, boil.Whitelist(
		sqlboiler.UserColumns.RoleHash,
		sqlboiler.UserColumns.UpdatedAt,
	))
	if updateErr != nil {
		r.l.Errorf(ctx, "Failed to update user role: %v", updateErr)
		return updateErr
	}

	r.l.Infof(ctx, "Updated user %s role to %s", opts.UserID, opts.Role)
	return nil
}

// Detail gets user by ID
func (r *implRepository) Detail(ctx context.Context, opts repository.DetailOptions) (model.User, error) {
	user, err := sqlboiler.Users(
		sqlboiler.UserWhere.ID.EQ(opts.UserID),
	).One(ctx, r.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return model.User{}, sql.ErrNoRows
		}
		r.l.Errorf(ctx, "Failed to query user: %v", err)
		return model.User{}, err
	}

	return *model.NewUserFromDB(user), nil
}
