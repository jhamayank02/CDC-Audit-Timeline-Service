package auditlog

import (
	"context"
	"log/slog"
)

type Service interface {
	Record(ctx context.Context, auditLog AuditLog) error
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

func (s *service) Record(ctx context.Context, auditLog AuditLog) error {
	if err := s.store.Create(ctx, auditLog); err != nil {
		s.logger.Error("[SERVICE] failed to record audit log", "err", err)
		return err
	}
	return nil
}
