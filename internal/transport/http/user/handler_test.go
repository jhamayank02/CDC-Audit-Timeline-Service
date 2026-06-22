package userhttp

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
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
)

type mockService struct {
	createFunc  func(context.Context, user.CreateInput) (*user.User, error)
	updateFunc  func(context.Context, user.UpdateInput) (*user.User, error)
	getByIDFunc func(context.Context, string) (*user.User, error)
	listFunc    func(context.Context, int, int, string, string) ([]user.User, int, error)
}

func (m *mockService) Create(ctx context.Context, input user.CreateInput) (*user.User, error) {
	return m.createFunc(ctx, input)
}

func (m *mockService) Update(ctx context.Context, input user.UpdateInput) (*user.User, error) {
	return m.updateFunc(ctx, input)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*user.User, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockService) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]user.User, int, error) {
	return m.listFunc(ctx, limit, offset, orderBy, sortBy)
}

func newTestRouter(handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// CreateUser and UpdateUser read this value from the request context.
		c.Set(httpmiddleware.RequestorIDContextKey, uuid.NewString())
	})
	router.Any("/users", handler)
	router.Any("/users/:id", handler)
	return router
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()
	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, wantStatus, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"`+wantError+`"`) {
		t.Fatalf("body = %s, want error %q", recorder.Body.String(), wantError)
	}
}

func TestCreateUser_Success(t *testing.T) {
	expected := &user.User{ID: uuid.New(), FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"}
	handler := NewHandler(&mockService{
		createFunc: func(_ context.Context, input user.CreateInput) (*user.User, error) {
			if input.FirstName != "Jane" || input.LastName != "Doe" || input.Email != "jane@example.com" {
				t.Fatalf("unexpected create input: %+v", input)
			}
			if input.CreatedBy == "" {
				t.Fatal("CreatedBy must be populated from the request context")
			}
			return expected, nil
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"first_name":"Jane","last_name":"Doe","email":"jane@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	newTestRouter(handler.CreateUser).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"email":"jane@example.com"`) {
		t.Fatalf("body = %s, want created user", recorder.Body.String())
	}
}

func TestCreateUser_ServiceError(t *testing.T) {
	handler := NewHandler(&mockService{
		createFunc: func(context.Context, user.CreateInput) (*user.User, error) {
			return nil, errors.New("database unavailable")
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"first_name":"Jane","last_name":"Doe","email":"jane@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	newTestRouter(handler.CreateUser).ServeHTTP(recorder, request)

	assertErrorResponse(t, recorder, http.StatusInternalServerError, apperrors.ErrInternalServerError.Error())
}

func TestUpdateUser_Success(t *testing.T) {
	id := uuid.NewString()
	expected := &user.User{ID: uuid.MustParse(id), FirstName: "Jane", Email: "jane@example.com"}
	handler := NewHandler(&mockService{
		updateFunc: func(_ context.Context, input user.UpdateInput) (*user.User, error) {
			if input.ID != id || input.FirstName != "Jane" || input.UpdatedBy == "" {
				t.Fatalf("unexpected update input: %+v", input)
			}
			return expected, nil
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/users/"+id, strings.NewReader(`{"first_name":"Jane"}`))
	request.Header.Set("Content-Type", "application/json")
	newTestRouter(handler.UpdateUser).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"email":"jane@example.com"`) {
		t.Fatalf("status = %d, body = %s; want updated user", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateUser_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantError  string
	}{
		{"not found", user.ErrUserNotFound, http.StatusNotFound, user.ErrUserNotFound.Error()},
		{"unexpected error", errors.New("database unavailable"), http.StatusInternalServerError, apperrors.ErrInternalServerError.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&mockService{
				updateFunc: func(context.Context, user.UpdateInput) (*user.User, error) { return nil, tt.serviceErr },
			}, slog.New(slog.NewTextHandler(io.Discard, nil)))

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPatch, "/users/"+uuid.NewString(), strings.NewReader(`{"first_name":"Jane"}`))
			request.Header.Set("Content-Type", "application/json")
			newTestRouter(handler.UpdateUser).ServeHTTP(recorder, request)

			assertErrorResponse(t, recorder, tt.wantStatus, tt.wantError)
		})
	}
}

func TestGetUser_Success(t *testing.T) {
	id := uuid.NewString()
	handler := NewHandler(&mockService{
		getByIDFunc: func(_ context.Context, gotID string) (*user.User, error) {
			if gotID != id {
				t.Fatalf("id = %q, want %q", gotID, id)
			}
			return &user.User{ID: uuid.MustParse(id), Email: "jane@example.com"}, nil
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	newTestRouter(handler.GetUser).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/users/"+id, nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"email":"jane@example.com"`) {
		t.Fatalf("status = %d, body = %s; want user", recorder.Code, recorder.Body.String())
	}
}

func TestGetUser_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantError  string
	}{
		{"not found", user.ErrUserNotFound, http.StatusNotFound, user.ErrUserNotFound.Error()},
		{"unexpected error", errors.New("database unavailable"), http.StatusInternalServerError, apperrors.ErrInternalServerError.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&mockService{
				getByIDFunc: func(context.Context, string) (*user.User, error) { return nil, tt.serviceErr },
			}, slog.New(slog.NewTextHandler(io.Discard, nil)))

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/users/"+uuid.NewString(), nil)
			newTestRouter(handler.GetUser).ServeHTTP(recorder, request)

			assertErrorResponse(t, recorder, tt.wantStatus, tt.wantError)
		})
	}
}

func TestGetUsers_Success(t *testing.T) {
	handler := NewHandler(&mockService{
		listFunc: func(_ context.Context, limit, offset int, orderBy, sortBy string) ([]user.User, int, error) {
			if limit != 10 || offset != 0 || orderBy != "created_at" || sortBy != "asc" {
				t.Fatalf("unexpected list arguments: limit=%d offset=%d orderBy=%q sortBy=%q", limit, offset, orderBy, sortBy)
			}
			return []user.User{{ID: uuid.New(), Email: "jane@example.com"}}, 1, nil
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	newTestRouter(handler.GetUsers).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/users", nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total_results":1`) || !strings.Contains(recorder.Body.String(), `"email":"jane@example.com"`) {
		t.Fatalf("status = %d, body = %s; want users and total_results", recorder.Code, recorder.Body.String())
	}
}

func TestGetUsers_ServiceError(t *testing.T) {
	handler := NewHandler(&mockService{
		listFunc: func(context.Context, int, int, string, string) ([]user.User, int, error) {
			return nil, 0, errors.New("database unavailable")
		},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/users", nil)
	newTestRouter(handler.GetUsers).ServeHTTP(recorder, request)

	assertErrorResponse(t, recorder, http.StatusInternalServerError, apperrors.ErrInternalServerError.Error())
}
