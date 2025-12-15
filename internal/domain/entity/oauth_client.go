package entity

import "time"

type OAuthClient struct {
	ID         string
	Name       string
	Secret     string
	GrantTypes []string
	CreatedAt  time.Time
}
