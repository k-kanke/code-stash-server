package repository

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type CollectionRepository interface {
	List(ctx context.Context, userID string) ([]entity.Collection, error)
	Create(ctx context.Context, userID, name, description string) error
}
