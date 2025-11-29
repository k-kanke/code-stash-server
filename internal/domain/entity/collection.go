package entity

import "time"

type CollectionEntity struct {
	ID          string
	UserID      string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
