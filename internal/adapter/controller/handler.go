package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller/dto"
	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type CollectionUsecase interface {
	List(ctx context.Context, userID string) ([]entity.Collection, error)
}

type Handler struct {
	collectionUsecase CollectionUsecase
}

type Dependencies struct {
	Collections CollectionUsecase
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		collectionUsecase: deps.Collections,
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
