package auditlog

import (
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
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

	for i := range auditLogs {
		changes, err := calculateChangedFields(auditLogs[i])
		if err != nil {
			s.logger.Error("[SERVICE] failed to calculate changed fields", "event", auditLogs[i], "err", err)
		} else {
			auditLogs[i].Changes = changes
		}
	}

	return auditLogs, totalCount, nil
}

func calculateChangedFields(event AuditLog) (map[string]FieldChange, error) {
	changes := make(map[string]FieldChange)

	before, err := decodeJSONObject(event.Before)
	if err != nil {
		return nil, err
	}

	after, err := decodeJSONObject(event.After)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(before)+len(after))
	for key := range before {
		keys[key] = struct{}{}
	}
	for key := range after {
		keys[key] = struct{}{}
	}

	for key := range keys {
		oldValue, oldExists := before[key]
		newValue, newExists := after[key]
		if !oldExists {
			oldValue = nil
		}
		if !newExists {
			newValue = nil
		}

		if !reflect.DeepEqual(oldValue, newValue) {
			changes[key] = FieldChange{
				Old: oldValue,
				New: newValue,
			}
		}
	}

	return changes, nil
}

func decodeJSONObject(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, nil
	}

	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}

	return values, nil
}
