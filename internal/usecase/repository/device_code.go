package repository

import (
	"context"
	"errors"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
)

type DeviceCodeRepository interface {
	Create(ctx context.Context, code entity.DeviceCode) error
	FindByUserCode(ctx context.Context, userCode string) (*entity.DeviceCode, error)
	UpdateStatus(ctx context.Context, deviceCode string, from, to entity.DeviceCodeStatus, userID *string) (bool, error)
}

var (
	ErrDeviceCodeConflict = errors.New("device code already exists")
	ErrUserCodeConflict   = errors.New("user code already exists")
)
