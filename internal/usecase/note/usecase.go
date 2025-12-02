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

func (uc *Usecase) Create(ctx context.Context, userID, collectionID string, folderID *string, title, code, language, note string, tags []string) error {
	return uc.repo.Create(ctx, userID, collectionID, folderID, title, code, language, note, tags)
}
