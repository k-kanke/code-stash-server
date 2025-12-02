//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller"
	infraRepo "github.com/k-kanke/code-stash-server/internal/infra/repository"
	"github.com/k-kanke/code-stash-server/internal/usecase/collection"
	"github.com/k-kanke/code-stash-server/internal/usecase/folder"
)

func InitializeHandler(db *sql.DB) (*controller.Handler, error) {
	wire.Build(
		infraRepo.NewCollectionPGRepository,
		infraRepo.NewFolderPGRepository,
		collection.NewUsecase,
		folder.NewUsecase,
		wire.Bind(new(controller.CollectionUsecase), new(*collection.Usecase)),
		wire.Bind(new(controller.FolderUsecase), new(*folder.Usecase)),
		wire.Struct(new(controller.Dependencies), "Collections", "Folders"),
		controller.NewHandler,
	)
	return nil, nil
}
