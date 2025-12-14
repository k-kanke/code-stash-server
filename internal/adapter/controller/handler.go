package controller

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller/dto"
	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	folderUsecase "github.com/k-kanke/code-stash-server/internal/usecase/folder"
	noteUsecase "github.com/k-kanke/code-stash-server/internal/usecase/note"
)

type CollectionUsecase interface {
	List(ctx context.Context, userID string) ([]entity.Collection, error)
	Get(ctx context.Context, userID, collectionID string) (*entity.Collection, error)
	Create(ctx context.Context, userID, name, description string) error
}

type FolderUsecase interface {
	List(ctx context.Context, userID, collectionID string) ([]entity.Folder, error)
	Create(ctx context.Context, userID, collectionID string, parentFolderID *string, name string) error
	Delete(ctx context.Context, userID, collectionID, folderID string) error
}

type NoteUsecase interface {
	List(ctx context.Context, userID, collectionID string) ([]entity.Note, error)
	Get(ctx context.Context, userID, noteID string) (*entity.Note, error)
	Create(ctx context.Context, in noteUsecase.CreateInput) error
	Update(ctx context.Context, in noteUsecase.UpdateInput) error
	Delete(ctx context.Context, in noteUsecase.DeleteInput) error
}

type NoteCommentUsecase interface {
	ListByNote(ctx context.Context, userID, noteID string) ([]entity.NoteComment, error)
	Create(ctx context.Context, in noteUsecase.CommentCreateInput) (*entity.NoteComment, error)
	Update(ctx context.Context, in noteUsecase.CommentUpdateInput) (*entity.NoteComment, error)
	Delete(ctx context.Context, userID, commentID string) error
}

type Handler struct {
	collectionUsecase CollectionUsecase
	folderUsecase     FolderUsecase
	noteUsecase       NoteUsecase
	commentUsecase    NoteCommentUsecase
}

type Dependencies struct {
	Collections CollectionUsecase
	Folders     FolderUsecase
	Notes       NoteUsecase
	Comments    NoteCommentUsecase
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		collectionUsecase: deps.Collections,
		folderUsecase:     deps.Folders,
		noteUsecase:       deps.Notes,
		commentUsecase:    deps.Comments,
	}
}

func (h *Handler) ListCollections(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collections, err := h.collectionUsecase.List(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch collections",
		})
	}

	resp := make([]dto.Collection, 0, len(collections))
	for _, col := range collections {
		resp = append(resp, dto.Collection{
			ID:          col.ID,
			Name:        col.Name,
			Description: col.Description,
			NoteCount:   col.NoteCount,
			CreatedAt:   col.CreatedAt,
			UpdatedAt:   col.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateCollection(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	var req dto.CreateCollectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
	}

	if err := h.collectionUsecase.Create(c.Request().Context(), userID, req.Name, req.Description); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create collection",
		})
	}

	return c.NoContent(http.StatusCreated)
}

func (h *Handler) GetCollection(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "collection id is required",
		})
	}

	collection, err := h.collectionUsecase.Get(c.Request().Context(), userID, collectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "collection not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch collection",
		})
	}

	resp := dto.Collection{
		ID:          collection.ID,
		Name:        collection.Name,
		Description: collection.Description,
		NoteCount:   collection.NoteCount,
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListFolders(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "collection id is required",
		})
	}

	folders, err := h.folderUsecase.List(c.Request().Context(), userID, collectionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch folders",
		})
	}

	resp := make([]dto.Folder, 0, len(folders))
	for _, f := range folders {
		resp = append(resp, dto.Folder{
			ID:             f.ID,
			CollectionID:   f.CollectionID,
			ParentFolderID: f.ParentFolderID,
			Name:           f.Name,
			SortOrder:      f.SortOrder,
			CreatedAt:      f.CreatedAt,
			UpdatedAt:      f.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateFolder(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "collection id is required",
		})
	}

	var req dto.CreateFolderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
	}

	if err := h.folderUsecase.Create(c.Request().Context(), userID, collectionID, req.ParentFolderID, req.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "collection not found",
			})
		}
		if errors.Is(err, folderUsecase.ErrNameConflict) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "a folder or note with the same name already exists in this location",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create folder",
		})
	}

	return c.NoContent(http.StatusCreated)
}

func (h *Handler) DeleteFolder(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "collection id is required",
		})
	}

	folderID := strings.TrimSpace(c.Param("folderId"))
	if folderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "folder id is required",
		})
	}

	if err := h.folderUsecase.Delete(c.Request().Context(), userID, collectionID, folderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "folder not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to delete folder",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListNotes(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "collection id is required",
		})
	}

	notes, err := h.noteUsecase.List(c.Request().Context(), userID, collectionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch notes",
		})
	}

	resp := make([]dto.NoteSummary, 0, len(notes))
	for _, n := range notes {
		resp = append(resp, dto.NoteSummary{
			ID:        n.ID,
			Title:     n.Title,
			Language:  n.Language,
			Tags:      n.Tags,
			Snippet:   buildSnippet(n),
			FolderID:  n.FolderID,
			UpdatedAt: n.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetNote(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	noteID := strings.TrimSpace(c.Param("id"))
	if noteID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "note id is required",
		})
	}

	note, err := h.noteUsecase.Get(c.Request().Context(), userID, noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "note not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch note",
		})
	}

	resp := dto.NoteDetail{
		ID:           note.ID,
		CollectionID: note.CollectionID,
		FolderID:     note.FolderID,
		Title:        note.Title,
		Language:     note.Language,
		Tags:         note.Tags,
		Code:         note.Code,
		Note:         note.Note,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateNote(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	noteID := strings.TrimSpace(c.Param("id"))
	if noteID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "note id is required",
		})
	}

	var req dto.UpdateNoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	params := noteUsecase.UpdateInput{
		UserID: userID,
		NoteID: noteID,
	}
	hasUpdate := false

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "title cannot be empty",
			})
		}
		params.Title = &title
		hasUpdate = true
	}

	if req.Language != nil {
		language := strings.TrimSpace(*req.Language)
		if language == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "language cannot be empty",
			})
		}
		params.Language = &language
		hasUpdate = true
	}

	if req.Code != nil {
		code := strings.TrimSpace(*req.Code)
		if code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "code cannot be empty",
			})
		}
		params.Code = &code
		hasUpdate = true
	}

	if req.Note != nil {
		params.Note = req.Note
		hasUpdate = true
	}

	if req.FolderID != nil {
		folderID := strings.TrimSpace(*req.FolderID)
		params.FolderID = &folderID
		hasUpdate = true
	}

	if req.Tags != nil {
		tags := req.Tags
		params.Tags = &tags
		hasUpdate = true
	}

	if !hasUpdate {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "no fields to update",
		})
	}

	if err := h.noteUsecase.Update(c.Request().Context(), params); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "note not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to update note",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteNote(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	noteID := strings.TrimSpace(c.Param("id"))
	if noteID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "note id is required",
		})
	}

	in := noteUsecase.DeleteInput{
		UserID: userID,
		NoteID: noteID,
	}

	if err := h.noteUsecase.Delete(c.Request().Context(), in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "note not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to delete note",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListNoteComments(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	noteID := strings.TrimSpace(c.Param("id"))
	if noteID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "note id is required",
		})
	}

	comments, err := h.commentUsecase.ListByNote(c.Request().Context(), userID, noteID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch comments",
		})
	}

	resp := make([]dto.NoteComment, 0, len(comments))
	for _, comment := range comments {
		resp = append(resp, toCommentDTO(comment))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateNoteComment(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	noteID := strings.TrimSpace(c.Param("id"))
	if noteID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "note id is required",
		})
	}

	var req dto.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "body is required",
		})
	}

	var parentCommentID *string
	if req.ParentCommentID != nil {
		parent := strings.TrimSpace(*req.ParentCommentID)
		if parent != "" {
			parentCommentID = &parent
		}
	}

	comment, err := h.commentUsecase.Create(c.Request().Context(), noteUsecase.CommentCreateInput{
		UserID:          userID,
		NoteID:          noteID,
		Body:            req.Body,
		LineStart:       req.LineStart,
		LineEnd:         req.LineEnd,
		ParentCommentID: parentCommentID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "note not found",
			})
		}
		log.Printf("failed to create comment: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create comment",
		})
	}

	return c.JSON(http.StatusCreated, toCommentDTO(*comment))
}

func (h *Handler) UpdateNoteComment(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	commentID := strings.TrimSpace(c.Param("id"))
	if commentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "comment id is required",
		})
	}

	var req dto.UpdateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.Body != nil {
		trimmed := strings.TrimSpace(*req.Body)
		if trimmed == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "body cannot be empty",
			})
		}
		req.Body = &trimmed
	}

	if req.Body == nil && req.LineStart == nil && req.LineEnd == nil && req.Resolved == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "no update fields provided",
		})
	}

	comment, err := h.commentUsecase.Update(c.Request().Context(), noteUsecase.CommentUpdateInput{
		UserID:    userID,
		CommentID: commentID,
		Body:      req.Body,
		LineStart: req.LineStart,
		LineEnd:   req.LineEnd,
		Resolved:  req.Resolved,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "comment not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to update comment",
		})
	}

	return c.JSON(http.StatusOK, toCommentDTO(*comment))
}

func (h *Handler) DeleteNoteComment(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	commentID := strings.TrimSpace(c.Param("id"))
	if commentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "comment id is required",
		})
	}

	if err := h.commentUsecase.Delete(c.Request().Context(), userID, commentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "comment not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to delete comment",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) CreateNote(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	collectionID := strings.TrimSpace(c.Param("id"))
	if collectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "collection id is required",
		})
	}

	var req dto.CreateNoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "title is required",
		})
	}

	language := strings.TrimSpace(req.Language)
	if language == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "language is required",
		})
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "code is required",
		})
	}

	var folderID *string
	if req.FolderID != nil {
		if trimmed := strings.TrimSpace(*req.FolderID); trimmed != "" {
			folder := trimmed
			folderID = &folder
		}
	}

	in := noteUsecase.CreateInput{
		UserID:       userID,
		CollectionID: collectionID,
		FolderID:     folderID,
		Title:        title,
		Language:     language,
		Code:         code,
		Note:         req.Note,
		Tags:         req.Tags,
	}

	if err := h.noteUsecase.Create(c.Request().Context(), in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "collection or folder not found",
			})
		}
		if errors.Is(err, noteUsecase.ErrTitleConflict) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "a folder or note with the same name already exists in this location",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create note",
		})
	}

	return c.NoContent(http.StatusCreated)
}

func toCommentDTO(comment entity.NoteComment) dto.NoteComment {
	return dto.NoteComment{
		ID:              comment.ID,
		NoteID:          comment.NoteID,
		AuthorID:        comment.AuthorID,
		Body:            comment.Body,
		LineStart:       comment.LineStart,
		LineEnd:         comment.LineEnd,
		ParentCommentID: comment.ParentCommentID,
		Resolved:        comment.Resolved,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
	}
}

func buildSnippet(n entity.Note) string {
	if snippet := strings.TrimSpace(n.Note); snippet != "" {
		return truncate(snippet, 120)
	}
	if snippet := strings.TrimSpace(n.Code); snippet != "" {
		return truncate(snippet, 120)
	}
	return ""
}

func truncate(src string, max int) string {
	runes := []rune(src)
	if len(runes) <= max {
		return src
	}
	return string(runes[:max])
}
