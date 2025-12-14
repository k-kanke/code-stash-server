package infra

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/domain/service"
)

const userIDContextKey = "user_id"

func NewAuthMiddleware(tokenService *service.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
			}

			claims, err := tokenService.Verify(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}

			c.Set(userIDContextKey, claims.Subject)
			return next(c)
		}
	}
}

func UserIDFromContext(c echo.Context) (string, bool) {
	v := c.Get(userIDContextKey)
	if id, ok := v.(string); ok && id != "" {
		return id, true
	}
	return "", false
}
