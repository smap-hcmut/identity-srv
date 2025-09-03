package session

import "time"

type CreateSessionOptions struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	UserAgent    string
	IPAddress    string
	DeviceName   string
	ExpiresAt    time.Time
}
