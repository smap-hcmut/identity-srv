package http

import (
	"slices"

	"github.com/nguyentantai21042004/smap-api/internal/auth"
	"github.com/nguyentantai21042004/smap-api/pkg/util"
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

func (r registerReq) toInput() auth.RegisterInput {
	return auth.RegisterInput{
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

func (r sendOTPReq) toInput() auth.SendOTPInput {
	return auth.SendOTPInput{
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

func (r verifyOTPReq) toInput() auth.VerifyOTPInput {
	return auth.VerifyOTPInput{
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

func (r loginReq) toInput() auth.LoginInput {
	return auth.LoginInput{
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

func (h handler) newLoginResp(o auth.LoginOutput) *loginResp {
	return &loginResp{
		User: userObj{
			ID:    o.User.ID.Hex(),
			Email: o.User.Email,
			Name:  o.User.FullName,
			Role:  respObj{ID: o.Role.ID.Hex(), Name: o.Role.Name},
		},
		Token:     o.Token.AccessToken,
		ExpiresAt: util.FormatTime(o.Token.ExpiresAt, util.DateTimeFormat),
		SessionID: o.Token.SessionID,
		TokenType: o.Token.TokenType,
	}
}

// Social Login API Request
type socialLoginReq struct {
	Provider string `uri:"provider" binding:"required"`
	Redirect bool   `query:"redirect" binding:"omitempty"`
}

func (r socialLoginReq) validate() error {
	if !slices.Contains(auth.SocialProviders, r.Provider) {
		return errWrongBody
	}
	return nil
}

func (r socialLoginReq) toInput() auth.SocialLoginInput {
	return auth.SocialLoginInput{
		Provider: r.Provider,
		Redirect: r.Redirect,
	}
}

type socialLoginResp struct {
	URL string `json:"url"`
}

func (h handler) newSocialLoginResp(o auth.SocialLoginOutput) *socialLoginResp {
	return &socialLoginResp{
		URL: o.URL,
	}
}

// Social Callback API Request
type socialCallbackReq struct {
	Provider string `uri:"provider"`
	Code     string `form:"code" binding:"required"`
}

func (r socialCallbackReq) validate() error {
	if !slices.Contains(auth.SocialProviders, r.Provider) {
		return errWrongBody
	}
	return nil
}

func (r socialCallbackReq) toInput() auth.SocialCallbackInput {
	return auth.SocialCallbackInput{
		Provider: r.Provider,
		Code:     r.Code,
	}
}

type socialCallbackResp struct {
	User      userObj `json:"user"`
	Token     string  `json:"token"`
	ExpiresAt string  `json:"expires_at"`
	SessionID string  `json:"session_id"`
	TokenType string  `json:"token_type"`
}

func (h handler) newSocialCallbackResp(o auth.SocialCallbackOutput) *socialCallbackResp {
	return &socialCallbackResp{
		User: userObj{
			ID:    o.User.ID.Hex(),
			Email: o.User.Email,
			Name:  o.User.FullName,
			Role:  respObj{ID: o.Role.ID.Hex(), Name: o.Role.Name},
		},
		Token:     o.Token.AccessToken,
		ExpiresAt: util.FormatTime(o.Token.ExpiresAt, util.DateTimeFormat),
		SessionID: o.Token.SessionID,
		TokenType: o.Token.TokenType,
	}
}
