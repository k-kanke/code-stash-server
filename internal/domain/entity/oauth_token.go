package entity

import "time"

type OAuthToken struct {
	ID           string
	UserID       string
	ClientID     string
	AccessToken  string
	RefreshToken *string
	Scope        []string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
