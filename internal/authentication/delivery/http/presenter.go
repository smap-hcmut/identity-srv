package http

import (
	"smap-api/pkg/util"

	"smap-api/internal/authentication"
)

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
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName *string `json:"full_name,omitempty"`
}

type loginResp struct {
	User userObj `json:"user"`
}

func (h handler) newLoginResp(o authentication.LoginOutput) *loginResp {
	return &loginResp{
		User: userObj{
			ID:       o.User.ID,
			Email:    o.User.Email,
			FullName: o.User.Name,
		},
	}
}

type getMeResp struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName *string `json:"full_name,omitempty"`
}

func (h handler) newGetMeResp(o authentication.GetCurrentUserOutput) *getMeResp {
	return &getMeResp{
		ID:       o.User.ID,
		Email:    o.User.Email,
		FullName: o.User.Name,
	}
}
