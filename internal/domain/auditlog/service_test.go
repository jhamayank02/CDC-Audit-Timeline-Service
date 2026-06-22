package auditlog

import (
	"context"
	"errors"
	"testing"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
)

type mockStore struct {
	create func(context.Context, AuditLog) error
	get    func(context.Context, int, int, string, string) ([]AuditLog, int, error)
}

func (m mockStore) Create(ctx context.Context, log AuditLog) error { return m.create(ctx, log) }
func (m mockStore) Get(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]AuditLog, int, error) {
	return m.get(ctx, limit, offset, orderBy, sortBy)
}

func TestServiceRecord(t *testing.T) {
	entry := AuditLog{TableName: "users", Operation: OperationUpdate}
	t.Run("success", func(t *testing.T) {
		called := false
		svc := NewService(mockStore{create: func(_ context.Context, got AuditLog) error { called = got.TableName == entry.TableName; return nil }}, logger.New())
		if err := svc.Record(context.Background(), entry); err != nil || !called {
			t.Fatalf("Record() error = %v, called = %t", err, called)
		}
	})
	t.Run("store error", func(t *testing.T) {
		wantErr := errors.New("database error")
		svc := NewService(mockStore{create: func(context.Context, AuditLog) error { return wantErr }}, logger.New())
		if err := svc.Record(context.Background(), entry); !errors.Is(err, wantErr) {
			t.Fatalf("Record() error = %v, want %v", err, wantErr)
		}
	})
}

func TestServiceGetAuditLogs(t *testing.T) {
	t.Run("success calculates changes", func(t *testing.T) {
		store := mockStore{get: func(_ context.Context, limit, offset int, orderBy, sortBy string) ([]AuditLog, int, error) {
			if limit != 10 || offset != 0 || orderBy != "created_at" || sortBy != "desc" {
				return nil, 0, errors.New("unexpected get arguments")
			}
			return []AuditLog{{TableName: "users", Before: []byte(`{"status":"pending"}`), After: []byte(`{"status":"active"}`)}}, 1, nil
		}}
		logs, total, err := NewService(store, logger.New()).GetAuditLogs(context.Background(), 10, 0, "created_at", "desc")
		if err != nil || total != 1 || len(logs) != 1 {
			t.Fatalf("GetAuditLogs() = %#v, %d, %v", logs, total, err)
		}
		change := logs[0].Changes["status"]
		if change.Old != "pending" || change.New != "active" {
			t.Errorf("status change = %#v", change)
		}
	})
	t.Run("store error", func(t *testing.T) {
		wantErr := errors.New("database error")
		svc := NewService(mockStore{get: func(context.Context, int, int, string, string) ([]AuditLog, int, error) { return nil, 0, wantErr }}, logger.New())
		if _, _, err := svc.GetAuditLogs(context.Background(), 10, 0, "created_at", "desc"); !errors.Is(err, wantErr) {
			t.Fatalf("GetAuditLogs() error = %v, want %v", err, wantErr)
		}
	})
}
