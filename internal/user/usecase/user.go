package usecase

import (
	"context"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/mongo"
	"github.com/nguyentantai21042004/smap-api/pkg/otp"
)

const (
	RoleUserCode = "GUEST"
)

func (uc implUsecase) Get(ctx context.Context, sc models.Scope, ip user.GetInput) (user.GetUserOutput, error) {
	u, p, err := uc.repo.Get(ctx, sc, user.GetOptions{
		Filter:   ip.Filter,
		PagQuery: ip.PagQuery,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.Get.repo.Get: %v", err)
		return user.GetUserOutput{}, err
	}

	rIDs := make([]string, 0, len(u))
	for _, u := range u {
		rIDs = append(rIDs, u.RoleID.Hex())
	}

	r, err := uc.roleUC.List(ctx, sc, role.ListInput{
		Filter: role.Filter{
			IDs: rIDs,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.Get.roleUC.List: %v", err)
		return user.GetUserOutput{}, err
	}

	return user.GetUserOutput{
		Users:      u,
		Roles:      r.Roles,
		Pagination: p,
	}, nil
}

func (uc implUsecase) GetOne(ctx context.Context, sc models.Scope, ip user.GetOneInput) (models.User, error) {
	u, err := uc.repo.GetOne(ctx, sc, user.GetOneOptions(ip))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.l.Warnf(ctx, "internal.user.usecase.GetOne.repo.GetOne: %v", err)
			return models.User{}, user.ErrUserNotFound
		}
		uc.l.Errorf(ctx, "internal.user.usecase.GetOne.repo.GetOne: %v", err)
		return models.User{}, err
	}

	return u, nil
}

func (uc implUsecase) Detail(ctx context.Context, sc models.Scope, ID string) (user.UserOutput, error) {
	u, err := uc.repo.Detail(ctx, sc, ID)
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.Detail.repo.Detail: %v", err)
		return user.UserOutput{}, err
	}

	r, err := uc.roleUC.Detail(ctx, sc, u.RoleID.Hex())
	if err != nil {
		if err == role.ErrRoleNotFound {
			uc.l.Warnf(ctx, "internal.user.usecase.Detail.roleUC.Detail: %v", err)
			return user.UserOutput{}, user.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "internal.user.usecase.Detail.roleUC.Detail: %v", err)
		return user.UserOutput{}, err
	}

	return user.UserOutput{
		User: u,
		Role: r.Role,
	}, nil
}

func (uc implUsecase) Create(ctx context.Context, sc models.Scope, ip user.CreateInput) (user.UserOutput, error) {
	otp, otpExpiredAt := otp.GenerateOTP(uc.clock())

	r, err := uc.roleUC.GetOne(ctx, sc, role.GetOneInput{
		Filter: role.Filter{
			Code: []string{RoleUserCode},
		},
	})
	if err != nil {
		if err == role.ErrRoleNotFound {
			uc.l.Warnf(ctx, "internal.user.usecase.Create.roleUC.GetOne: %v", err)
			return user.UserOutput{}, user.ErrRoleNotFound
		}
		uc.l.Errorf(ctx, "internal.user.usecase.Create.roleUC.GetOne: %v", err)
		return user.UserOutput{}, err
	}

	u, err := uc.repo.Create(ctx, sc, user.CreateOptions{
		Email:        ip.Email,
		Password:     ip.Password,
		FullName:     ip.FullName,
		OTP:          otp,
		OTPExpiredAt: otpExpiredAt,
		IsVerified:   ip.IsVerified,
		RoleID:       r.Role.ID.Hex(),
		Provider:     ip.Provider,
		ProviderID:   ip.ProviderID,
		AvatarURL:    ip.AvatarURL,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.Create.repo.Create: %v", err)
		return user.UserOutput{}, err
	}

	return user.UserOutput{
		User: u,
		Role: r.Role,
	}, nil
}

func (uc implUsecase) UpdateVerified(ctx context.Context, sc models.Scope, ip user.UpdateVerifiedInput) (user.UserOutput, error) {
	u, err := uc.repo.Detail(ctx, sc, ip.UserID)
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.UpdateVerified.repo.Detail: %v", err)
		return user.UserOutput{}, err
	}

	uo, err := uc.repo.UpdateVerified(ctx, sc, user.UpdateVerifiedOptions{
		ID:           u.ID.Hex(),
		OTP:          ip.OTP,
		OTPExpiredAt: ip.OTPExpiredAt,
		IsVerified:   ip.IsVerified,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.UpdateVerified.repo.UpdateVerified: %v", err)
		return user.UserOutput{}, err
	}

	return user.UserOutput{
		User: uo,
		Role: models.Role{},
	}, nil
}

func (uc implUsecase) DetailMe(ctx context.Context, sc models.Scope) (user.UserOutput, error) {
	u, err := uc.repo.Detail(ctx, sc, sc.UserID)
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.DetailMe.repo.Detail: %v", err)
		return user.UserOutput{}, err
	}

	if u.ID.Hex() != sc.UserID {
		uc.l.Warnf(ctx, "internal.user.usecase.DetailMe.repo.Detail: %v", "user id not match")
		return user.UserOutput{}, user.ErrPermissionDenied
	}

	r, err := uc.roleUC.Detail(ctx, sc, u.RoleID.Hex())
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.DetailMe.roleUC.Detail: %v", err)
		return user.UserOutput{}, err
	}

	return user.UserOutput{
		User: u,
		Role: r.Role,
	}, nil
}

func (uc implUsecase) UpdateAvatar(ctx context.Context, sc models.Scope, ip user.UpdateAvatarInput) error {
	_, err := uc.repo.Detail(ctx, sc, ip.UserID)
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.UpdateAvatar.repo.Detail: %v", err)
		return err
	}

	// Update the avatar
	_, err = uc.repo.UpdateAvatar(ctx, sc, user.UpdateAvatarOptions{
		ID:        ip.UserID,
		AvatarURL: ip.AvatarURL,
	})
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.UpdateAvatar.repo.UpdateAvatar: %v", err)
		return err
	}

	return nil
}

func (uc implUsecase) Delete(ctx context.Context, sc models.Scope, ids []string) error {
	err := uc.repo.Delete(ctx, sc, ids)
	if err != nil {
		uc.l.Errorf(ctx, "internal.user.usecase.Delete.repo.Delete: %v", err)
		return err
	}

	return nil
}
