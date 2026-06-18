package subscription

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidUserID        = errors.New("user_id must be in valid uuid format")
	ErrInvalidOrderBy       = errors.New("orderBy must be one of id, user_id, plan_name, status, start_date, end_date, auto_renew, created_at, updated_at")
)

type Subscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PlanName  string    `json:"plan_name"`
	Status    string    `json:"status"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	AutoRenew bool      `json:"auto_renew"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type CreateInput struct {
	UserID    string
	PlanName  string
	Status    string
	StartDate string
	EndDate   string
	AutoRenew *bool
	CreatedBy string
}

type UpdateInput struct {
	ID        string
	Status    string
	AutoRenew *bool
	UpdatedBy string
}
