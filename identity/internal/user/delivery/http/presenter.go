package http

import (
	"smap-api/internal/model"
	"smap-api/internal/user"
	"smap-api/pkg/paginator"
	"smap-api/pkg/response"
)

// Request DTOs

type UpdateProfileRequest struct {
	FullName  string `json:"full_name" binding:"required"`
	AvatarURL string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ListRequest struct {
	IDs []string `form:"ids[]"`
}

type GetRequest struct {
	paginator.PaginateQuery
	IDs []string `form:"ids[]"`
}

// Response DTOs

type UserResponse struct {
	ID        string            `json:"id"`
	Username  string            `json:"username"`
	FullName  *string           `json:"full_name,omitempty"`
	AvatarURL *string           `json:"avatar_url,omitempty"`
	Role      string            `json:"role,omitempty"`
	IsActive  *bool             `json:"is_active,omitempty"`
	CreatedAt response.DateTime `json:"created_at"`
	UpdatedAt response.DateTime `json:"updated_at"`
}

type ListUserResponse struct {
	Users []UserResponse `json:"users"`
}

type GetUserResponse struct {
	Users     []UserResponse      `json:"users"`
	Paginator paginator.Paginator `json:"paginator"`
}

// Converters

func toUserResponse(u model.User) UserResponse {
	role := u.GetRole()
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Role:      role,
		IsActive:  u.IsActive,
		CreatedAt: response.DateTime(u.CreatedAt),
		UpdatedAt: response.DateTime(u.UpdatedAt),
	}
}

func toListUserResponse(users []model.User) ListUserResponse {
	resp := ListUserResponse{
		Users: make([]UserResponse, 0, len(users)),
	}

	for _, u := range users {
		resp.Users = append(resp.Users, toUserResponse(u))
	}

	return resp
}

func toGetUserResponse(output user.GetUserOutput) GetUserResponse {
	resp := GetUserResponse{
		Users:     make([]UserResponse, 0, len(output.Users)),
		Paginator: output.Paginator,
	}

	for _, u := range output.Users {
		resp.Users = append(resp.Users, toUserResponse(u))
	}

	return resp
}
