package dto

import "time"

type NoteComment struct {
	ID        string    `json:"id"`
	NoteID    string    `json:"noteId"`
	AuthorID  string    `json:"authorId"`
	Body      string    `json:"body"`
	LineStart *int      `json:"lineStart,omitempty"`
	LineEnd   *int      `json:"lineEnd,omitempty"`
	Resolved  bool      `json:"resolved"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateCommentRequest struct {
	Body      string `json:"body"`
	LineStart *int   `json:"lineStart,omitempty"`
	LineEnd   *int   `json:"lineEnd,omitempty"`
}

type UpdateCommentRequest struct {
	Body      *string `json:"body,omitempty"`
	LineStart *int    `json:"lineStart,omitempty"`
	LineEnd   *int    `json:"lineEnd,omitempty"`
	Resolved  *bool   `json:"resolved,omitempty"`
}
