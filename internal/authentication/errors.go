package authentication

import "errors"

var (
	ErrUsernameExisted  = errors.New("username existed")
	ErrUserNotFound     = errors.New("user not found")
	ErrWrongPassword    = errors.New("wrong password")
	ErrWrongOTP         = errors.New("wrong OTP")
	ErrOTPExpired       = errors.New("OTP expired")
	ErrTooManyAttempts  = errors.New("too many attempts")
	ErrUserNotVerified  = errors.New("user not verified")
	ErrInvalidProvider  = errors.New("invalid provider")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrUserVerified     = errors.New("user verified")
	ErrScopeNotFound    = errors.New("scope not found")
	ErrDomainNotAllowed = errors.New("domain not allowed")
	ErrAccountBlocked   = errors.New("account blocked")
)
