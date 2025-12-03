package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller/dto"
	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type CollectionUsecase interface {
	List(ctx context.Context, userID string) ([]entity.Collection, error)
	Get(ctx context.Context, userID, collectionID string) (*entity.Collection, error)
	Create(ctx context.Context, userID, name, description string) error
}

type FolderUsecase interface {
	List(ctx context.Context, userID, collectionID string) ([]entity.Folder, error)
	Create(ctx context.Context, userID, collectionID string, parentFolderID *string, name string) error
}

type NoteUsecase interface {
	List(ctx context.Context, userID, collectionID string) ([]entity.Note, error)
	Get(ctx context.Context, userID, noteID string) (*entity.Note, error)
	Create(ctx context.Context, userID, collectionID string, folderID *string, title, code, language, note string, tags []string) error
}

type Handler struct {
	collectionUsecase CollectionUsecase
	folderUsecase     FolderUsecase
	noteUsecase       NoteUsecase
}

type Dependencies struct {
	Collections CollectionUsecase
	Folders     FolderUsecase
	Notes       NoteUsecase
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		collectionUsecase: deps.Collections,
		folderUsecase:     deps.Folders,
		noteUsecase:       deps.Notes,
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create folder",
		})
	}

	return c.NoContent(http.StatusCreated)
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

	if err := h.noteUsecase.Create(
		c.Request().Context(),
		userID,
		collectionID,
		folderID,
		title,
		code,
		language,
		req.Note,
		req.Tags,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "collection or folder not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create note",
		})
	}

	return c.NoContent(http.StatusCreated)
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
