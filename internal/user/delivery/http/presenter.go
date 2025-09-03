package http

import "github.com/nguyentantai21042004/smap-api/internal/user"

type detailResp struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	FullName   string  `json:"full_name,omitempty"`
	IsVerified bool    `json:"is_verified"`
	AvatarURL  string  `json:"avatar_url,omitempty"`
	Provider   string  `json:"provider,omitempty"`
	Role       respObj `json:"role,omitempty"`
}

type respObj struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h handler) newDetailResp(o user.UserOutput) detailResp {

	resp := detailResp{
		ID:    o.User.ID.Hex(),
		Email: o.User.Email,
		Role: respObj{
			ID:   o.Role.ID.Hex(),
			Name: o.Role.Name,
		},
	}

	if o.User.FullName != "" {
		resp.FullName = o.User.FullName
	}

	if o.User.IsVerified {
		resp.IsVerified = o.User.IsVerified
	}

	if o.User.AvatarURL != "" {
		resp.AvatarURL = o.User.AvatarURL
	}

	if o.User.Provider != "" {
		resp.Provider = o.User.Provider
	}

	return resp
}

type updateAvatarReq struct {
	UserID    string `json:"user_id" binding:"required"`
	AvatarURL string `json:"avatar_url" binding:"required"`
}

func (r updateAvatarReq) validate() error {

	return nil
}

func (r updateAvatarReq) toInput() user.UpdateAvatarInput {
	return user.UpdateAvatarInput{
		UserID:    r.UserID,
		AvatarURL: r.AvatarURL,
	}
}

type updateAvatarResp struct {
	ID        string `json:"id"`
	AvatarURL string `json:"avatar_url"`
}
