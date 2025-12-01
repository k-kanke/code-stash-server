//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller"
	infraRepo "github.com/k-kanke/code-stash-server/internal/infra/repository"
	"github.com/k-kanke/code-stash-server/internal/usecase/collection"
)

func InitializeHandler(db *sql.DB) (*controller.Handler, error) {
	wire.Build(
		infraRepo.NewCollectionPGRepository,
		collection.NewUsecase,
		wire.Bind(new(controller.CollectionUsecase), new(*collection.Usecase)),
		wire.Struct(new(controller.Dependencies), "Collections"),
		controller.NewHandler,
	)
	return nil, nil
}
