package repository

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type FolderRepository interface {
	ListByCollection(ctx context.Context, userID, collectionID string) ([]entity.Folder, error)
	Create(ctx context.Context, userID, collectionID string, parentFolderID *string, name string) error
}
