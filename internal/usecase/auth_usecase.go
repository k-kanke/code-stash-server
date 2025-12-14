package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	userRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Usecase struct {
	repo userRepo.UserRepository
}

func NewUsecase(repo userRepo.UserRepository) *Usecase {
	return &Usecase{repo: repo}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

func (uc *Usecase) Register(ctx context.Context, in RegisterInput) (*entity.User, error) {
	name := strings.TrimSpace(in.Name)
	email := normalizeEmail(in.Email)
	password := strings.TrimSpace(in.Password)

	if name == "" || email == "" || password == "" {
		return nil, errors.New("invalid input")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := entity.User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hashed),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}

	return &user, nil
}

func (uc *Usecase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	email = normalizeEmail(email)
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return true
	}
	return false
}
