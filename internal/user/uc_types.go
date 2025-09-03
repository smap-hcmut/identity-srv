package user

import (
	"strings"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/pkg/paginator"
)

type GetInput struct {
	Filter   UserFilter
	PagQuery paginator.PaginateQuery
}

type GetUserOutput struct {
	Users      []models.User
	Roles      []models.Role
	Pagination paginator.Paginator
}

type DetailOutput struct {
	User models.User
	Role models.Role
}

type GetOneInput struct {
	Email string
}

type CreateInput struct {
	Provider   string
	ProviderID string
	Email      string
	Password   string
	FullName   string
	AvatarURL  string
	IsVerified bool
}

type UserOutput struct {
	User models.User
	Role models.Role
}

type UpdateVerifiedInput struct {
	UserID       string
	OTP          string
	OTPExpiredAt time.Time
	IsVerified   bool
}

type UpdateAvatarInput struct {
	UserID    string
	AvatarURL string
}

var CommonImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tiff", ".tif", ".ico", ".heic", ".heif", ".raw", ".cr2", ".nef", ".arw"}

func IsValidImageExtension(imageURL string) bool {
	for _, ext := range CommonImageExtensions {
		if strings.HasSuffix(imageURL, ext) {
			return true
		}
	}
	return false
}
