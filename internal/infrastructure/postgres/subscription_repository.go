package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/subscription"
)

type SubscriptionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSubscriptionRepository(db *sql.DB, logger *slog.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, logger: logger}
}

var (
	createSubscriptionQuery = `
		INSERT INTO subscriptions (id, user_id, plan_name, status, start_date, end_date, auto_renew, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7::BOOLEAN, TRUE), NULLIF($8, '')::UUID, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	updateSubscriptionQuery = `
		UPDATE subscriptions SET status = COALESCE(NULLIF($2, ''), status), auto_renew = COALESCE($3::BOOLEAN, auto_renew), updated_by = COALESCE(NULLIF($4, '')::UUID, updated_by), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	getSubscriptionQuery = `
		SELECT id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at FROM subscriptions WHERE id = $1
	`
	listSubscriptionsQuery = `
		SELECT id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at, COUNT(*) OVER() AS total_count FROM subscriptions
	`
)

func (r *SubscriptionRepository) Create(ctx context.Context, input subscription.CreateInput) (*subscription.Subscription, error) {
	id := uuid.New()

	var result subscription.Subscription
	err := r.db.QueryRowContext(ctx, createSubscriptionQuery, id, input.UserID, input.PlanName, input.Status, input.StartDate, input.EndDate, input.AutoRenew, input.CreatedBy).
		Scan(&result.ID, &result.UserID, &result.PlanName, &result.Status, &result.StartDate, &result.EndDate, &result.AutoRenew, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan subscription", "err", err)
		return nil, err
	}

	return &result, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, input subscription.UpdateInput) (*subscription.Subscription, error) {
	var result subscription.Subscription
	err := r.db.QueryRowContext(ctx, updateSubscriptionQuery, input.ID, input.Status, input.AutoRenew, input.UpdatedBy).
		Scan(&result.ID, &result.UserID, &result.PlanName, &result.Status, &result.StartDate, &result.EndDate, &result.AutoRenew, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan subscription", "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, err
	}

	return &result, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	var result subscription.Subscription
	err := r.db.QueryRowContext(ctx, getSubscriptionQuery, id).
		Scan(&result.ID, &result.UserID, &result.PlanName, &result.Status, &result.StartDate, &result.EndDate, &result.AutoRenew, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan subscription", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, err
	}

	return &result, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]subscription.Subscription, int, error) {
	query := fmt.Sprintf("%s ORDER BY %s %s LIMIT %d OFFSET %d", listSubscriptionsQuery, orderBy, sortBy, limit, offset)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subscriptions := make([]subscription.Subscription, 0)
	var totalCount int
	for rows.Next() {
		var result subscription.Subscription
		err := rows.Scan(&result.ID, &result.UserID, &result.PlanName, &result.Status, &result.StartDate, &result.EndDate, &result.AutoRenew, &result.CreatedBy, &result.UpdatedBy, &result.CreatedAt, &result.UpdatedAt, &totalCount)
		if err != nil {
			r.logger.Error("[REPOSITORY] failed to scan subscription", "err", err)
			return nil, 0, err
		}
		subscriptions = append(subscriptions, result)
	}

	return subscriptions, totalCount, rows.Err()
}
