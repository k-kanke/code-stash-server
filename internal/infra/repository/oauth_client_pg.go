package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type oauthClientPGRepository struct {
	db *sql.DB
}

func NewOAuthClientPGRepository(db *sql.DB) usecaseRepo.OAuthClientRepository {
	return &oauthClientPGRepository{db: db}
}

func (r *oauthClientPGRepository) FindByID(ctx context.Context, id string) (*entity.OAuthClient, error) {
	const query = `
SELECT id, name, secret, grant_types, created_at
FROM oauth_clients
WHERE id = $1`

	var client entity.OAuthClient
	var grantTypes pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&client.ID,
		&client.Name,
		&client.Secret,
		&grantTypes,
		&client.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if grantTypes != nil {
		client.GrantTypes = append([]string(nil), grantTypes...)
	}

	return &client, nil
}
