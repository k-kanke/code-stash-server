package oauth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

const (
	defaultCodeTTL           = 10 * time.Minute
	defaultPollInterval      = 5 * time.Second
	defaultDeviceCodeBytes   = 32
	defaultUserCodeDigits    = 6
	defaultGenerationRetries = 5
)

var (
	ErrInvalidClientID       = errors.New("client_id is required")
	ErrClientNotFound        = errors.New("client not found")
	ErrClientNotAllowed      = errors.New("client cannot use device authorization grant")
	ErrGenerationConflicted  = errors.New("device_code generation conflicted repeatedly")
	ErrVerificationURIAbsent = errors.New("verification uri is required")
)

type DeviceCodeInput struct {
	ClientID string
	Scope    []string
}

type DeviceCodeResult struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               time.Duration
	IntervalSec             int
}

type DeviceCodeConfig struct {
	VerificationURI     string
	CodeTTL             time.Duration
	PollInterval        time.Duration
	DeviceCodeBytes     int
	UserCodeDigits      int
	MaxGenerateAttempts int
}

type DeviceCodeUsecase struct {
	clientRepo usecaseRepo.OAuthClientRepository
	codeRepo   usecaseRepo.DeviceCodeRepository
	config     DeviceCodeConfig
	now        func() time.Time
}

func NewDeviceCodeUsecase(clientRepo usecaseRepo.OAuthClientRepository, codeRepo usecaseRepo.DeviceCodeRepository, cfg DeviceCodeConfig) (*DeviceCodeUsecase, error) {
	cfg = applyDeviceConfigDefaults(cfg)
	if strings.TrimSpace(cfg.VerificationURI) == "" {
		return nil, ErrVerificationURIAbsent
	}
	return &DeviceCodeUsecase{
		clientRepo: clientRepo,
		codeRepo:   codeRepo,
		config:     cfg,
		now:        time.Now,
	}, nil
}

func applyDeviceConfigDefaults(cfg DeviceCodeConfig) DeviceCodeConfig {
	if cfg.CodeTTL == 0 {
		cfg.CodeTTL = defaultCodeTTL
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}
	if cfg.DeviceCodeBytes == 0 {
		cfg.DeviceCodeBytes = defaultDeviceCodeBytes
	}
	if cfg.UserCodeDigits == 0 {
		cfg.UserCodeDigits = defaultUserCodeDigits
	}
	if cfg.MaxGenerateAttempts == 0 {
		cfg.MaxGenerateAttempts = defaultGenerationRetries
	}
	return cfg
}

func (uc *DeviceCodeUsecase) IssueDeviceCode(ctx context.Context, in DeviceCodeInput) (*DeviceCodeResult, error) {
	clientID := strings.TrimSpace(in.ClientID)
	if clientID == "" {
		return nil, ErrInvalidClientID
	}

	client, err := uc.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	if !supportsDeviceGrant(client.GrantTypes) {
		return nil, ErrClientNotAllowed
	}

	expiresAt := uc.now().Add(uc.config.CodeTTL)
	scope := normalizeScope(in.Scope)

	for attempt := 0; attempt < uc.config.MaxGenerateAttempts; attempt++ {
		deviceCode, userCode, genErr := uc.generateCodes()
		if genErr != nil {
			return nil, genErr
		}

		entityCode := entity.DeviceCode{
			DeviceCode:  deviceCode,
			UserCode:    userCode,
			ClientID:    client.ID,
			Scope:       scope,
			Status:      entity.DeviceCodeStatusPending,
			ExpiresAt:   expiresAt,
			IntervalSec: secondsOrOne(uc.config.PollInterval),
		}

		if err := uc.codeRepo.Create(ctx, entityCode); err != nil {
			if errors.Is(err, usecaseRepo.ErrDeviceCodeConflict) || errors.Is(err, usecaseRepo.ErrUserCodeConflict) {
				continue
			}
			return nil, err
		}

		return &DeviceCodeResult{
			DeviceCode:              deviceCode,
			UserCode:                userCode,
			VerificationURI:         uc.config.VerificationURI,
			VerificationURIComplete: buildVerificationURI(uc.config.VerificationURI, userCode),
			ExpiresIn:               uc.config.CodeTTL,
			IntervalSec:             secondsOrOne(uc.config.PollInterval),
		}, nil
	}

	return nil, ErrGenerationConflicted
}

func (uc *DeviceCodeUsecase) generateCodes() (string, string, error) {
	deviceCode, err := generateDeviceCode(uc.config.DeviceCodeBytes)
	if err != nil {
		return "", "", err
	}
	userCode, err := generateUserCode(uc.config.UserCodeDigits)
	if err != nil {
		return "", "", err
	}
	return deviceCode, userCode, nil
}

func generateDeviceCode(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func generateUserCode(digits int) (string, error) {
	if digits <= 0 {
		digits = defaultUserCodeDigits
	}
	max := big.NewInt(1)
	max.Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n.Int64()), nil
}

func supportsDeviceGrant(grants []string) bool {
	for _, grant := range grants {
		switch strings.ToLower(strings.TrimSpace(grant)) {
		case "device_code", "urn:ietf:params:oauth:grant-type:device_code":
			return true
		}
	}
	return false
}

func normalizeScope(scope []string) []string {
	if len(scope) == 0 {
		return nil
	}
	out := make([]string, 0, len(scope))
	for _, s := range scope {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildVerificationURI(base, code string) string {
	if strings.TrimSpace(base) == "" {
		return ""
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	q := parsed.Query()
	q.Set("user_code", code)
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func secondsOrOne(d time.Duration) int {
	sec := int(d / time.Second)
	if sec <= 0 {
		return 1
	}
	return sec
}
