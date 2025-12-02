package folder

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	"github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type Usecase struct {
	repo repository.FolderRepository
}

func NewUsecase(repo repository.FolderRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) List(ctx context.Context, userID, collectionID string) ([]entity.Folder, error) {
	return uc.repo.ListByCollection(ctx, userID, collectionID)
}
