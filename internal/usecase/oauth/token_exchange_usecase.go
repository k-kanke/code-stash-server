package oauth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

const (
	deviceCodeGrantURN = "urn:ietf:params:oauth:grant-type:device_code"
	deviceCodeGrant    = "device_code"
)

var (
	ErrUnsupportedGrantType   = errors.New("unsupported_grant_type")
	ErrInvalidClientSecret    = errors.New("invalid client credentials")
	ErrInvalidDeviceCodeGrant = errors.New("invalid device code grant")
	ErrAuthorizationPending   = errors.New("authorization_pending")
	ErrAccessDenied           = errors.New("access_denied")
	ErrDeviceCodeAlreadyUsed  = errors.New("device code already used")
)

type TokenService interface {
	Generate(userID string) (string, error)
}

type DeviceCodeGrantInput struct {
	GrantType    string
	ClientID     string
	ClientSecret string
	DeviceCode   string
}

type TokenResult struct {
	AccessToken  string
	RefreshToken *string
	ExpiresIn    time.Duration
	Scope        []string
	TokenType    string
}

type TokenExchangeConfig struct {
	AccessTokenTTL    time.Duration
	RefreshTokenBytes int
}

type TokenExchangeUsecase struct {
	clientRepo usecaseRepo.OAuthClientRepository
	deviceRepo usecaseRepo.DeviceCodeRepository
	tokenRepo  usecaseRepo.OAuthTokenRepository
	tokenSvc   TokenService
	config     TokenExchangeConfig
	now        func() time.Time
}

func NewTokenExchangeUsecase(
	clientRepo usecaseRepo.OAuthClientRepository,
	deviceRepo usecaseRepo.DeviceCodeRepository,
	tokenRepo usecaseRepo.OAuthTokenRepository,
	tokenSvc TokenService,
	cfg TokenExchangeConfig,
) *TokenExchangeUsecase {
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = 24 * time.Hour
	}
	if cfg.RefreshTokenBytes <= 0 {
		cfg.RefreshTokenBytes = 32
	}
	return &TokenExchangeUsecase{
		clientRepo: clientRepo,
		deviceRepo: deviceRepo,
		tokenRepo:  tokenRepo,
		tokenSvc:   tokenSvc,
		config:     cfg,
		now:        time.Now,
	}
}

func (uc *TokenExchangeUsecase) ExchangeDeviceCode(ctx context.Context, in DeviceCodeGrantInput) (*TokenResult, error) {
	grantType := strings.TrimSpace(in.GrantType)
	if grantType != deviceCodeGrant && grantType != deviceCodeGrantURN {
		return nil, ErrUnsupportedGrantType
	}

	clientID := strings.TrimSpace(in.ClientID)
	if clientID == "" {
		return nil, ErrInvalidClientID
	}

	deviceCode := strings.TrimSpace(in.DeviceCode)
	if deviceCode == "" {
		return nil, ErrInvalidDeviceCodeGrant
	}

	client, err := uc.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidClientID
		}
		return nil, err
	}

	if strings.TrimSpace(client.Secret) != "" {
		if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(in.ClientSecret)), []byte(client.Secret)) != 1 {
			return nil, ErrInvalidClientSecret
		}
	}

	record, err := uc.deviceRepo.FindByDeviceCode(ctx, deviceCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidDeviceCodeGrant
		}
		return nil, err
	}

	if record.ClientID != client.ID {
		return nil, ErrInvalidDeviceCodeGrant
	}

	if uc.isExpired(record) {
		return nil, ErrDeviceCodeExpired
	}

	switch record.Status {
	case entity.DeviceCodeStatusPending:
		return nil, ErrAuthorizationPending
	case entity.DeviceCodeStatusDenied:
		return nil, ErrAccessDenied
	case entity.DeviceCodeStatusConsumed:
		return nil, ErrDeviceCodeAlreadyUsed
	case entity.DeviceCodeStatusApproved:
	default:
		return nil, ErrInvalidDeviceCodeGrant
	}

	if record.UserID == nil || strings.TrimSpace(*record.UserID) == "" {
		return nil, ErrInvalidDeviceCodeGrant
	}

	accessToken, err := uc.tokenSvc.Generate(*record.UserID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	token := entity.OAuthToken{
		ID:           uuid.NewString(),
		UserID:       *record.UserID,
		ClientID:     record.ClientID,
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		Scope:        append([]string(nil), record.Scope...),
		ExpiresAt:    uc.now().Add(uc.config.AccessTokenTTL),
		CreatedAt:    uc.now(),
		UpdatedAt:    uc.now(),
	}

	if err := uc.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	if _, err := uc.deviceRepo.UpdateStatus(ctx, record.DeviceCode, entity.DeviceCodeStatusApproved, entity.DeviceCodeStatusConsumed, record.UserID); err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		ExpiresIn:    uc.config.AccessTokenTTL,
		Scope:        token.Scope,
		TokenType:    "Bearer",
	}, nil
}

func (uc *TokenExchangeUsecase) generateRefreshToken() (string, error) {
	buf := make([]byte, uc.config.RefreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (uc *TokenExchangeUsecase) isExpired(code *entity.DeviceCode) bool {
	if code == nil {
		return true
	}
	return !code.ExpiresAt.IsZero() && !uc.now().Before(code.ExpiresAt)
}
