package controller

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/k-kanke/code-stash-server/internal/adapter/controller/dto"
	"github.com/k-kanke/code-stash-server/internal/usecase/oauth"
)

type OAuthHandler struct {
	deviceCodeUsecase *oauth.DeviceCodeUsecase
	tokenUsecase      *oauth.TokenExchangeUsecase
}

type OAuthHandlerDependencies struct {
	DeviceCodeUsecase *oauth.DeviceCodeUsecase
	TokenUsecase      *oauth.TokenExchangeUsecase
}

func NewOAuthHandler(deps OAuthHandlerDependencies) *OAuthHandler {
	return &OAuthHandler{
		deviceCodeUsecase: deps.DeviceCodeUsecase,
		tokenUsecase:      deps.TokenUsecase,
	}
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

func (h *OAuthHandler) GetDeviceCodeStatus(c echo.Context) error {
	if _, err := requireUserID(c); err != nil {
		return err
	}

	userCode := strings.TrimSpace(c.QueryParam("user_code"))
	if userCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_code_required"})
	}

	result, err := h.deviceCodeUsecase.DescribeDeviceCode(c.Request().Context(), userCode)
	if err != nil {
		return h.handleDeviceCodeError(c, err)
	}

	return c.JSON(http.StatusOK, buildDeviceCodeStatusResponse(result))
}

func (h *OAuthHandler) VerifyDeviceCode(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req dto.VerifyDeviceCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	userCode := strings.TrimSpace(req.UserCode)
	if userCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_code_required"})
	}

	result, err := h.deviceCodeUsecase.ApproveDeviceCode(c.Request().Context(), userCode, userID)
	if err != nil {
		return h.handleDeviceCodeError(c, err)
	}

	return c.JSON(http.StatusOK, buildDeviceCodeStatusResponse(result))
}

func (h *OAuthHandler) handleDeviceCodeError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, oauth.ErrUserCodeRequired):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_code_required"})
	case errors.Is(err, oauth.ErrDeviceCodeNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "invalid_user_code"})
	case errors.Is(err, oauth.ErrDeviceCodeExpired):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "expired_token"})
	case errors.Is(err, oauth.ErrDeviceCodeNotPending):
		return c.JSON(http.StatusConflict, map[string]string{"error": "already_processed"})
	case errors.Is(err, oauth.ErrClientNotFound):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_client"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server_error"})
	}
}

func buildDeviceCodeStatusResponse(result *oauth.DeviceCodeStatusResult) dto.DeviceCodeStatusResponse {
	expiresAt := result.ExpiresAt.UTC().Format(time.RFC3339)
	return dto.DeviceCodeStatusResponse{
		DeviceCode: result.DeviceCode,
		UserCode:   result.UserCode,
		ClientID:   result.ClientID,
		ClientName: result.ClientName,
		Scope:      append([]string(nil), result.Scope...),
		Status:     string(result.Status),
		ExpiresAt:  expiresAt,
		Interval:   result.IntervalSec,
	}
}

func (h *OAuthHandler) ExchangeToken(c echo.Context) error {
	var req dto.TokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	result, err := h.tokenUsecase.ExchangeDeviceCode(c.Request().Context(), oauth.DeviceCodeGrantInput{
		GrantType:    req.GrantType,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		DeviceCode:   req.DeviceCode,
	})
	if err != nil {
		return h.handleTokenError(c, err)
	}

	scope := strings.Join(result.Scope, " ")
	resp := dto.TokenResponse{
		AccessToken:  result.AccessToken,
		TokenType:    result.TokenType,
		ExpiresIn:    int(result.ExpiresIn.Seconds()),
		RefreshToken: result.RefreshToken,
	}
	if scope != "" {
		resp.Scope = scope
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *OAuthHandler) handleTokenError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, oauth.ErrUnsupportedGrantType):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
	case errors.Is(err, oauth.ErrInvalidClientID), errors.Is(err, oauth.ErrInvalidClientSecret):
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
	case errors.Is(err, oauth.ErrInvalidDeviceCodeGrant), errors.Is(err, oauth.ErrDeviceCodeAlreadyUsed):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
	case errors.Is(err, oauth.ErrAuthorizationPending):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "authorization_pending"})
	case errors.Is(err, oauth.ErrAccessDenied):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "access_denied"})
	case errors.Is(err, oauth.ErrDeviceCodeExpired):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "expired_token"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server_error"})
	}
}
