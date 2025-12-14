package dto

import "time"

type Folder struct {
	ID             string    `json:"id"`
	CollectionID   string    `json:"collection_id"`
	ParentFolderID *string   `json:"parent_folder_id,omitempty"`
	Name           string    `json:"name"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateFolderRequest struct {
	ParentFolderID *string `json:"parent_folder_id"`
	Name           string  `json:"name" validate:"required"`
}

type UpdateFolderRequest struct {
	Name string `json:"name" validate:"required"`
}
