package router

import (
	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller"
)

func RegisterRouter(e *echo.Echo, h *controller.Handler, auth *controller.AuthHandler, oauth *controller.OAuthHandler, authMiddleware echo.MiddlewareFunc) {
	api := e.Group("/api")

	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)

	oauthGroup := e.Group("/oauth")
	oauthGroup.POST("/device/code", oauth.CreateDeviceCode)
	oauthProtected := oauthGroup.Group("")
	oauthProtected.Use(authMiddleware)
	oauthProtected.GET("/device/verify", oauth.GetDeviceCodeStatus)
	oauthProtected.POST("/device/verify", oauth.VerifyDeviceCode)

	protected := api.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/collections", h.ListCollections)
		protected.GET("/collections/:id", h.GetCollection)
		protected.POST("/collections", h.CreateCollection)

		protected.GET("/collections/:id/folders", h.ListFolders)
		protected.POST("/collections/:id/folders", h.CreateFolder)
		protected.PATCH("/collections/:id/folders/:folderId", h.UpdateFolder)
		protected.DELETE("/collections/:id/folders/:folderId", h.DeleteFolder)

		protected.GET("/collections/:id/notes", h.ListNotes)
		protected.POST("/collections/:id/notes", h.CreateNote)
		protected.GET("/note/:id", h.GetNote)
		protected.PATCH("/note/:id", h.UpdateNote)
		protected.DELETE("/note/:id", h.DeleteNote)
		protected.GET("/note/:id/comments", h.ListNoteComments)
		protected.POST("/note/:id/comments", h.CreateNoteComment)
		protected.PATCH("/comments/:id", h.UpdateNoteComment)
		protected.DELETE("/comments/:id", h.DeleteNoteComment)
	}
}
