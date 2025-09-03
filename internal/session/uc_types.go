package session

import (
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

type CreateSessionInput struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	UserAgent    string
	IPAddress    string
	DeviceName   string
	ExpiresAt    time.Time
}

type CreateSessionOutput struct {
	Session models.Session
}
