package subscription

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (*Subscription, error)
	Update(ctx context.Context, input UpdateInput) (*Subscription, error)
	GetByID(ctx context.Context, id string) (*Subscription, error)
	List(ctx context.Context, limit int, offset int, orderBy, sortBy string) ([]Subscription, int, error)
}

type service struct {
	store   Store
	userSvc user.Service
	logger  *slog.Logger
}

func NewService(store Store, userSvc user.Service, logger *slog.Logger) Service {
	return &service{
		store:   store,
		userSvc: userSvc,
		logger:  logger,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Subscription, error) {
	_, err := s.userSvc.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			s.logger.Error("[SERVICE] user not found", "user_id", input.UserID)
			return nil, ErrUserNotFound
		}
		s.logger.Error("[SERVICE] failed to check user exists", "err", err)
		return nil, err
	}

	subscription, err := s.store.Create(ctx, input)
	if err != nil {
		s.logger.Error("[SERVICE] failed to create subscription", "err", err)
		return nil, err
	}
	return subscription, nil
}

func (s *service) Update(ctx context.Context, input UpdateInput) (*Subscription, error) {
	subscription, err := s.store.Update(ctx, input)
	if err != nil {
		s.logger.Error("[SERVICE] failed to update subscription", "err", err)
		return nil, err
	}
	return subscription, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Subscription, error) {
	subscription, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get subscription", "err", err)
		return nil, err
	}
	return subscription, nil
}

func (s *service) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error) {
	subscriptions, totalCount, err := s.store.List(ctx, limit, offset, orderBy, sortBy)
	if err != nil {
		s.logger.Error("[SERVICE] failed to list subscriptions", "err", err)
		return nil, 0, err
	}
	return subscriptions, totalCount, nil
}
