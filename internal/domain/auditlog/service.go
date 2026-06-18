package auditlog

import (
	"context"
	"log/slog"
)

type Service interface {
	Record(ctx context.Context, auditLog AuditLog) error
	GetAuditLogs(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]AuditLog, int, error)
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

func (s *service) GetAuditLogs(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]AuditLog, int, error) {
	auditLogs, totalCount, err := s.store.Get(ctx, limit, offset, orderBy, sortBy)
	if err != nil {
		s.logger.Error("[SERVICE] failed to get audit logs", "err", err)
		return nil, 0, err
	}
	return auditLogs, totalCount, nil
}
