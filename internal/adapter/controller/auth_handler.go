package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller/dto"
	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	authusecase "github.com/k-kanke/code-stash-server/internal/usecase"
)

type TokenService interface {
	Generate(userID string) (string, error)
}

type AuthDependencies struct {
	Usecase      *authusecase.Usecase
	TokenService TokenService
}

type AuthHandler struct {
	authUsecase  *authusecase.Usecase
	tokenService TokenService
}

func NewAuthHandler(deps AuthDependencies) *AuthHandler {
	return &AuthHandler{
		authUsecase:  deps.Usecase,
		tokenService: deps.TokenService,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	user, err := h.authUsecase.Register(c.Request().Context(), authusecase.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == authusecase.ErrEmailAlreadyExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already registered"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	token, err := h.tokenService.Generate(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
	}

	return c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: token,
		User:  toAuthUser(*user),
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	user, err := h.authUsecase.Login(c.Request().Context(), strings.TrimSpace(req.Email), strings.TrimSpace(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	token, err := h.tokenService.Generate(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  toAuthUser(*user),
	})
}

func toAuthUser(user entity.User) dto.AuthUser {
	return dto.AuthUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
