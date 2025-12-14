package note

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	"github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type CommentUsecase struct {
	repo repository.NoteCommentRepository
}

func NewCommentUsecase(repo repository.NoteCommentRepository) *CommentUsecase {
	return &CommentUsecase{repo: repo}
}

func (uc *CommentUsecase) ListByNote(ctx context.Context, userID, noteID string) ([]entity.NoteComment, error) {
	return uc.repo.ListByNote(ctx, userID, noteID)
}

type CommentCreateInput struct {
	UserID          string
	NoteID          string
	Body            string
	LineStart       *int
	LineEnd         *int
	ParentCommentID *string
}

func (uc *CommentUsecase) Create(ctx context.Context, in CommentCreateInput) (*entity.NoteComment, error) {
	payload := repository.CreateNoteCommentInput{
		UserID:          in.UserID,
		NoteID:          in.NoteID,
		Body:            in.Body,
		LineStart:       in.LineStart,
		LineEnd:         in.LineEnd,
		ParentCommentID: in.ParentCommentID,
	}
	return uc.repo.Create(ctx, payload)
}

type CommentUpdateInput struct {
	UserID    string
	CommentID string
	Body      *string
	LineStart *int
	LineEnd   *int
	Resolved  *bool
}

func (uc *CommentUsecase) Update(ctx context.Context, in CommentUpdateInput) (*entity.NoteComment, error) {
	payload := repository.UpdateNoteCommentInput{
		UserID:    in.UserID,
		CommentID: in.CommentID,
		Body:      in.Body,
		LineStart: in.LineStart,
		LineEnd:   in.LineEnd,
		Resolved:  in.Resolved,
	}
	return uc.repo.Update(ctx, payload)
}

func (uc *CommentUsecase) Delete(ctx context.Context, userID, commentID string) error {
	return uc.repo.Delete(ctx, userID, commentID)
}
