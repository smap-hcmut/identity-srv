package mongo

import (
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo implRepository) buildUserModel(sc models.Scope, opt user.CreateOptions) (models.User, error) {
	now := repo.clock()
	roleID, err := primitive.ObjectIDFromHex(opt.RoleID)
	if err != nil {
		return models.User{}, err
	}
	
	user := models.User{
		ID:           repo.db.NewObjectID(),
		Email:        opt.Email,
		PasswordHash: opt.Password,
		FullName:     opt.FullName,
		IsVerified:   opt.IsVerified,
		AvatarURL:    opt.AvatarURL,
		Provider:     opt.Provider,
		ProviderID:   opt.ProviderID,
		OTP:          opt.OTP,
		OTPExpiredAt: opt.OTPExpiredAt,
		RoleID:       roleID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return user, nil
}

func (repo implRepository) buildUpdateVerified(opt user.UpdateVerifiedOptions) (models.User, bson.M) {
	update := bson.M{
		"updated_at":     repo.clock(),
		"otp":            opt.OTP,
		"otp_expired_at": opt.OTPExpiredAt,
		"is_verified":    opt.IsVerified,
	}

	// Mock user for response - in real implementation you'd fetch the updated user
	user := models.User{
		OTP:          opt.OTP,
		OTPExpiredAt: opt.OTPExpiredAt,
		IsVerified:   opt.IsVerified,
	}

	return user, update
}

func (repo implRepository) buildUpdateAvatar(opt user.UpdateAvatarOptions) (models.User, bson.M) {
	update := bson.M{
		"updated_at":  repo.clock(),
		"avatar_url": opt.AvatarURL,
	}

	// Mock user for response - in real implementation you'd fetch the updated user
	user := models.User{
		AvatarURL: opt.AvatarURL,
	}

	return user, update
}