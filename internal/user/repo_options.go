package user

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

type UserFilter struct {
	IDs       []string
	RoleIDs   []string
	Providers []string
	Emails    []string
	NameCodes []string
}

type ListOptions struct {
	Filter UserFilter
}

type GetOptions struct {
	Filter   UserFilter
	PagQuery paginator.PaginateQuery
}

type GetOneOptions struct {
	Email string
}

type CreateOptions struct {
	Email        string
	Password     string
	FullName     string
	IsVerified   bool
	AvatarURL    string
	Provider     string
	ProviderID   string
	OTP          string
	OTPExpiredAt time.Time
	RoleID       string
}

type UpdateVerifiedOptions struct {
	ID           string
	OTP          string
	OTPExpiredAt time.Time
	IsVerified   bool
}

type UpdateAvatarOptions struct {
	ID        string
	AvatarURL string
}

type DeleteOptions struct {
	IDs []string
}
