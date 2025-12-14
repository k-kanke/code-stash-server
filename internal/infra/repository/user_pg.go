package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type userPGRepository struct {
	db *sql.DB
}

func NewUserPGRepository(db *sql.DB) usecaseRepo.UserRepository {
	return &userPGRepository{db: db}
}

func (r *userPGRepository) Create(ctx context.Context, user entity.User) error {
	const query = `
INSERT INTO users (id, name, email, password_hash)
VALUES ($1, $2, $3, $4)`

	id := user.ID
	if strings.TrimSpace(id) == "" {
		id = uuid.NewString()
	}

	_, err := r.db.ExecContext(ctx, query, id, user.Name, user.Email, user.PasswordHash)
	return err
}

func (r *userPGRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	const query = `
SELECT id, name, email, password_hash, created_at
FROM users
WHERE email = $1`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userPGRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	const query = `
SELECT id, name, email, password_hash, created_at
FROM users
WHERE id = $1`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
