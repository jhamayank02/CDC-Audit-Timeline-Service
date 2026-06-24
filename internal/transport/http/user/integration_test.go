//go:build integration
// +build integration

package userhttp_test

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
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/postgres"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
	userhttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/user"
)

const requestorID = "11111111-1111-1111-1111-111111111111"

type integrationApp struct {
	router      *gin.Engine
	userService user.Service
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
	handler := userhttp.NewHandler(userService, log)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	api.Use(httpmiddleware.NewRequestorMiddleware(userService, log).Handler())
	api.POST("/users/", handler.CreateUser)
	api.PUT("/users/:id", handler.UpdateUser)
	api.GET("/users/:id", handler.GetUser)
	api.GET("/users/", handler.GetUsers)
	return &integrationApp{router: router, userService: userService}
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

func seedUser(t *testing.T, app *integrationApp) *user.User {
	t.Helper()
	result, err := app.userService.Create(context.Background(), user.CreateInput{
		FirstName: "John", LastName: "Doe", Email: fmt.Sprintf("john.%s@example.com", uuid.NewString()), CreatedBy: requestorID,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return result
}

func performRequest(t *testing.T, router http.Handler, method, path string, body any, includeRequestor bool) *httptest.ResponseRecorder {
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if includeRequestor {
		req.Header.Set("X-Requestor-Id", requestorID)
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

func TestCreateUserIntegration(t *testing.T) {
	tests := []struct {
		name       string
		body       userhttp.CreateUserRequest
		wantStatus int
		assert     func(*testing.T, *integrationApp, *httptest.ResponseRecorder)
	}{
		{
			name: "creates user", body: userhttp.CreateUserRequest{FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"}, wantStatus: http.StatusCreated,
			assert: func(t *testing.T, app *integrationApp, recorder *httptest.ResponseRecorder) {
				created := decode[user.User](t, recorder)
				if created.ID == uuid.Nil || created.CreatedBy != requestorID || created.Email != "jane@example.com" {
					t.Fatalf("unexpected created user: %+v", created)
				}
				persisted, err := app.userService.GetByID(context.Background(), created.ID.String())
				if err != nil || persisted.Email != created.Email {
					t.Fatalf("unexpected persisted user: %+v, err = %v", persisted, err)
				}
			},
		},
		{name: "rejects invalid payload", body: userhttp.CreateUserRequest{FirstName: "J", LastName: "Doe", Email: "not-an-email"}, wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			recorder := performRequest(t, app.router, http.MethodPost, "/api/users/", test.body, true)
			requireStatus(t, recorder, test.wantStatus)
			if test.assert != nil {
				test.assert(t, app, recorder)
			}
		})
	}
}

func TestUpdateUserIntegration(t *testing.T) {
	tests := []struct {
		name       string
		path       func(*testing.T, *integrationApp) string
		body       userhttp.UpdateUserRequest
		wantStatus int
		assert     func(*testing.T, *integrationApp, *httptest.ResponseRecorder)
	}{
		{
			name: "updates user", path: func(t *testing.T, app *integrationApp) string { return "/api/users/" + seedUser(t, app).ID.String() }, body: userhttp.UpdateUserRequest{FirstName: "Jane"}, wantStatus: http.StatusOK,
			assert: func(t *testing.T, app *integrationApp, recorder *httptest.ResponseRecorder) {
				updated := decode[user.User](t, recorder)
				if updated.FirstName != "Jane" || updated.UpdatedBy != requestorID {
					t.Fatalf("unexpected updated user: %+v", updated)
				}
			},
		},
		{name: "rejects empty update", path: func(t *testing.T, app *integrationApp) string { return "/api/users/" + seedUser(t, app).ID.String() }, body: userhttp.UpdateUserRequest{}, wantStatus: http.StatusBadRequest},
		{name: "rejects invalid id", path: func(*testing.T, *integrationApp) string { return "/api/users/not-a-uuid" }, body: userhttp.UpdateUserRequest{FirstName: "Jane"}, wantStatus: http.StatusBadRequest},
		{name: "returns not found", path: func(*testing.T, *integrationApp) string { return "/api/users/" + uuid.NewString() }, body: userhttp.UpdateUserRequest{FirstName: "Jane"}, wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			recorder := performRequest(t, app.router, http.MethodPut, test.path(t, app), test.body, true)
			requireStatus(t, recorder, test.wantStatus)
			if test.assert != nil {
				test.assert(t, app, recorder)
			}
		})
	}
}

func TestGetUserIntegration(t *testing.T) {
	tests := []struct {
		name       string
		path       func(*testing.T, *integrationApp) string
		wantStatus int
	}{
		{name: "gets user", path: func(t *testing.T, app *integrationApp) string { return "/api/users/" + seedUser(t, app).ID.String() }, wantStatus: http.StatusOK},
		{name: "rejects invalid id", path: func(*testing.T, *integrationApp) string { return "/api/users/not-a-uuid" }, wantStatus: http.StatusBadRequest},
		{name: "returns not found", path: func(*testing.T, *integrationApp) string { return "/api/users/" + uuid.NewString() }, wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			recorder := performRequest(t, app.router, http.MethodGet, test.path(t, app), nil, true)
			requireStatus(t, recorder, test.wantStatus)
			if test.wantStatus == http.StatusOK && decode[user.User](t, recorder).ID == uuid.Nil {
				t.Fatal("expected a user ID")
			}
		})
	}
}

func TestGetUsersIntegration(t *testing.T) {
	tests := []struct {
		name, path string
		wantStatus int
	}{
		{name: "lists users", path: "/api/users/?limit=10&page=1&orderBy=email&sortBy=asc", wantStatus: http.StatusOK},
		{name: "rejects invalid order", path: "/api/users/?orderBy=name", wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupIntegrationApp(t)
			seedUser(t, app)
			recorder := performRequest(t, app.router, http.MethodGet, test.path, nil, true)
			requireStatus(t, recorder, test.wantStatus)
			if test.wantStatus == http.StatusOK {
				response := decode[struct {
					Users        []user.User `json:"users"`
					TotalResults int         `json:"total_results"`
				}](t, recorder)
				if len(response.Users) != 2 || response.TotalResults != 2 {
					t.Fatalf("unexpected user list: %+v", response)
				}
			}
		})
	}
}
