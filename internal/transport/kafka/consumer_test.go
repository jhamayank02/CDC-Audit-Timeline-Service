package kafka

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	kafkago "github.com/segmentio/kafka-go"
)

type mockAuditService struct {
	recordFunc func(context.Context, auditlog.AuditLog) error
}

func (m *mockAuditService) Record(ctx context.Context, event auditlog.AuditLog) error {
	return m.recordFunc(ctx, event)
}

func (m *mockAuditService) GetAuditLogs(context.Context, int, int, string, string) ([]auditlog.AuditLog, int, error) {
	return nil, 0, nil
}

func newTestConsumer(auditor auditlog.Service) *Consumer {
	return &Consumer{auditor: auditor, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestConsumerProcessMessage(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		wantCalls  int
		wantTable  string
		wantOp     auditlog.Operation
		wantBefore string
		wantAfter  string
	}{
		{
			name:      "create event records audit log",
			value:     `{"payload":{"before":null,"after":{"id":"1","email":"new@example.com"},"source":{"table":"users"},"op":"c"}}`,
			wantCalls: 1, wantTable: "users", wantOp: auditlog.OperationCreate, wantBefore: "null", wantAfter: `{"id":"1","email":"new@example.com"}`,
		},
		{
			name:      "update event records audit log",
			value:     `{"payload":{"before":{"status":"active"},"after":{"status":"inactive"},"source":{"table":"subscriptions"},"op":"u"}}`,
			wantCalls: 1, wantTable: "subscriptions", wantOp: auditlog.OperationUpdate, wantBefore: `{"status":"active"}`, wantAfter: `{"status":"inactive"}`,
		},
		{
			name:      "delete event records audit log",
			value:     `{"payload":{"before":{"id":"1"},"after":null,"source":{"table":"users"},"op":"d"}}`,
			wantCalls: 1, wantTable: "users", wantOp: auditlog.OperationDelete, wantBefore: `{"id":"1"}`, wantAfter: "null",
		},
		{
			name:  "read event is skipped",
			value: `{"payload":{"before":null,"after":{"id":"1"},"source":{"table":"users"},"op":"r"}}`,
		},
		{
			name:  "invalid JSON is skipped",
			value: `{not-json}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			consumer := newTestConsumer(&mockAuditService{recordFunc: func(_ context.Context, got auditlog.AuditLog) error {
				calls++
				if got.TableName != tt.wantTable || got.Operation != tt.wantOp || string(got.Before) != tt.wantBefore || string(got.After) != tt.wantAfter {
					t.Fatalf("audit log = %+v, want table=%q operation=%q before=%s after=%s", got, tt.wantTable, tt.wantOp, tt.wantBefore, tt.wantAfter)
				}
				return nil
			}})

			consumer.processMessage(context.Background(), kafkago.Message{Topic: "db.public.users", Value: []byte(tt.value)})
			if calls != tt.wantCalls {
				t.Fatalf("Record calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}

func TestConsumerProcessMessage_AuditServiceError(t *testing.T) {
	calls := 0
	consumer := newTestConsumer(&mockAuditService{recordFunc: func(context.Context, auditlog.AuditLog) error {
		calls++
		return errors.New("database unavailable")
	}})

	consumer.processMessage(context.Background(), kafkago.Message{Value: []byte(`{"payload":{"source":{"table":"users"},"op":"c"}}`)})
	if calls != 1 {
		t.Fatalf("Record calls = %d, want 1", calls)
	}
}
