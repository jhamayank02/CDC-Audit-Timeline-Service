//go:build integration
// +build integration

package subscriptionhttp_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/subscription"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/postgres"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
	subscriptionhttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/subscription"
)

const requestorID = "11111111-1111-1111-1111-111111111111"

type integrationApp struct {
	router              *gin.Engine
	userService         user.Service
	subscriptionService subscription.Service
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
	subscriptionService := subscription.NewService(postgres.NewSubscriptionRepository(db, log), userService, log)
	handler := subscriptionhttp.NewHandler(subscriptionService, log)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	api.Use(httpmiddleware.NewRequestorMiddleware(userService, log).Handler())
	api.POST("/subscriptions/", handler.CreateSubscription)
	api.PUT("/subscriptions/:id", handler.UpdateSubscription)
	api.GET("/subscriptions/:id", handler.GetSubscription)
	api.GET("/subscriptions/", handler.GetSubscriptions)

	return &integrationApp{router: router, userService: userService, subscriptionService: subscriptionService}
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

func createAccount(t *testing.T, app *integrationApp) *user.User {
	t.Helper()
	account, err := app.userService.Create(context.Background(), user.CreateInput{
		FirstName: "Subscription",
		LastName:  "Owner",
		Email:     fmt.Sprintf("subscription-owner.%s@example.com", uuid.NewString()),
		CreatedBy: requestorID,
	})
	if err != nil {
		t.Fatalf("create subscription owner: %v", err)
	}
	return account
}

func performRequest(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("X-Requestor-Id", requestorID)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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

func decode[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var response T
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, recorder.Body.String())
	}
	return response
}

func seedSubscription(t *testing.T, app *integrationApp) *subscription.Subscription {
	t.Helper()
	account := createAccount(t, app)
	autoRenew := true
	result, err := app.subscriptionService.Create(context.Background(), subscription.CreateInput{
		UserID: account.ID.String(), PlanName: "pro", Status: "active", StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-01T00:00:00Z", AutoRenew: &autoRenew, CreatedBy: requestorID,
	})
	if err != nil {
		t.Fatalf("seed subscription: %v", err)
	}
	return result
}

func TestCreateSubscriptionIntegration(t *testing.T) {
	tests := []struct {
		name       string
		request    func(*testing.T, *integrationApp) subscriptionhttp.CreateSubscriptionRequest
		wantStatus int
		assert     func(*testing.T, *integrationApp, *httptest.ResponseRecorder)
	}{
		{
			name: "creates subscription", wantStatus: http.StatusCreated,
			request: func(t *testing.T, app *integrationApp) subscriptionhttp.CreateSubscriptionRequest {
				autoRenew := true
				return subscriptionhttp.CreateSubscriptionRequest{UserID: createAccount(t, app).ID.String(), PlanName: "pro", Status: "active", StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-01T00:00:00Z", AutoRenew: &autoRenew}
			},
			assert: func(t *testing.T, app *integrationApp, recorder *httptest.ResponseRecorder) {
				created := decode[subscription.Subscription](t, recorder)
				if created.ID == uuid.Nil || !created.AutoRenew || created.CreatedBy != requestorID {
					t.Fatalf("unexpected created subscription: %+v", created)
				}
				if _, err := app.subscriptionService.GetByID(context.Background(), created.ID.String()); err != nil {
					t.Fatalf("get persisted subscription: %v", err)
				}
			},
		},
		{name: "rejects invalid payload", wantStatus: http.StatusBadRequest, request: func(*testing.T, *integrationApp) subscriptionhttp.CreateSubscriptionRequest {
			return subscriptionhttp.CreateSubscriptionRequest{UserID: "invalid", PlanName: "gold", Status: "unknown"}
		}},
		{name: "rejects unknown user", wantStatus: http.StatusNotFound, request: func(*testing.T, *integrationApp) subscriptionhttp.CreateSubscriptionRequest {
			return subscriptionhttp.CreateSubscriptionRequest{UserID: uuid.NewString(), PlanName: "basic", Status: "active", StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-01T00:00:00Z"}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			recorder := performRequest(t, app.router, http.MethodPost, "/api/subscriptions/", test.request(t, app))
			requireStatus(t, recorder, test.wantStatus)
			if test.assert != nil {
				test.assert(t, app, recorder)
			}
		})
	}
}

func TestUpdateSubscriptionIntegration(t *testing.T) {
	tests := []struct {
		name       string
		path       func(*testing.T, *integrationApp) string
		body       subscriptionhttp.UpdateSubscriptionRequest
		wantStatus int
		assert     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "updates subscription", path: func(t *testing.T, app *integrationApp) string {
				return "/api/subscriptions/" + seedSubscription(t, app).ID.String()
			}, wantStatus: http.StatusOK,
			body: func() subscriptionhttp.UpdateSubscriptionRequest {
				value := false
				return subscriptionhttp.UpdateSubscriptionRequest{Status: "cancelled", AutoRenew: &value}
			}(),
			assert: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				updated := decode[subscription.Subscription](t, recorder)
				if updated.Status != "cancelled" || updated.AutoRenew || updated.UpdatedBy != requestorID {
					t.Fatalf("unexpected updated subscription: %+v", updated)
				}
			},
		},
		{name: "rejects invalid id", path: func(*testing.T, *integrationApp) string { return "/api/subscriptions/not-a-uuid" }, body: subscriptionhttp.UpdateSubscriptionRequest{Status: "inactive"}, wantStatus: http.StatusBadRequest},
		{name: "returns not found", path: func(*testing.T, *integrationApp) string { return "/api/subscriptions/" + uuid.NewString() }, body: subscriptionhttp.UpdateSubscriptionRequest{Status: "inactive"}, wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			recorder := performRequest(t, app.router, http.MethodPut, test.path(t, app), test.body)
			requireStatus(t, recorder, test.wantStatus)
			if test.assert != nil {
				test.assert(t, recorder)
			}
		})
	}
}

func TestGetSubscriptionIntegration(t *testing.T) {
	tests := []struct {
		name       string
		path       func(*testing.T, *integrationApp) string
		wantStatus int
	}{
		{name: "gets subscription", path: func(t *testing.T, app *integrationApp) string {
			return "/api/subscriptions/" + seedSubscription(t, app).ID.String()
		}, wantStatus: http.StatusOK},
		{name: "rejects invalid id", path: func(*testing.T, *integrationApp) string { return "/api/subscriptions/not-a-uuid" }, wantStatus: http.StatusBadRequest},
		{name: "returns not found", path: func(*testing.T, *integrationApp) string { return "/api/subscriptions/" + uuid.NewString() }, wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			recorder := performRequest(t, app.router, http.MethodGet, test.path(t, app), nil)
			requireStatus(t, recorder, test.wantStatus)
			if test.wantStatus == http.StatusOK && decode[subscription.Subscription](t, recorder).ID == uuid.Nil {
				t.Fatal("expected a subscription ID")
			}
		})
	}
}

func TestGetSubscriptionsIntegration(t *testing.T) {
	tests := []struct {
		name, path string
		wantStatus int
	}{
		{name: "lists subscriptions", path: "/api/subscriptions/?limit=10&page=1&orderBy=status&sortBy=asc", wantStatus: http.StatusOK},
		{name: "rejects invalid query", path: "/api/subscriptions/?sortBy=sideways", wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			seedSubscription(t, app)
			recorder := performRequest(t, app.router, http.MethodGet, test.path, nil)
			requireStatus(t, recorder, test.wantStatus)
			if test.wantStatus == http.StatusOK {
				response := decode[struct {
					Subscriptions []subscription.Subscription `json:"subscriptions"`
					TotalResults  int                         `json:"total_results"`
				}](t, recorder)
				if len(response.Subscriptions) != 1 || response.TotalResults != 1 {
					t.Fatalf("unexpected subscription list: %+v", response)
				}
			}
		})
	}
}
