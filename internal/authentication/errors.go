package authentication

import "errors"

var (
	ErrUsernameExisted       = errors.New("username existed")
	ErrUserNotFound          = errors.New("user not found")
	ErrWrongPassword         = errors.New("wrong password")
	ErrWrongOTP              = errors.New("wrong OTP")
	ErrOTPExpired            = errors.New("OTP expired")
	ErrTooManyAttempts       = errors.New("too many attempts")
	ErrUserNotVerified       = errors.New("user not verified")
	ErrInvalidProvider       = errors.New("invalid provider")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrUserVerified          = errors.New("user verified")
	ErrScopeNotFound         = errors.New("scope not found")
	ErrDomainNotAllowed      = errors.New("domain not allowed")
	ErrAccountBlocked        = errors.New("account blocked")
	ErrConfigurationMissing  = errors.New("configuration missing")
	ErrInvalidRedirectURL    = errors.New("invalid redirect url")
	ErrRedirectURLNotAllowed = errors.New("redirect url not allowed")
	ErrInternalSystem        = errors.New("internal system error")
	ErrUserCreation          = errors.New("failed to create or update user")
)
