package subscription

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

var (
	create_subscription_query = `
		INSERT INTO subscriptions (id, user_id, plan_name, status, start_date, end_date, auto_renew, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7::BOOLEAN, TRUE), NULLIF($8, '')::UUID, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	update_subscription_query = `
		UPDATE subscriptions SET status = COALESCE(NULLIF($2, ''), status), auto_renew = COALESCE($3::BOOLEAN, auto_renew), updated_by = COALESCE(NULLIF($4, '')::UUID, updated_by), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at
	`
	get_subscription_query = `
		SELECT id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at FROM subscriptions WHERE id = $1
	`
	get_subscriptions_query = `
		SELECT id, user_id, plan_name, status, start_date, end_date, auto_renew, COALESCE(created_by::TEXT, '') AS created_by, COALESCE(updated_by::TEXT, '') AS updated_by, created_at, updated_at, COUNT(*) OVER() AS total_count FROM subscriptions
	`
)

func (r *repository) CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (*Subscription, error) {
	id := uuid.New()

	var subscription Subscription
	err := r.db.QueryRowContext(ctx, create_subscription_query, id, req.UserID, req.PlanName, req.Status, req.StartDate, req.EndDate, req.AutoRenew, req.CreatedBy).
		Scan(&subscription.ID, &subscription.UserID, &subscription.PlanName, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.AutoRenew, &subscription.CreatedBy, &subscription.UpdatedBy, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan subscription", "err", err)
		return nil, err
	}

	return &subscription, nil
}

func (r *repository) UpdateSubscription(ctx context.Context, req *UpdateSubscriptionReq) (*Subscription, error) {
	var subscription Subscription
	err := r.db.QueryRowContext(ctx, update_subscription_query, req.Id, req.Status, req.AutoRenew, req.UpdatedBy).
		Scan(&subscription.ID, &subscription.UserID, &subscription.PlanName, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.AutoRenew, &subscription.CreatedBy, &subscription.UpdatedBy, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan subscription", "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *repository) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	var subscription Subscription
	err := r.db.QueryRowContext(ctx, get_subscription_query, id).
		Scan(&subscription.ID, &subscription.UserID, &subscription.PlanName, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.AutoRenew, &subscription.CreatedBy, &subscription.UpdatedBy, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to scan subscription", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *repository) GetSubscriptions(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error) {
	query := fmt.Sprintf("%s ORDER BY %s %s LIMIT %d OFFSET %d", get_subscriptions_query, orderBy, sortBy, limit, offset)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subscriptions := make([]Subscription, 0)
	var totalCount int
	for rows.Next() {
		var subscription Subscription
		err := rows.Scan(&subscription.ID, &subscription.UserID, &subscription.PlanName, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.AutoRenew, &subscription.CreatedBy, &subscription.UpdatedBy, &subscription.CreatedAt, &subscription.UpdatedAt, &totalCount)
		if err != nil {
			r.logger.Error("[REPOSITORY] failed to scan subscription", "err", err)
			return nil, 0, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, totalCount, nil
}
