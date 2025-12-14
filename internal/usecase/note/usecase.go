package note

import (
	"context"
	"errors"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	"github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

var ErrTitleConflict = errors.New("note title already exists in this location")

type Usecase struct {
	repo repository.NoteRepository
}

func NewUsecase(repo repository.NoteRepository) *Usecase {
	return &Usecase{repo: repo}
}

type CreateInput struct {
	UserID       string
	CollectionID string
	FolderID     *string
	Title        string
	Language     string
	Code         string
	Note         string
	Tags         []string
}

type UpdateInput struct {
	UserID   string
	NoteID   string
	FolderID *string
	Title    *string
	Language *string
	Code     *string
	Note     *string
	Tags     *[]string
}

type DeleteInput struct {
	UserID string
	NoteID string
}

func (uc *Usecase) List(ctx context.Context, userID, collectionID string) ([]entity.Note, error) {
	return uc.repo.ListByCollection(ctx, userID, collectionID)
}

func (uc *Usecase) Get(ctx context.Context, userID, noteID string) (*entity.Note, error) {
	return uc.repo.Get(ctx, userID, noteID)
}

func (uc *Usecase) Create(ctx context.Context, in CreateInput) error {
	if err := uc.repo.Create(ctx, in.UserID, in.CollectionID, in.FolderID, in.Title, in.Code, in.Language, in.Note, in.Tags); err != nil {
		if errors.Is(err, repository.ErrNoteTitleConflict) {
			return ErrTitleConflict
		}
		return err
	}
	return nil
}

func (uc *Usecase) Update(ctx context.Context, in UpdateInput) error {
	update := repository.NoteUpdate{
		FolderID: in.FolderID,
		Title:    in.Title,
		Language: in.Language,
		Code:     in.Code,
		Note:     in.Note,
		Tags:     in.Tags,
	}
	return uc.repo.Update(ctx, in.UserID, in.NoteID, update)
}

func (uc *Usecase) Delete(ctx context.Context, in DeleteInput) error {
	return uc.repo.Delete(ctx, in.UserID, in.NoteID)
}
