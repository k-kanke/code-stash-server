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

func (r *deviceCodePGRepository) FindByUserCode(ctx context.Context, userCode string) (*entity.DeviceCode, error) {
	return r.findOne(ctx, "user_code = $1", userCode)
}

func (r *deviceCodePGRepository) UpdateStatus(ctx context.Context, deviceCode string, from, to entity.DeviceCodeStatus, userID *string) (bool, error) {
	const query = `
UPDATE device_codes
SET status = $3, user_id = $4, updated_at = now()
WHERE device_code = $1 AND status = $2`

	var userValue any
	if userID != nil {
		userValue = *userID
	}

	result, err := r.db.ExecContext(ctx, query, deviceCode, string(from), string(to), userValue)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

func (r *deviceCodePGRepository) FindByDeviceCode(ctx context.Context, deviceCode string) (*entity.DeviceCode, error) {
	return r.findOne(ctx, "device_code = $1", deviceCode)
}

func (r *deviceCodePGRepository) findOne(ctx context.Context, where string, arg any) (*entity.DeviceCode, error) {
	query := `
SELECT device_code, user_code, client_id, user_id, scope, status, expires_at, interval_sec, created_at, updated_at
FROM device_codes
WHERE ` + where

	var code entity.DeviceCode
	var scope pq.StringArray
	var userID sql.NullString

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&code.DeviceCode,
		&code.UserCode,
		&code.ClientID,
		&userID,
		&scope,
		&code.Status,
		&code.ExpiresAt,
		&code.IntervalSec,
		&code.CreatedAt,
		&code.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		code.UserID = &userID.String
	}
	if scope != nil {
		code.Scope = append([]string(nil), scope...)
	}

	return &code, nil
}
