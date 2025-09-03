package usecase

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"

	"github.com/nguyentantai21042004/smap-api/internal/auth"
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/session"
	"github.com/nguyentantai21042004/smap-api/internal/user"
	"github.com/nguyentantai21042004/smap-api/pkg/email"
	optUC "github.com/nguyentantai21042004/smap-api/pkg/otp"
	"github.com/nguyentantai21042004/smap-api/pkg/scope"
)

// ErrEmailExisted
func (uc implUsecase) Register(ctx context.Context, sc models.Scope, ip auth.RegisterInput) (auth.RegisterOutput, error) {
	// Check if email already exists
	_, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{Email: ip.Email})
	switch {
	case err == nil:
		uc.l.Warnf(ctx, "auth.usecase.Register.userUC.GetOne: %v", err)
		return auth.RegisterOutput{}, auth.ErrEmailExisted
	case err != user.ErrUserNotFound:
		uc.l.Errorf(ctx, "auth.usecase.Register.userUC.GetOne: %v", err)
		return auth.RegisterOutput{}, err
	}

	// Encrypt password
	enPss, err := uc.encrypt.Encrypt(ip.Password)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Register.encrypt.Encrypt: %v", err)
		return auth.RegisterOutput{}, err
	}

	// Create new user
	uco, err := uc.userUC.Create(ctx, sc, user.CreateInput{
		Provider:   auth.Web,
		Email:      ip.Email,
		Password:   enPss,
		IsVerified: false,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Register.userUC.Create: %v", err)
		return auth.RegisterOutput{}, err
	}

	return auth.RegisterOutput{
		User: uco.User,
	}, nil
}

// ErrWrongPassword, ErrUserNotFound
func (uc implUsecase) SendOTP(ctx context.Context, sc models.Scope, ip auth.SendOTPInput) error {
	// Get user by email
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Email: ip.Email,
	})
	if err != nil {
		if err == user.ErrUserNotFound {
			uc.l.Warnf(ctx, "auth.usecase.SendOTP.userUC.GetOne: %v", err)
			return auth.ErrUserNotFound
		}
		uc.l.Errorf(ctx, "auth.usecase.SendOTP.userUC.GetOne: %v", err)
		return err
	}

	// Verify password
	pass, err := uc.encrypt.Decrypt(u.PasswordHash)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SendOTP.encrypt.Decrypt: %v", err)
		return err
	}
	if pass != ip.Password {
		uc.l.Warnf(ctx, "auth.usecase.SendOTP: %v", auth.ErrWrongPassword)
		return auth.ErrWrongPassword
	}

	if u.IsVerified {
		uc.l.Warnf(ctx, "auth.usecase.SendOTP: %v", auth.ErrUserVerified)
		return auth.ErrUserVerified
	}

	// Generate new OTP if needed
	now := uc.clock()
	if u.OTP == "" || u.OTPExpiredAt.Before(now) {
		otp, otpExpiredAt := optUC.GenerateOTP(now)
		_, err = uc.userUC.UpdateVerified(ctx, sc, user.UpdateVerifiedInput{
			UserID:       u.ID.Hex(),
			OTP:          otp,
			OTPExpiredAt: otpExpiredAt,
			IsVerified:   false,
		})
		if err != nil {
			uc.l.Errorf(ctx, "auth.usecase.SendOTP.userUC.UpdateVerified: %v", err)
			return err
		}
		u.OTP = otp
		u.OTPExpiredAt = otpExpiredAt
	}

	// Prepare email
	name := u.Email
	if u.FullName != "" {
		name = u.FullName
	}

	expireMin := int(u.OTPExpiredAt.Sub(now).Minutes())
	email, err := email.NewEmail(ctx, email.EmailMeta{
		Recipient:    u.Email,
		TemplateType: email.EmailVerificationTemplate,
	}, email.EmailVerification{
		Name:         name,
		Email:        u.Email,
		OTP:          u.OTP,
		OTPExpireMin: strconv.Itoa(expireMin),
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SendOTP.email.NewEmail: %v", err)
		return err
	}

	// Send email
	if err = uc.PubSendEmailMsg(ctx, sc, auth.PubSendEmailMsgInput{
		Recipient: email.Recipient,
		Subject:   email.Subject,
		Body:      email.Body,
	}); err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SendOTP.PubSendEmailMsg: %v", err)
		return err
	}

	return nil
}

// ErrWrongOTP, ErrUserNotFound
func (uc implUsecase) VerifyOTP(ctx context.Context, sc models.Scope, ip auth.VerifyOTPInput) error {
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Email: ip.Email,
	})
	if err != nil {
		if err == user.ErrUserNotFound {
			uc.l.Warnf(ctx, "auth.usecase.VerifyOTP.userUC.GetOne: %v", err)
			return auth.ErrUserNotFound
		}
		uc.l.Errorf(ctx, "auth.usecase.VerifyOTP.userUC.GetOne: %v", err)
		return err
	}

	now := uc.clock()
	if u.OTP != ip.OTP {
		uc.l.Warnf(ctx, "auth.usecase.VerifyOTP: %v", auth.ErrWrongOTP)
		return auth.ErrWrongOTP
	}

	if u.OTPExpiredAt.Before(now) {
		uc.l.Warnf(ctx, "auth.usecase.VerifyOTP: %v", auth.ErrOTPExpired)
		return auth.ErrOTPExpired
	}

	if _, err = uc.userUC.UpdateVerified(ctx, sc, user.UpdateVerifiedInput{
		UserID:     u.ID.Hex(),
		IsVerified: true,
	}); err != nil {
		uc.l.Errorf(ctx, "auth.usecase.VerifyOTP.userUC.UpdateVerified: %v", err)
		return err
	}

	return nil
}

// ErrWrongPassword, ErrUserNotFound
func (uc implUsecase) Login(ctx context.Context, sc models.Scope, ip auth.LoginInput) (auth.LoginOutput, error) {
	// Get user by email
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Email: ip.Email,
	})
	if err != nil {
		if err == user.ErrUserNotFound {
			uc.l.Warnf(ctx, "auth.usecase.Login.userUC.GetOne: %v", err)
			return auth.LoginOutput{}, auth.ErrUserNotFound
		}
		uc.l.Errorf(ctx, "auth.usecase.Login.userUC.GetOne: %v", err)
		return auth.LoginOutput{}, err
	}

	// Validate user verification status
	if !u.IsVerified {
		uc.l.Warnf(ctx, "auth.usecase.Login: %v", auth.ErrUserNotVerified)
		return auth.LoginOutput{}, auth.ErrUserNotVerified
	}

	// Verify password
	pass, err := uc.encrypt.Decrypt(u.PasswordHash)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Login.encrypt.Decrypt: %v", err)
		return auth.LoginOutput{}, err
	}
	if pass != ip.Password {
		uc.l.Warnf(ctx, "auth.usecase.Login: %v", auth.ErrWrongPassword)
		return auth.LoginOutput{}, auth.ErrWrongPassword
	}

	// Calculate token expiry times
	accessExpiry := 1 * time.Hour
	refreshExpiry := 7 * 24 * time.Hour
	if ip.Remember {
		refreshExpiry = 30 * 24 * time.Hour
	}

	// Create access token
	now := uc.clock()
	accessToken, err := uc.scope.CreateToken(scope.Payload{
		StandardClaims: jwt.StandardClaims{
			Audience:  "authenticate-api",
			ExpiresAt: now.Add(accessExpiry).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "authenticate-api",
			NotBefore: now.Unix(),
			Subject:   u.ID.Hex(),
		},
		UserID:   u.ID.Hex(),
		Username: u.FullName,
		Type:     "access",
		Refresh:  false,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Login.scope.CreateToken: %v", err)
		return auth.LoginOutput{}, err
	}

	// Create session
	refreshToken := primitive.NewObjectID().Hex()
	so, err := uc.sessionUC.Create(ctx, sc, session.CreateSessionInput{
		UserID:       u.ID.Hex(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(refreshExpiry),
		UserAgent:    ip.UserAgent,
		IPAddress:    ip.IPAddress,
		DeviceName:   ip.DeviceName,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Login.sessionUC.Create: %v", err)
		return auth.LoginOutput{}, err
	}

	// Get user role
	r, err := uc.roleUC.Detail(ctx, sc, u.RoleID.Hex())
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.Login.roleUC.GetOne: %v", err)
		return auth.LoginOutput{}, err
	}

	return auth.LoginOutput{
		User: u,
		Role: r.Role,
		Token: auth.TokenOutput{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    so.Session.ExpiresAt,
			SessionID:    so.Session.ID.Hex(),
			TokenType:    "Bearer",
		},
	}, nil
}

func (uc implUsecase) SocialLogin(ctx context.Context, sc models.Scope, ip auth.SocialLoginInput) (auth.SocialLoginOutput, error) {
	if !slices.Contains(auth.SocialProviders, ip.Provider) {
		uc.l.Warnf(ctx, "auth.usecase.SocialLogin: %v", auth.ErrInvalidProvider)
		return auth.SocialLoginOutput{}, auth.ErrInvalidProvider
	}

	oauthConfig, err := uc.getOAuthConfig(ctx, ip.Provider)
	if err != nil {
		uc.l.Warnf(ctx, "auth.usecase.SocialLogin.getOAuthConfig: %v", err)
		return auth.SocialLoginOutput{}, err
	}

	// Generate random state
	state := primitive.NewObjectID().Hex()
	url := oauthConfig.Config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)

	return auth.SocialLoginOutput{
		URL: url,
	}, nil
}

func (uc implUsecase) SocialCallback(ctx context.Context, sc models.Scope, ip auth.SocialCallbackInput) (auth.SocialCallbackOutput, error) {
	if !slices.Contains(auth.SocialProviders, ip.Provider) {
		uc.l.Warnf(ctx, "auth.usecase.SocialCallback: %v", auth.ErrInvalidProvider)
		return auth.SocialCallbackOutput{}, auth.ErrInvalidProvider
	}

	su, err := uc.getUserInfo(ctx, auth.GetUserInfoInput{
		Provider: ip.Provider,
		Code:     ip.Code,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SocialCallback.getUserInfo: %v", err)
		return auth.SocialCallbackOutput{}, err
	}

	if su.Email == "" {
		uc.l.Warnf(ctx, "auth.usecase.SocialCallback: %v", auth.ErrInvalidEmail)
		return auth.SocialCallbackOutput{}, auth.ErrInvalidEmail
	}

	// Check if user already exists
	u, err := uc.userUC.GetOne(ctx, sc, user.GetOneInput{
		Email: su.Email,
	})
	if err != nil {
		if err == user.ErrUserNotFound {
			pss, err := uc.encrypt.Encrypt(ip.Provider)
			if err != nil {
				uc.l.Errorf(ctx, "auth.usecase.SocialCallback.encrypt.Encrypt: %v", err)
				return auth.SocialCallbackOutput{}, err
			}

			uco, err := uc.userUC.Create(ctx, sc, user.CreateInput{
				Provider:   ip.Provider,
				ProviderID: su.ID,
				Email:      su.Email,
				Password:   pss,
				FullName:   su.Name,
				AvatarURL:  su.AvatarURL,
				IsVerified: true,
			})
			if err != nil {
				uc.l.Errorf(ctx, "auth.usecase.SocialCallback.userUC.Create: %v", err)
				return auth.SocialCallbackOutput{}, err
			}
			u = uco.User
		} else {
			uc.l.Errorf(ctx, "auth.usecase.SocialCallback.userUC.GetOne: %v", err)
			return auth.SocialCallbackOutput{}, err
		}
	}

	// Check provider and providerID
	if u.Provider != ip.Provider || u.ProviderID != su.ID {
		uc.l.Warnf(ctx, "auth.usecase.SocialCallback: %v", auth.ErrInvalidProvider)
		return auth.SocialCallbackOutput{}, auth.ErrInvalidProvider
	}

	pss, err := uc.encrypt.Decrypt(u.PasswordHash)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SocialCallback.encrypt.Decrypt: %v", err)
		return auth.SocialCallbackOutput{}, err
	}

	if pss != ip.Provider {
		uc.l.Warnf(ctx, "auth.usecase.SocialCallback: %v", auth.ErrInvalidProvider)
		return auth.SocialCallbackOutput{}, auth.ErrInvalidProvider
	}

	to, err := uc.generateTokenAndSession(ctx, auth.GenerateTokenAndSessionInput{
		UserID:   u.ID.Hex(),
		Username: u.FullName,
		Remember: false,
		Scope:    sc,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SocialCallback.generateTokenAndSession: %v", err)
		return auth.SocialCallbackOutput{}, err
	}

	r, err := uc.roleUC.Detail(ctx, sc, u.RoleID.Hex())
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.SocialCallback.roleUC.Detail: %v", err)
		return auth.SocialCallbackOutput{}, err
	}

	return auth.SocialCallbackOutput{
		User:  u,
		Role:  r.Role,
		Token: to,
	}, nil
}
