package user

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

func (s *service) CreateUser(ctx context.Context, req *CreateUserReq) (*User, error) {
	user, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		s.logger.Error("[SERVICE] failed to create user", "err", err)
		return nil, err
	}
	return user, err
}

func (s *service) UpdateUser(ctx context.Context, req *UpdateUserReq) (*User, error) {
	user, err := s.repo.UpdateUser(ctx, req)
	if err != nil {
		s.logger.Error("[SERVICE] failed to update user", "err", err)
		return nil, err
	}
	return user, err
}

func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get user", "err", err)
		return nil, err
	}
	return user, err
}

func (s *service) GetUsers(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error) {
	users, totalCount, err := s.repo.GetUsers(ctx, limit, offset, orderBy, sortBy)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get users", "err", err)
		return nil, 0, err
	}
	return users, totalCount, err
}
