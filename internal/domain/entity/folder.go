package entity

import "time"

type Folder struct {
	ID             string
	UserID         string
	CollectionID   string
	ParentFolderID *string
	Name           string
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
