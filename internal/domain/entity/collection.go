package entity

import "time"

type Collection struct {
	ID          string
	UserID      string
	Name        string
	Description string
	NoteCount   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
