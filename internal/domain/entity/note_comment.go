package entity

import "time"

type NoteComment struct {
	ID              string
	NoteID          string
	AuthorID        string
	Body            string
	LineStart       *int
	LineEnd         *int
	ParentCommentID *string
	Resolved        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
