package user

import (
	"context"
	"log/slog"
)

type Service interface {
	Create(ctx context.Context, input CreateInput) (*User, error)
	Update(ctx context.Context, input UpdateInput) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error)
}

type service struct {
	store  Store
	logger *slog.Logger
}

func NewService(store Store, logger *slog.Logger) Service {
	return &service{
		store:  store,
		logger: logger,
	}
}

func (s *service) Create(ctx context.Context, input CreateInput) (*User, error) {
	user, err := s.store.Create(ctx, input)
	if err != nil {
		s.logger.Error("[SERVICE] failed to create user", "err", err)
		return nil, err
	}
	return user, nil
}

func (s *service) Update(ctx context.Context, input UpdateInput) (*User, error) {
	user, err := s.store.Update(ctx, input)
	if err != nil {
		s.logger.Error("[SERVICE] failed to update user", "err", err)
		return nil, err
	}
	return user, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	user, err := s.store.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get user", "err", err)
		return nil, err
	}
	return user, nil
}

func (s *service) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error) {
	users, totalCount, err := s.store.List(ctx, limit, offset, orderBy, sortBy)
	if err != nil {
		s.logger.Error("[SERVICE] failed to list users", "err", err)
		return nil, 0, err
	}
	return users, totalCount, nil
}
