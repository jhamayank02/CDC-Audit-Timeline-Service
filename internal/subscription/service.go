package subscription

import (
	"context"
	"log/slog"
)

type service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) (*Subscription, error) {
	exists, err := s.repo.UserExists(ctx, req.UserID)
	if err != nil {
		s.logger.Error("[SERVICE] failed to check user exists", "err", err)
		return nil, err
	}
	if !exists {
		s.logger.Error("[SERVICE] user not found", "user_id", req.UserID)
		return nil, ErrUserNotFound
	}

	subscription, err := s.repo.CreateSubscription(ctx, req)
	if err != nil {
		s.logger.Error("[SERVICE] failed to create subscription", "err", err)
		return nil, err
	}
	return subscription, err
}

func (s *service) UpdateSubscription(ctx context.Context, req *UpdateSubscriptionReq) (*Subscription, error) {
	subscription, err := s.repo.UpdateSubscription(ctx, req)
	if err != nil {
		s.logger.Error("[SERVICE] failed to update subscription", "err", err)
		return nil, err
	}
	return subscription, err
}

func (s *service) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	subscription, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get subscription", "err", err)
		return nil, err
	}
	return subscription, err
}

func (s *service) GetSubscriptions(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error) {
	subscriptions, totalCount, err := s.repo.GetSubscriptions(ctx, limit, offset, orderBy, sortBy)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get subscriptions", "err", err)
		return nil, 0, err
	}
	return subscriptions, totalCount, err
}
