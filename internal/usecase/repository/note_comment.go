package repository

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type NoteCommentRepository interface {
	ListByNote(ctx context.Context, userID, noteID string) ([]entity.NoteComment, error)
	Create(ctx context.Context, in CreateNoteCommentInput) (*entity.NoteComment, error)
	Update(ctx context.Context, in UpdateNoteCommentInput) (*entity.NoteComment, error)
	Delete(ctx context.Context, userID, commentID string) error
}

type CreateNoteCommentInput struct {
	UserID    string
	NoteID    string
	Body      string
	LineStart *int
	LineEnd   *int
}

type UpdateNoteCommentInput struct {
	UserID    string
	CommentID string
	Body      *string
	LineStart *int
	LineEnd   *int
	Resolved  *bool
}
