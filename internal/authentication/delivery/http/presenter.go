package http

import (
	"smap-api/internal/authentication"
)

type respObj struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type userObj struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName *string `json:"full_name,omitempty"`
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
