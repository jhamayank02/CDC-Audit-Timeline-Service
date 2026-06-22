package subscription

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
)

type mockStore struct {
	create func(context.Context, CreateInput) (*Subscription, error)
	update func(context.Context, UpdateInput) (*Subscription, error)
	get    func(context.Context, string) (*Subscription, error)
	list   func(context.Context, int, int, string, string) ([]Subscription, int, error)
}

func (m mockStore) Create(ctx context.Context, input CreateInput) (*Subscription, error) {
	return m.create(ctx, input)
}
func (m mockStore) Update(ctx context.Context, input UpdateInput) (*Subscription, error) {
	return m.update(ctx, input)
}
func (m mockStore) GetByID(ctx context.Context, id string) (*Subscription, error) {
	return m.get(ctx, id)
}
func (m mockStore) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error) {
	return m.list(ctx, limit, offset, orderBy, sortBy)
}

type mockUserService struct {
	get func(context.Context, string) (*user.User, error)
}

func (m mockUserService) Create(context.Context, user.CreateInput) (*user.User, error) {
	return nil, nil
}
func (m mockUserService) Update(context.Context, user.UpdateInput) (*user.User, error) {
	return nil, nil
}
func (m mockUserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	return m.get(ctx, id)
}
func (m mockUserService) List(context.Context, int, int, string, string) ([]user.User, int, error) {
	return nil, 0, nil
}

func TestServiceCreate(t *testing.T) {
	input := CreateInput{UserID: uuid.NewString(), PlanName: "Pro", Status: "active"}
	created := &Subscription{ID: uuid.New(), UserID: uuid.MustParse(input.UserID), PlanName: input.PlanName, Status: input.Status}

	t.Run("success", func(t *testing.T) {
		svc := NewService(mockStore{create: func(context.Context, CreateInput) (*Subscription, error) { return created, nil }}, mockUserService{get: func(context.Context, string) (*user.User, error) { return &user.User{}, nil }}, logger.New())
		got, err := svc.Create(context.Background(), input)
		if err != nil || got.ID != created.ID {
			t.Fatalf("Create() = %#v, %v; want %#v, nil", got, err, created)
		}
	})

	t.Run("maps missing user", func(t *testing.T) {
		svc := NewService(mockStore{}, mockUserService{get: func(context.Context, string) (*user.User, error) { return nil, user.ErrUserNotFound }}, logger.New())
		if _, err := svc.Create(context.Background(), input); !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("Create() error = %v, want %v", err, ErrUserNotFound)
		}
	})

	t.Run("store error", func(t *testing.T) {
		wantErr := errors.New("database error")
		svc := NewService(mockStore{create: func(context.Context, CreateInput) (*Subscription, error) { return nil, wantErr }}, mockUserService{get: func(context.Context, string) (*user.User, error) { return &user.User{}, nil }}, logger.New())
		if _, err := svc.Create(context.Background(), input); !errors.Is(err, wantErr) {
			t.Fatalf("Create() error = %v, want %v", err, wantErr)
		}
	})
}

func TestServiceUpdateGetAndList(t *testing.T) {
	updated := &Subscription{ID: uuid.New(), Status: "cancelled"}
	store := mockStore{
		update: func(context.Context, UpdateInput) (*Subscription, error) { return updated, nil },
		get:    func(context.Context, string) (*Subscription, error) { return updated, nil },
		list: func(_ context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error) {
			if limit != 10 || offset != 0 || orderBy != "created_at" || sortBy != "desc" {
				return nil, 0, errors.New("unexpected list arguments")
			}
			return []Subscription{*updated}, 1, nil
		},
	}
	svc := NewService(store, mockUserService{}, logger.New())
	if got, err := svc.Update(context.Background(), UpdateInput{ID: updated.ID.String()}); err != nil || got.ID != updated.ID {
		t.Fatalf("Update() = %#v, %v", got, err)
	}
	if got, err := svc.GetByID(context.Background(), updated.ID.String()); err != nil || got.ID != updated.ID {
		t.Fatalf("GetByID() = %#v, %v", got, err)
	}
	if got, total, err := svc.List(context.Background(), 10, 0, "created_at", "desc"); err != nil || len(got) != 1 || total != 1 {
		t.Fatalf("List() = %#v, %d, %v", got, total, err)
	}
}
