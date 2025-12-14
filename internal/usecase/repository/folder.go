package repository

import (
	"context"
	"errors"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

var ErrFolderNameConflict = errors.New("folder name already exists")

type FolderRepository interface {
	ListByCollection(ctx context.Context, userID, collectionID string) ([]entity.Folder, error)
	Create(ctx context.Context, userID, collectionID string, parentFolderID *string, name string) error
	Delete(ctx context.Context, userID, collectionID, folderID string) error
	UpdateName(ctx context.Context, userID, collectionID, folderID, name string) error
}
