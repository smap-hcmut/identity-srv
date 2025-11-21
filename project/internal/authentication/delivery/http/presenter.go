package http

import (
	"smap-api/pkg/util"

	"smap-api/internal/authentication"
)

// Register API Request
type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (r registerReq) validate() error {
	if err := util.IsEmail(r.Email); err != nil {
		return errWrongBody
	}
	if err := util.IsPassword(r.Password); err != nil {
		return errWrongBody
	}
	return nil
}

func (r registerReq) toInput() authentication.RegisterInput {
	return authentication.RegisterInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

// Send OTP API Request
type sendOTPReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (r sendOTPReq) validate() error {
	if err := util.IsEmail(r.Email); err != nil {
		return errWrongBody
	}
	if err := util.IsPassword(r.Password); err != nil {
		return errWrongBody
	}
	return nil
}

func (r sendOTPReq) toInput() authentication.SendOTPInput {
	return authentication.SendOTPInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

// Verify OTP API Request
type verifyOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,min=6"`
}

func (r verifyOTPReq) validate() error {
	if err := util.IsEmail(r.Email); err != nil {
		return errWrongBody
	}
	if err := util.IsOTP(r.OTP); err != nil {
		return errWrongBody
	}
	return nil
}

func (r verifyOTPReq) toInput() authentication.VerifyOTPInput {
	return authentication.VerifyOTPInput{
		Email: r.Email,
		OTP:   r.OTP,
	}
}

// Login API Request
type loginReq struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	Remember   bool   `json:"remember"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	DeviceName string `json:"device_name"`
}

func (r loginReq) validate() error {
	if err := util.IsEmail(r.Email); err != nil {
		return errWrongBody
	}
	if err := util.IsPassword(r.Password); err != nil {
		return errWrongBody
	}
	return nil
}

func (r loginReq) toInput() authentication.LoginInput {
	return authentication.LoginInput{
		Email:      r.Email,
		Password:   r.Password,
		Remember:   r.Remember,
		UserAgent:  r.UserAgent,
		IPAddress:  r.IPAddress,
		DeviceName: r.DeviceName,
	}
}

type respObj struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type userObj struct {
	ID    string  `json:"id"`
	Email string  `json:"email"`
	Name  string  `json:"name,omitempty"`
	Role  respObj `json:"role"`
}

type loginResp struct {
	User      userObj `json:"user"`
	Token     string  `json:"token"`
	ExpiresAt string  `json:"expires_at"`
	SessionID string  `json:"session_id"`
	TokenType string  `json:"token_type"`
}

func (h handler) newLoginResp(o authentication.LoginOutput) *loginResp {
	return &loginResp{
		User: userObj{
			ID:    o.User.ID,
			Email: o.User.Username,
			Name:  *o.User.FullName,
		},
		Token:     o.Token.AccessToken,
		TokenType: o.Token.TokenType,
	}
}