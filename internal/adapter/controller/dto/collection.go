package dto

import "time"

type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	NoteCount   int       `json:"note_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCollectionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}
