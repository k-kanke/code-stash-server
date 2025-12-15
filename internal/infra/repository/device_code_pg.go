package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/k-kanke/code-stash-server/internal/domain/entity"
	usecaseRepo "github.com/k-kanke/code-stash-server/internal/usecase/repository"
)

type deviceCodePGRepository struct {
	db *sql.DB
}

func NewDeviceCodePGRepository(db *sql.DB) usecaseRepo.DeviceCodeRepository {
	return &deviceCodePGRepository{db: db}
}

func (r *deviceCodePGRepository) Create(ctx context.Context, code entity.DeviceCode) error {
	const query = `
INSERT INTO device_codes (
	device_code, user_code, client_id, user_id, scope, status, expires_at, interval_sec
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
)`

	var userID any
	if code.UserID != nil {
		userID = *code.UserID
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		code.DeviceCode,
		code.UserCode,
		code.ClientID,
		userID,
		pq.StringArray(code.Scope),
		string(code.Status),
		code.ExpiresAt,
		code.IntervalSec,
	)
	if err != nil {
		return mapDeviceCodeError(err)
	}
	return nil
}

func mapDeviceCodeError(err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Constraint {
		case "device_codes_pkey":
			return usecaseRepo.ErrDeviceCodeConflict
		case "device_codes_user_code_key":
			return usecaseRepo.ErrUserCodeConflict
		}
	}
	return err
}
