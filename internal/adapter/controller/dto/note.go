package dto

import "time"

type NoteSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Language  string    `json:"language"`
	Tags      []string  `json:"tags"`
	Snippet   string    `json:"snippet"`
	FolderID  *string   `json:"folder_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteDetail struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	FolderID     *string   `json:"folder_id"`
	Title        string    `json:"title"`
	Language     string    `json:"language"`
	Tags         []string  `json:"tags"`
	Code         string    `json:"code"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateNoteRequest struct {
	FolderID *string  `json:"folder_id"`
	Title    string   `json:"title" validate:"required"`
	Language string   `json:"language" validate:"required"`
	Tags     []string `json:"tags"`
	Code     string   `json:"code" validate:"required"`
	Note     string   `json:"note"`
}

type UpdateNoteRequest struct {
	FolderID *string  `json:"folder_id"`
	Title    *string  `json:"title"`
	Language *string  `json:"language"`
	Tags     []string `json:"tags"`
	Code     *string  `json:"code"`
	Note     *string  `json:"note"`
}
