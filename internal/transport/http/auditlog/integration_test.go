//go:build integration
// +build integration

package auditloghttp_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/postgres"
	auditloghttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/auditlog"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
)

const requestorID = "11111111-1111-1111-1111-111111111111"

type integrationApp struct {
	router          *gin.Engine
	auditLogService auditlog.Service
}

func setupIntegrationApp(t *testing.T) *integrationApp {
	t.Helper()

	log := logger.New()
	cfg, err := config.LoadTest(log)
	if err != nil {
		t.Fatalf("load test configuration: %v", err)
	}

	db, err := postgres.NewDB(cfg.TestDB, log)
	if err != nil {
		t.Fatalf("initialize test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	resetDatabase(t, db)

	userService := user.NewService(postgres.NewUserRepository(db, log), log)
	auditLogService := auditlog.NewService(postgres.NewAuditLogRepository(db, log), log)
	handler := auditloghttp.NewHandler(auditLogService, log)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	api.Use(httpmiddleware.NewRequestorMiddleware(userService, log).Handler())
	api.GET("/audit-logs/", handler.GetAuditLogs)

	return &integrationApp{router: router, auditLogService: auditLogService}
}

func resetDatabase(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`TRUNCATE users, subscriptions, audit_logs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate test database: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, first_name, last_name, email) VALUES ($1, $2, $3, $4)`, requestorID, "Requestor", "User", "requestor@example.com"); err != nil {
		t.Fatalf("seed requestor: %v", err)
	}
}

func performRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Requestor-Id", requestorID)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func requireStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, want, recorder.Body.String())
	}
}

func TestAuditLogAPI_Integration(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		seed       bool
		assert     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "lists calculated changes", path: "/api/audit-logs/?limit=10&page=1&orderBy=created_at&sortBy=desc", wantStatus: http.StatusOK, seed: true,
			assert: func(t *testing.T, response *httptest.ResponseRecorder) {
				var result struct {
					AuditLogs    []auditlog.AuditLog `json:"audit_logs"`
					TotalResults int                 `json:"total_results"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
					t.Fatalf("decode audit log response: %v; body = %s", err, response.Body.String())
				}
				if len(result.AuditLogs) != 1 || result.TotalResults != 1 {
					t.Fatalf("unexpected audit log response: %+v", result)
				}
				if change, ok := result.AuditLogs[0].Changes["email"]; !ok || change.Old != "old@example.com" || change.New != "new@example.com" {
					t.Fatalf("unexpected email change: %+v", result.AuditLogs[0].Changes)
				}
			},
		},
		{name: "rejects invalid limit", path: "/api/audit-logs/?limit=invalid", wantStatus: http.StatusBadRequest},
		{name: "rejects invalid page", path: "/api/audit-logs/?page=invalid", wantStatus: http.StatusBadRequest},
		{name: "rejects invalid order", path: "/api/audit-logs/?orderBy=email", wantStatus: http.StatusBadRequest},
		{name: "rejects invalid sort", path: "/api/audit-logs/?sortBy=sideways", wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			if test.seed {
				if err := app.auditLogService.Record(context.Background(), auditlog.AuditLog{TableName: "users", Operation: auditlog.OperationUpdate, Before: json.RawMessage(`{"email":"old@example.com","name":"John"}`), After: json.RawMessage(`{"email":"new@example.com","name":"John"}`)}); err != nil {
					t.Fatalf("record audit log: %v", err)
				}
			}
			response := performRequest(app.router, test.path)
			requireStatus(t, response, test.wantStatus)
			if test.assert != nil {
				test.assert(t, response)
			}
		})
	}
}

func TestAuditLogAPI_IntegrationRejectsUnknownRequestor(t *testing.T) {
	tests := []struct {
		name, requestorID string
		wantStatus        int
	}{
		{name: "rejects unknown requestor", requestorID: uuid.NewString(), wantStatus: http.StatusUnauthorized},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			req := httptest.NewRequest(http.MethodGet, "/api/audit-logs/", nil)
			req.Header.Set("X-Requestor-Id", test.requestorID)
			recorder := httptest.NewRecorder()
			app.router.ServeHTTP(recorder, req)
			requireStatus(t, recorder, test.wantStatus)
		})
	}
}
