package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller/dto"
	"github.com/k-kanke/code-stash-server/internal/usecase/oauth"
)

type OAuthHandler struct {
	deviceCodeUsecase *oauth.DeviceCodeUsecase
}

func NewOAuthHandler(usecase *oauth.DeviceCodeUsecase) *OAuthHandler {
	return &OAuthHandler{deviceCodeUsecase: usecase}
}

func (h *OAuthHandler) CreateDeviceCode(c echo.Context) error {
	var req dto.DeviceCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	clientID := strings.TrimSpace(req.ClientID)
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_client"})
	}

	scopes := parseScope(req.Scope)

	result, err := h.deviceCodeUsecase.IssueDeviceCode(c.Request().Context(), oauth.DeviceCodeInput{
		ClientID: clientID,
		Scope:    scopes,
	})
	if err != nil {
		switch {
		case errors.Is(err, oauth.ErrInvalidClientID):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_client"})
		case errors.Is(err, oauth.ErrClientNotFound):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_client"})
		case errors.Is(err, oauth.ErrClientNotAllowed):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "unauthorized_client"})
		case errors.Is(err, oauth.ErrVerificationURIAbsent):
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server_error"})
		case errors.Is(err, oauth.ErrGenerationConflicted):
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server_error"})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server_error"})
		}
	}

	resp := dto.DeviceCodeResponse{
		DeviceCode:              result.DeviceCode,
		UserCode:                result.UserCode,
		VerificationURI:         result.VerificationURI,
		VerificationURIComplete: result.VerificationURIComplete,
		ExpiresIn:               int(result.ExpiresIn.Seconds()),
		Interval:                result.IntervalSec,
	}

	return c.JSON(http.StatusOK, resp)
}

func parseScope(scope string) []string {
	if strings.TrimSpace(scope) == "" {
		return nil
	}
	fields := strings.Fields(scope)
	if len(fields) == 0 {
		return nil
	}
	return fields
}
