package folder

import (
	"context"
	"errors"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	"github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

var ErrNameConflict = errors.New("folder name already exists in this location")

type Usecase struct {
	repo repository.FolderRepository
}

func NewUsecase(repo repository.FolderRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) List(ctx context.Context, userID, collectionID string) ([]entity.Folder, error) {
	return uc.repo.ListByCollection(ctx, userID, collectionID)
}

func (uc *Usecase) Create(ctx context.Context, userID, collectionID string, parentFolderID *string, name string) error {
	if err := uc.repo.Create(ctx, userID, collectionID, parentFolderID, name); err != nil {
		if errors.Is(err, repository.ErrFolderNameConflict) {
			return ErrNameConflict
		}
		return err
	}
	return nil
}

func (uc *Usecase) Delete(ctx context.Context, userID, collectionID, folderID string) error {
	return uc.repo.Delete(ctx, userID, collectionID, folderID)
}

func (uc *Usecase) Rename(ctx context.Context, userID, collectionID, folderID, name string) error {
	if err := uc.repo.UpdateName(ctx, userID, collectionID, folderID, name); err != nil {
		if errors.Is(err, repository.ErrFolderNameConflict) {
			return ErrNameConflict
		}
		return err
	}
	return nil
}
