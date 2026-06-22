package subscriptionhttp

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
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/subscription"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
)

type mockService struct {
	createFunc  func(context.Context, subscription.CreateInput) (*subscription.Subscription, error)
	updateFunc  func(context.Context, subscription.UpdateInput) (*subscription.Subscription, error)
	getByIDFunc func(context.Context, string) (*subscription.Subscription, error)
	listFunc    func(context.Context, int, int, string, string) ([]subscription.Subscription, int, error)
}

func (m *mockService) Create(ctx context.Context, input subscription.CreateInput) (*subscription.Subscription, error) {
	return m.createFunc(ctx, input)
}
func (m *mockService) Update(ctx context.Context, input subscription.UpdateInput) (*subscription.Subscription, error) {
	return m.updateFunc(ctx, input)
}
func (m *mockService) GetByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockService) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]subscription.Subscription, int, error) {
	return m.listFunc(ctx, limit, offset, orderBy, sortBy)
}

func testRouter(handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(httpmiddleware.RequestorIDContextKey, uuid.NewString()) })
	router.Any("/subscriptions", handler)
	router.Any("/subscriptions/:id", handler)
	return router
}

func assertError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()
	if recorder.Code != wantStatus || !strings.Contains(recorder.Body.String(), `"error":"`+wantError+`"`) {
		t.Fatalf("status = %d, body = %s; want %d and error %q", recorder.Code, recorder.Body.String(), wantStatus, wantError)
	}
}

func TestCreateSubscription_SuccessAndUserNotFound(t *testing.T) {
	userID := uuid.NewString()
	t.Run("success", func(t *testing.T) {
		handler := NewHandler(&mockService{createFunc: func(_ context.Context, input subscription.CreateInput) (*subscription.Subscription, error) {
			if input.UserID != userID || input.CreatedBy == "" {
				t.Fatalf("unexpected create input: %+v", input)
			}
			return &subscription.Subscription{ID: uuid.New(), UserID: uuid.MustParse(userID), PlanName: "pro", Status: "active"}, nil
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(`{"user_id":"`+userID+`","plan_name":"pro","status":"active","start_date":"2026-01-01","end_date":"2027-01-01","auto_renew":true}`))
		request.Header.Set("Content-Type", "application/json")
		testRouter(handler.CreateSubscription).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusCreated || !strings.Contains(recorder.Body.String(), `"plan_name":"pro"`) {
			t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("user not found", func(t *testing.T) {
		handler := NewHandler(&mockService{createFunc: func(context.Context, subscription.CreateInput) (*subscription.Subscription, error) {
			return nil, subscription.ErrUserNotFound
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(`{"user_id":"`+userID+`","plan_name":"pro","status":"active","start_date":"2026-01-01","end_date":"2027-01-01"}`))
		request.Header.Set("Content-Type", "application/json")
		testRouter(handler.CreateSubscription).ServeHTTP(recorder, request)
		assertError(t, recorder, http.StatusNotFound, subscription.ErrUserNotFound.Error())
	})
}

func TestUpdateAndGetSubscription(t *testing.T) {
	id := uuid.NewString()
	t.Run("update success", func(t *testing.T) {
		handler := NewHandler(&mockService{updateFunc: func(_ context.Context, input subscription.UpdateInput) (*subscription.Subscription, error) {
			return &subscription.Subscription{ID: uuid.MustParse(input.ID), Status: input.Status}, nil
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/subscriptions/"+id, strings.NewReader(`{"status":"inactive"}`))
		request.Header.Set("Content-Type", "application/json")
		testRouter(handler.UpdateSubscription).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"status":"inactive"`) {
			t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("update not found", func(t *testing.T) {
		handler := NewHandler(&mockService{updateFunc: func(context.Context, subscription.UpdateInput) (*subscription.Subscription, error) {
			return nil, subscription.ErrSubscriptionNotFound
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/subscriptions/"+id, strings.NewReader(`{"status":"inactive"}`))
		request.Header.Set("Content-Type", "application/json")
		testRouter(handler.UpdateSubscription).ServeHTTP(recorder, request)
		assertError(t, recorder, http.StatusNotFound, subscription.ErrSubscriptionNotFound.Error())
	})
	t.Run("get success", func(t *testing.T) {
		handler := NewHandler(&mockService{getByIDFunc: func(_ context.Context, gotID string) (*subscription.Subscription, error) {
			return &subscription.Subscription{ID: uuid.MustParse(gotID), PlanName: "pro"}, nil
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		testRouter(handler.GetSubscription).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/subscriptions/"+id, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"plan_name":"pro"`) {
			t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("get server error", func(t *testing.T) {
		handler := NewHandler(&mockService{getByIDFunc: func(context.Context, string) (*subscription.Subscription, error) {
			return nil, errors.New("database unavailable")
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		testRouter(handler.GetSubscription).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/subscriptions/"+id, nil))
		assertError(t, recorder, http.StatusInternalServerError, apperrors.ErrInternalServerError.Error())
	})
}

func TestGetSubscriptions_SuccessAndError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler := NewHandler(&mockService{listFunc: func(_ context.Context, limit, offset int, orderBy, sortBy string) ([]subscription.Subscription, int, error) {
			return []subscription.Subscription{{ID: uuid.New(), PlanName: "pro"}}, 1, nil
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		testRouter(handler.GetSubscriptions).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/subscriptions", nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total_results":1`) {
			t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("server error", func(t *testing.T) {
		handler := NewHandler(&mockService{listFunc: func(context.Context, int, int, string, string) ([]subscription.Subscription, int, error) {
			return nil, 0, errors.New("database unavailable")
		}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		recorder := httptest.NewRecorder()
		testRouter(handler.GetSubscriptions).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/subscriptions", nil))
		assertError(t, recorder, http.StatusInternalServerError, apperrors.ErrInternalServerError.Error())
	})
}
