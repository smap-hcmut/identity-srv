package usecase

import (
	"context"
	"smap-api/internal/authentication"
	"smap-api/internal/model"
	"smap-api/internal/user"
	"smap-api/pkg/email"
	"smap-api/pkg/encrypter"
	"smap-api/pkg/scope"
	"smap-api/pkg/util"
	"strconv"

	"github.com/golang-jwt/jwt"
)

// Login implements authentication.UseCase.
func (uc *implUsecase) Login(ctx context.Context, sc model.Scope, ip authentication.LoginInput) (authentication.LoginOutput, error) {
	// Retrieve user by email.
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Username: ip.Email,
	})
	if err != nil {
		if err == user.ErrUserNotFound {
			uc.l.Warnf(ctx, "authentication.usecase.Login.userUC.GetOne: %v", err)
			return authentication.LoginOutput{}, authentication.ErrUserNotFound
		}
		uc.l.Errorf(ctx, "authentication.usecase.Login.userUC.GetOne: %v", err)
		return authentication.LoginOutput{}, err
	}

	// Check if the user account is active.
	if !*u.IsActive {
		uc.l.Warnf(ctx, "authentication.usecase.Login: %v", authentication.ErrUserNotVerified)
		return authentication.LoginOutput{}, authentication.ErrUserNotVerified
	}

	// Verify the provided password matches the stored password.
	pass, err := uc.encrypt.Decrypt(*u.PasswordHash)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Login.encrypt.Decrypt: %v", err)
		return authentication.LoginOutput{}, err
	}
	if pass != ip.Password {
		uc.l.Warnf(ctx, "authentication.usecase.Login: %v", authentication.ErrWrongPassword)
		return authentication.LoginOutput{}, authentication.ErrWrongPassword
	}

	// Generate a new access token for the user session.
	now := uc.clock()
	accessToken, err := uc.scope.CreateToken(scope.Payload{
		StandardClaims: jwt.StandardClaims{
			Audience:  model.SMAPAPI,
			ExpiresAt: 0,
			IssuedAt:  now.Unix(),
			Issuer:    model.SMAPAPI,
			NotBefore: now.Unix(),
			Subject:   u.ID,
		},
		UserID:   u.ID,
		Username: u.Username,
		Type:     model.ScopeTypeAccess,
		Refresh:  false,
	})
	if err != nil {
		uc.l.Errorf(ctx, "authentication.usecase.Login.scope.CreateToken: %v", err)
		return authentication.LoginOutput{}, err
	}

	return authentication.LoginOutput{
		User: u,
		Token: authentication.TokenOutput{
			AccessToken: accessToken,
			TokenType:   "Bearer",
		},
	}, nil
}

// Register implements authentication.UseCase.
func (uc *implUsecase) Register(ctx context.Context, sc model.Scope, ip authentication.RegisterInput) (authentication.RegisterOutput, error) {
	// Step 1: Check if a user with the specified username (email) already exists.
	_, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{Username: ip.Email})
	switch {
	case err == nil:
		uc.l.Warnf(ctx, "authentication.usecase.Register.userUC.GetOne: %v", err)
		return authentication.RegisterOutput{}, authentication.ErrUsernameExisted
	case err != user.ErrUserNotFound:
		uc.l.Errorf(ctx, "authentication.usecase.Register.userUC.GetOne: %v", err)
		return authentication.RegisterOutput{}, err
	}

	// Step 2: Securely hash the provided password before storing.
	hash, err := encrypter.HashPassword(ip.Password)
	if err != nil {
		uc.l.Errorf(ctx, "authentication.usecase.Register.encrypter.HashPassword: %v", err)
		return authentication.RegisterOutput{}, err
	}

	// Step 3: Create a new user with the checked username and hashed password.
	uco, err := uc.userUC.Create(ctx, sc, user.CreateInput{
		Username: ip.Email,
		Password: hash,
	})
	if err != nil {
		uc.l.Errorf(ctx, "authentication.usecase.Register.userUC.Create: %v", err)
		return authentication.RegisterOutput{}, err
	}

	// Return registration result containing the new user's information.
	return authentication.RegisterOutput{
		User: uco.User,
	}, nil
}

// SendOTP implements authentication.UseCase.
func (uc *implUsecase) SendOTP(ctx context.Context, sc model.Scope, ip authentication.SendOTPInput) error {
	// Step 1: Retrieve the user by the provided username (email).
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Username: ip.Email,
	})
	if err != nil {
		// If the user does not exist, return the ErrUserNotFound error.
		if err == user.ErrUserNotFound {
			uc.l.Warnf(ctx, "authentication.usecase.SendOTP.userUC.GetOne: %v", err)
			return authentication.ErrUserNotFound
		}
		// For all other errors, log and return the error.
		uc.l.Errorf(ctx, "authentication.usecase.SendOTP.userUC.GetOne: %v", err)
		return err
	}

	// Step 2: Verify the provided password matches the user's stored password.
	pass, err := uc.encrypt.Decrypt(*u.PasswordHash)
	if err != nil {
		// Log and return if password decryption fails.
		uc.l.Errorf(ctx, "authentication.usecase.SendOTP.encrypt.Decrypt: %v", err)
		return err
	}
	if pass != ip.Password {
		// If the password does not match, log and return wrong password error.
		uc.l.Warnf(ctx, "authentication.usecase.SendOTP: %v", authentication.ErrWrongPassword)
		return authentication.ErrWrongPassword
	}

	// Step 3: Check if the user account is already verified (active).
	if *u.IsActive {
		uc.l.Warnf(ctx, "authentication.usecase.SendOTP: %v", authentication.ErrUserVerified)
		return authentication.ErrUserVerified
	}

	// Step 4: If the OTP does not exist or has expired, generate a new OTP and update the user.
	now := uc.clock()
	if u.OTP == nil || u.OTPExpiredAt.Before(now) {
		otp, otpExpiredAt := util.GenerateOTP()
		_, err = uc.userUC.Update(ctx, sc, user.UpdateInput{
			ID:           u.ID,
			OTP:          &otp,
			OTPExpiredAt: &otpExpiredAt,
		})
		if err != nil {
			uc.l.Errorf(ctx, "authentication.usecase.SendOTP.userUC.Update: %v", err)
			return err
		}
		u.OTP = &otp
		u.OTPExpiredAt = &otpExpiredAt
	}

	// Step 5: Prepare the recipient name for the verification email.
	name := u.Username
	if u.FullName != nil {
		name = *u.FullName
	}

	// Step 6: Calculate time in minutes until OTP expires, then prepare the email content and metadata.
	expireMin := int(u.OTPExpiredAt.Sub(now).Minutes())
	email, err := email.NewEmail(ctx, email.EmailMeta{
		Recipient:    u.Username,
		TemplateType: email.EmailVerificationTemplate,
	}, email.EmailVerification{
		Name:         name,
		Email:        u.Username,
		OTP:          *u.OTP,
		OTPExpireMin: strconv.Itoa(expireMin),
	})
	if err != nil {
		// Log and return if email generation fails.
		uc.l.Errorf(ctx, "authentication.usecase.SendOTP.email.NewEmail: %v", err)
		return err
	}

	// Step 7: Publish the email message to the appropriate channel for sending.
	if err = uc.PublishSendEmail(ctx, sc, authentication.PublishSendEmailMsgInput{
		Recipient: email.Recipient,
		Subject:   email.Subject,
		Body:      email.Body,
	}); err != nil {
		// Log and return any error occurred during publishing the email message.
		uc.l.Errorf(ctx, "authentication.usecase.SendOTP.PublishSendEmail: %v", err)
		return err
	}

	// If all steps succeed, return nil to indicate success.
	return nil
}

// VerifyOTP implements authentication.UseCase.
// VerifyOTP checks the OTP provided by the user and activates the account upon successful verification.
func (uc *implUsecase) VerifyOTP(ctx context.Context, sc model.Scope, ip authentication.VerifyOTPInput) error {
	// Step 1: Retrieve user information based on the provided email.
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Username: ip.Email,
	})
	if err != nil {
		// If the user is not found, log and return a specific error.
		if err == user.ErrUserNotFound {
			uc.l.Warnf(ctx, "authentication.usecase.VerifyOTP.userUC.GetOne: %v", err)
			return authentication.ErrUserNotFound
		}
		// Log and return any other errors during user retrieval.
		uc.l.Errorf(ctx, "authentication.usecase.VerifyOTP.userUC.GetOne: %v", err)
		return err
	}

	// Get the current time to check OTP expiration.
	now := uc.clock()

	// Step 2: Compare the provided OTP with the one stored in the system.
	if u.OTP == nil || *u.OTP != ip.OTP {
		// If the OTP is incorrect, log and return an error.
		uc.l.Warnf(ctx, "authentication.usecase.VerifyOTP: %v", authentication.ErrWrongOTP)
		return authentication.ErrWrongOTP
	}

	// Step 3: Check if the OTP has expired.
	if u.OTPExpiredAt == nil || u.OTPExpiredAt.Before(now) {
		// If the OTP is expired, log and return an error.
		uc.l.Warnf(ctx, "authentication.usecase.VerifyOTP: %v", authentication.ErrOTPExpired)
		return authentication.ErrOTPExpired
	}

	// Step 4: Activate the user account if the OTP is valid.
	if _, err = uc.userUC.Update(ctx, sc, user.UpdateInput{
		ID:       u.ID,
		IsActive: &[]bool{true}[0],
	}); err != nil {
		// Log and return any error occurred during user activation.
		uc.l.Errorf(ctx, "authentication.usecase.VerifyOTP.userUC.Update: %v", err)
		return err
	}

	// If everything is successful, return nil.
	return nil
}
