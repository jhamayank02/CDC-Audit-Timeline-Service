package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
)

func TestAuditLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuditLogRepository(db, logger.New())

	t.Run("success", func(t *testing.T) {
		tableName := "test_audit_" + uuid.NewString()
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM audit_logs WHERE table_name = $1", tableName)
		})

		err := repo.Create(context.Background(), auditlog.AuditLog{
			TableName: tableName,
			Operation: auditlog.OperationUpdate,
			Before:    []byte(`{"status":"pending"}`),
			After:     []byte(`{"status":"active"}`),
		})
		if err != nil {
			t.Fatalf("create audit log: %v", err)
		}

		var operation, before, after string
		if err := db.QueryRow(
			"SELECT operation, before::text, after::text FROM audit_logs WHERE table_name = $1",
			tableName,
		).Scan(&operation, &before, &after); err != nil {
			t.Fatalf("read created audit log: %v", err)
		}
		if operation != string(auditlog.OperationUpdate) || before != `{"status": "pending"}` || after != `{"status": "active"}` {
			t.Errorf("stored audit log = operation %q, before %s, after %s", operation, before, after)
		}
	})

	t.Run("empty JSON is stored as null", func(t *testing.T) {
		tableName := "test_audit_" + uuid.NewString()
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM audit_logs WHERE table_name = $1", tableName)
		})

		if err := repo.Create(context.Background(), auditlog.AuditLog{
			TableName: tableName,
			Operation: auditlog.OperationCreate,
		}); err != nil {
			t.Fatalf("create audit log: %v", err)
		}

		var before, after string
		if err := db.QueryRow(
			"SELECT before::text, after::text FROM audit_logs WHERE table_name = $1",
			tableName,
		).Scan(&before, &after); err != nil {
			t.Fatalf("read created audit log: %v", err)
		}
		if before != "null" || after != "null" {
			t.Errorf("empty JSON stored as before %s, after %s; want null values", before, after)
		}
	})
}

func TestAuditLogRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuditLogRepository(db, logger.New())
	tableName := "test_audit_" + uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM audit_logs WHERE table_name = $1", tableName)
	})

	if err := repo.Create(context.Background(), auditlog.AuditLog{
		TableName: tableName,
		Operation: auditlog.OperationCreate,
		After:     []byte(`{"id":"test"}`),
	}); err != nil {
		t.Fatalf("create audit log: %v", err)
	}

	logs, total, err := repo.Get(context.Background(), 100, 0, "created_at", "desc")
	if err != nil {
		t.Fatalf("get audit logs: %v", err)
	}
	if total < len(logs) {
		t.Errorf("total = %d, fewer than returned logs %d", total, len(logs))
	}
	for _, log := range logs {
		if log.TableName == tableName {
			if log.Operation != auditlog.OperationCreate || string(log.After) != `{"id": "test"}` {
				t.Errorf("audit log = %+v, want created entry for %q", log, tableName)
			}
			return
		}
	}
	t.Fatalf("created audit log for %q not returned", tableName)
}
