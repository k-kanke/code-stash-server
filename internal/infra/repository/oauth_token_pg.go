package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type oauthTokenPGRepository struct {
	db *sql.DB
}

func NewOAuthTokenPGRepository(db *sql.DB) usecaseRepo.OAuthTokenRepository {
	return &oauthTokenPGRepository{db: db}
}

func (r *oauthTokenPGRepository) Create(ctx context.Context, token entity.OAuthToken) error {
	const query = `
INSERT INTO oauth_tokens (
	id, user_id, client_id, access_token, refresh_token, scope, expires_at, created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, now(), now()
)`

	id := token.ID
	if strings.TrimSpace(id) == "" {
		id = uuid.NewString()
	}

	var refresh any
	if token.RefreshToken != nil && *token.RefreshToken != "" {
		refresh = *token.RefreshToken
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		id,
		token.UserID,
		token.ClientID,
		token.AccessToken,
		refresh,
		pq.StringArray(token.Scope),
		token.ExpiresAt,
	)
	return err
}
