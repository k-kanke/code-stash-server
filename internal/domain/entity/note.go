package entity

import "time"

type Note struct {
	ID           string
	UserID       string
	CollectionID string
	FolderID     *string
	Title        string
	Language     string
	Tags         []string
	Code         string
	Note         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
