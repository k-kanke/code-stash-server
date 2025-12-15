package repository

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type OAuthClientRepository interface {
	FindByID(ctx context.Context, id string) (*entity.OAuthClient, error)
}
