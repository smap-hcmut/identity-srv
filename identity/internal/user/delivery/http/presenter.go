package http

import (
	"smap-api/internal/model"
	"smap-api/internal/user"
	"smap-api/pkg/paginator"
	"smap-api/pkg/response"
)

// Request DTOs
type UpdateProfileReq struct {
	FullName  string `json:"full_name" binding:"required"`
	AvatarURL string `json:"avatar_url"`
}

func (r UpdateProfileReq) toInput() user.UpdateProfileInput {
	return user.UpdateProfileInput{
		FullName:  r.FullName,
		AvatarURL: r.AvatarURL,
	}
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (r ChangePasswordReq) toInput() user.ChangePasswordInput {
	return user.ChangePasswordInput{
		OldPassword: r.OldPassword,
		NewPassword: r.NewPassword,
	}
}

type ListReq struct {
	IDs []string `form:"ids[]"`
}

func (r ListReq) toInput() user.ListInput {
	return user.ListInput{
		Filter: user.Filter{
			IDs: r.IDs,
		},
	}
}

type GetReq struct {
	paginator.PaginateQuery
	IDs []string `form:"ids[]"`
}

func (r GetReq) toInput() user.GetInput {
	return user.GetInput{
		Filter: user.Filter{
			IDs: r.IDs,
		},
		PaginateQuery: paginator.PaginateQuery{
			Page:  r.Page,
			Limit: r.Limit,
		},
	}
}

// Response DTOs
type UserResp struct {
	ID        string            `json:"id"`
	Username  string            `json:"username"`
	FullName  *string           `json:"full_name,omitempty"`
	AvatarURL *string           `json:"avatar_url,omitempty"`
	Role      string            `json:"role,omitempty"`
	IsActive  *bool             `json:"is_active,omitempty"`
	CreatedAt response.DateTime `json:"created_at"`
	UpdatedAt response.DateTime `json:"updated_at"`
}

type ListUserResp struct {
	Users []UserResp `json:"users"`
}

type GetUserResp struct {
	Users    []UserResp                  `json:"users"`
	Metadata paginator.PaginatorResponse `json:"metadata"`
}

// Converters
func (h handler) newUserResp(o user.UserOutput) UserResp {
	role := o.User.GetRole()
	return UserResp{
		ID:        o.User.ID,
		Username:  o.User.Username,
		FullName:  o.User.FullName,
		AvatarURL: o.User.AvatarURL,
		Role:      role,
		IsActive:  o.User.IsActive,
		CreatedAt: response.DateTime(o.User.CreatedAt),
		UpdatedAt: response.DateTime(o.User.UpdatedAt),
	}
}

func (h handler) newListUserResp(us []model.User) ListUserResp {
	resp := ListUserResp{
		Users: make([]UserResp, 0, len(us)),
	}

	for _, u := range us {
		resp.Users = append(resp.Users, UserResp{
			ID:        u.ID,
			Username:  u.Username,
			FullName:  u.FullName,
			AvatarURL: u.AvatarURL,
			Role:      u.GetRole(),
			IsActive:  u.IsActive,
			CreatedAt: response.DateTime(u.CreatedAt),
			UpdatedAt: response.DateTime(u.UpdatedAt),
		})
	}

	return resp
}

func (h handler) newGetUserResp(o user.GetUserOutput) GetUserResp {
	resp := GetUserResp{
		Users:    make([]UserResp, 0, len(o.Users)),
		Metadata: o.Paginator.ToResponse(),
	}

	for _, u := range o.Users {
		resp.Users = append(resp.Users, UserResp{
			ID:        u.ID,
			Username:  u.Username,
			FullName:  u.FullName,
			AvatarURL: u.AvatarURL,
			Role:      u.GetRole(),
			IsActive:  u.IsActive,
			CreatedAt: response.DateTime(u.CreatedAt),
			UpdatedAt: response.DateTime(u.UpdatedAt),
		})
	}

	return resp
}
