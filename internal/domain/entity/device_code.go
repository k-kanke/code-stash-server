package entity

import "time"

type DeviceCodeStatus string

const (
	DeviceCodeStatusPending  DeviceCodeStatus = "pending"
	DeviceCodeStatusApproved DeviceCodeStatus = "approved"
	DeviceCodeStatusDenied   DeviceCodeStatus = "denied"
	DeviceCodeStatusExpired  DeviceCodeStatus = "expired"
	DeviceCodeStatusConsumed DeviceCodeStatus = "consumed"
)

type DeviceCode struct {
	DeviceCode string
	UserCode   string
	ClientID   string
	UserID     *string
	Scope      []string
	Status     DeviceCodeStatus

	ExpiresAt   time.Time
	IntervalSec int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
