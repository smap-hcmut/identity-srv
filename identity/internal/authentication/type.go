package authentication

import (
	"smap-api/internal/model"
)

// Register
type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	User model.User
}

// Send OTP
type SendOTPInput struct {
	Email    string
	Password string
}

// Verify OTP
type VerifyOTPInput struct {
	Email string
	OTP   string
}

// Login
type LoginInput struct {
	Email      string
	Password   string
	Remember   bool
	UserAgent  string
	IPAddress  string
	DeviceName string
}

type LoginOutput struct {
	User  model.User
	Token TokenOutput
}

type TokenOutput struct {
	AccessToken string
	TokenType   string
}

// GetCurrentUser
type GetCurrentUserOutput struct {
	User model.User
}

// Producer
type PublishSendEmailMsgInput struct {
	Subject     string
	Recipient   string
	Body        string
	ReplyTo     string
	CcAddresses []string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}
