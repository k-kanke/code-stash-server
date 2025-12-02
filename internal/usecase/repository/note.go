package repository

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type NoteRepository interface {
	ListByCollection(ctx context.Context, userID, collectionID string) ([]entity.Note, error)
}
