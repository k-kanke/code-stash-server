package repository

import (
	"context"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type OAuthTokenRepository interface {
	Create(ctx context.Context, token entity.OAuthToken) error
}
