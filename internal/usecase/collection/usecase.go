package collection

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	"github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

// Usecase aggregates collection-specific application logic.
type Usecase struct {
	repo repository.CollectionRepository
}

func NewUsecase(repo repository.CollectionRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) List(ctx context.Context, userID string) ([]entity.Collection, error) {
	return uc.repo.List(ctx, userID)
}

func (uc *Usecase) Get(ctx context.Context, userID, collectionID string) (*entity.Collection, error) {
	return uc.repo.Get(ctx, userID, collectionID)
}

func (uc *Usecase) Create(ctx context.Context, userID, name, description string) error {
	return uc.repo.Create(ctx, userID, name, description)
}
