package subscription

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInternalServerError  = errors.New("internal server error")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidId            = errors.New("id must be in valid uuid format")
	ErrInvalidUserId        = errors.New("user_id must be in valid uuid format")
	ErrInvalidSortBy        = errors.New("sortBy must be asc or desc")
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

type CreateSubscriptionReq struct {
	UserID    string `json:"user_id" binding:"required,uuid"`
	PlanName  string `json:"plan_name" binding:"required,oneof=basic plus pro"`
	Status    string `json:"status" binding:"required,oneof=active inactive cancelled"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	AutoRenew *bool  `json:"auto_renew"`
	CreatedBy string `json:"created_by"`
}

type UpdateSubscriptionReq struct {
	Id        string `json:"id"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive cancelled"`
	AutoRenew *bool  `json:"auto_renew"`
	UpdatedBy string `json:"updated_by"`
}

type Repository interface {
	CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (*Subscription, error)
	UpdateSubscription(ctx context.Context, req *UpdateSubscriptionReq) (*Subscription, error)
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetSubscriptions(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error)
}

type Service interface {
	CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (*Subscription, error)
	UpdateSubscription(ctx context.Context, req *UpdateSubscriptionReq) (*Subscription, error)
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetSubscriptions(ctx context.Context, limit int, offset int, orderBy, sortBy string) ([]Subscription, int, error)
}
