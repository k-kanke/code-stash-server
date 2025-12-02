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

type Handler struct {
	collectionUsecase CollectionUsecase
	folderUsecase     FolderUsecase
}

type Dependencies struct {
	Collections CollectionUsecase
	Folders     FolderUsecase
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		collectionUsecase: deps.Collections,
		folderUsecase:     deps.Folders,
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
