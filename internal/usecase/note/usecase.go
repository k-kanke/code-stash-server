package note

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	"github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type Usecase struct {
	repo repository.NoteRepository
}

func NewUsecase(repo repository.NoteRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) List(ctx context.Context, userID, collectionID string) ([]entity.Note, error) {
	return uc.repo.ListByCollection(ctx, userID, collectionID)
}
