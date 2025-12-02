package router

import (
	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller"
)

func RegisterRouter(e *echo.Echo, h *controller.Handler) {
	api := e.Group("/api")
	// 認証ミドルウェア

	// Collections
	api.GET("/collections", h.ListCollections)
	api.GET("/collections/:id", h.GetCollection)
	api.POST("/collections", h.CreateCollection)

	// Folders
	api.GET("/collections/:id/folders", h.ListFolders)
	api.POST("/collections/:id/folders", h.CreateFolder)

	// Notes
	api.GET("/collections/:id/notes", h.ListNotes)
	api.POST("/collections/:id/notes", h.CreateNote)
}
