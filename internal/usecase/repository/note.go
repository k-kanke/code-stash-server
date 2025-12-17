package repository

import (
	"context"
	"errors"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

var ErrNoteTitleConflict = errors.New("note title already exists in this folder")

type NoteRepository interface {
	ListByCollection(ctx context.Context, userID, collectionID string) ([]entity.Note, error)
	Get(ctx context.Context, userID, noteID string) (*entity.Note, error)
	Create(ctx context.Context, userID, collectionID string, folderID *string, title, code, language, note string, tags []string) (string, error)
	Update(ctx context.Context, userID, noteID string, update NoteUpdate) error
	Delete(ctx context.Context, userID, noteID string) error
}

type NoteUpdate struct {
	FolderID *string
	Title    *string
	Language *string
	Code     *string
	Note     *string
	Tags     *[]string
}
