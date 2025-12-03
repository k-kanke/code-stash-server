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

type UpdateParams struct {
	FolderID *string
	Title    *string
	Language *string
	Code     *string
	Note     *string
	Tags     *[]string
}

func (uc *Usecase) List(ctx context.Context, userID, collectionID string) ([]entity.Note, error) {
	return uc.repo.ListByCollection(ctx, userID, collectionID)
}

func (uc *Usecase) Get(ctx context.Context, userID, noteID string) (*entity.Note, error) {
	return uc.repo.Get(ctx, userID, noteID)
}

func (uc *Usecase) Create(ctx context.Context, userID, collectionID string, folderID *string, title, code, language, note string, tags []string) error {
	return uc.repo.Create(ctx, userID, collectionID, folderID, title, code, language, note, tags)
}

func (uc *Usecase) Update(ctx context.Context, userID, noteID string, params UpdateParams) error {
	update := repository.NoteUpdate{
		FolderID: params.FolderID,
		Title:    params.Title,
		Language: params.Language,
		Code:     params.Code,
		Note:     params.Note,
		Tags:     params.Tags,
	}
	return uc.repo.Update(ctx, userID, noteID, update)
}
