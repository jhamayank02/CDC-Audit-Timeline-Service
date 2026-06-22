package auditloghttp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
)

type mockService struct {
	getAuditLogsFunc func(context.Context, int, int, string, string) ([]auditlog.AuditLog, int, error)
}

func (m *mockService) Record(context.Context, auditlog.AuditLog) error { return nil }

func (m *mockService) GetAuditLogs(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]auditlog.AuditLog, int, error) {
	return m.getAuditLogsFunc(ctx, limit, offset, orderBy, sortBy)
}

func TestGetAuditLogs_Success(t *testing.T) {
	handler := NewHandler(&mockService{
		getAuditLogsFunc: func(_ context.Context, limit, offset int, orderBy, sortBy string) ([]auditlog.AuditLog, int, error) {
			if limit != 10 || offset != 0 || orderBy != "created_at" || sortBy != "asc" {
				t.Fatalf("unexpected list arguments: limit=%d offset=%d orderBy=%q sortBy=%q", limit, offset, orderBy, sortBy)
			}
			return []auditlog.AuditLog{{Id: "event-1", TableName: "users", Operation: auditlog.OperationCreate}}, 1, nil
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	router := gin.New()
	router.GET("/audit-logs", handler.GetAuditLogs)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/audit-logs", nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"audit_logs"`) || !strings.Contains(recorder.Body.String(), `"total_results":1`) {
		t.Fatalf("status = %d, body = %s; want audit logs", recorder.Code, recorder.Body.String())
	}
}

func TestGetAuditLogs_ServiceError(t *testing.T) {
	handler := NewHandler(&mockService{
		getAuditLogsFunc: func(context.Context, int, int, string, string) ([]auditlog.AuditLog, int, error) {
			return nil, 0, errors.New("database unavailable")
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	router := gin.New()
	router.GET("/audit-logs", handler.GetAuditLogs)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/audit-logs", nil))

	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"error":"`+apperrors.ErrInternalServerError.Error()+`"`) {
		t.Fatalf("status = %d, body = %s; want internal server error", recorder.Code, recorder.Body.String())
	}
}
