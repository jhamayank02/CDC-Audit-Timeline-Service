package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
)

type mockStore struct {
	createFunc  func(ctx context.Context, input CreateInput) (*User, error)
	updateFunc  func(ctx context.Context, input UpdateInput) (*User, error)
	getByIdFunc func(ctx context.Context, id string) (*User, error)
	listFunc    func(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error)
}

func (m *mockStore) Create(ctx context.Context, input CreateInput) (*User, error) {
	return m.createFunc(ctx, input)
}

func (m *mockStore) Update(ctx context.Context, input UpdateInput) (*User, error) {
	return m.updateFunc(ctx, input)
}

func (m *mockStore) GetByID(ctx context.Context, id string) (*User, error) {
	return m.getByIdFunc(ctx, id)
}

func (m *mockStore) List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error) {
	return m.listFunc(ctx, limit, offset, orderBy, sortBy)
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateInput
		mockFunc  func(ctx context.Context, input CreateInput) (*User, error)
		wantErr   bool
		wantEmail string
	}{
		{
			name: "success",
			input: CreateInput{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@doe.com",
			},
			mockFunc: func(ctx context.Context, input CreateInput) (*User, error) {
				return &User{
					ID:        uuid.New(),
					FirstName: input.FirstName,
					LastName:  input.LastName,
					Email:     input.Email,
				}, nil
			},
			wantErr:   false,
			wantEmail: "john@doe.com",
		},
		{
			name: "duplicate email",
			input: CreateInput{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@doe.com",
			},
			mockFunc: func(ctx context.Context, input CreateInput) (*User, error) {
				return nil, errors.New("unique constraint failed on email")
			},
			wantErr: true,
		},
	}
	loggerMock := logger.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &mockStore{
				createFunc: tt.mockFunc,
			}
			service := NewService(mockStore, loggerMock)

			user, err := service.Create(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.Email != tt.wantEmail {
				t.Errorf(
					"expected email %s, got %s",
					tt.wantEmail,
					user.Email,
				)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		input     UpdateInput
		mockFunc  func(ctx context.Context, input UpdateInput) (*User, error)
		wantErr   bool
		wantEmail string
	}{
		{
			name: "success",
			input: UpdateInput{
				ID:        uuid.New().String(),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@doe.com",
				UpdatedBy: uuid.NewString(),
			},
			mockFunc: func(ctx context.Context, input UpdateInput) (*User, error) {
				return &User{
					ID:        uuid.MustParse(input.ID),
					FirstName: input.FirstName,
					LastName:  input.LastName,
					Email:     input.Email,
					UpdatedAt: input.UpdatedBy,
				}, nil
			},
			wantErr:   false,
			wantEmail: "john@doe.com",
		},
		{
			name: "user not found",
			input: UpdateInput{
				ID:        uuid.New().String(),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@doe.com",
				UpdatedBy: uuid.NewString(),
			},
			mockFunc: func(ctx context.Context, input UpdateInput) (*User, error) {
				return nil, errors.New("user not found")
			},
			wantErr: true,
		},
	}
	loggerMock := logger.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &mockStore{
				updateFunc: tt.mockFunc,
			}
			service := NewService(mockStore, loggerMock)

			user, err := service.Update(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.ID.String() != tt.input.ID {
				t.Errorf(
					"expected ID %s, got %s",
					tt.input.ID,
					user.ID.String(),
				)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		input     User
		mockFunc  func(ctx context.Context, id string) (*User, error)
		wantErr   bool
		wantEmail string
	}{
		{
			name: "success",
			input: User{
				ID:        uuid.New(),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@doe.com",
				CreatedBy: uuid.NewString(),
			},
			mockFunc: func(ctx context.Context, id string) (*User, error) {
				return &User{
					ID:        uuid.MustParse(id),
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@doe.com",
					CreatedBy: uuid.NewString(),
				}, nil
			},
			wantEmail: "john@doe.com",
		},
		{
			name: "user not found",
			input: User{
				ID: uuid.New(),
			},
			mockFunc: func(ctx context.Context, id string) (*User, error) {
				return nil, errors.New("user not found")
			},
			wantErr: true,
		},
	}
	loggerMock := logger.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &mockStore{
				getByIdFunc: tt.mockFunc,
			}
			service := NewService(mockStore, loggerMock)

			user, err := service.GetByID(context.Background(), tt.input.ID.String())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user.ID.String() != tt.input.ID.String() {
				t.Errorf(
					"expected ID %s, got %s",
					tt.input.ID,
					user.ID.String(),
				)
			}
			if user.Email != tt.wantEmail {
				t.Errorf("expected email %s, got %s", tt.wantEmail, user.Email)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		orderBy   string
		sortBy    string
		mockFunc  func(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error)
		want      []User
		wantTotal int
		wantErr   bool
	}{
		{
			name:    "success",
			limit:   10,
			offset:  0,
			orderBy: "created_at",
			sortBy:  "desc",
			want: []User{
				{
					ID:        uuid.New(),
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@doe.com",
					CreatedBy: uuid.NewString(),
				},
				{
					ID:        uuid.New(),
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@doe.com",
					CreatedBy: uuid.NewString(),
				},
			},
			wantTotal: 2,
			mockFunc: func(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error) {
				if limit != 10 || offset != 0 || orderBy != "created_at" || sortBy != "desc" {
					return nil, 0, errors.New("unexpected list arguments")
				}
				return []User{
					{ID: uuid.New(), Email: "john@doe.com"},
					{ID: uuid.New(), Email: "john@doe.com"},
				}, 2, nil
			},
		},
		{
			name:    "store error",
			limit:   10,
			orderBy: "created_at",
			sortBy:  "desc",
			mockFunc: func(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error) {
				return nil, 0, errors.New("database error")
			},
			wantErr: true,
		},
	}

	loggerMock := logger.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &mockStore{
				listFunc: tt.mockFunc,
			}
			service := NewService(mockStore, loggerMock)

			users, total, err := service.List(context.Background(), tt.limit, tt.offset, tt.orderBy, tt.sortBy)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if len(users) != len(tt.want) {
				t.Fatalf("returned %d users, want %d", len(users), len(tt.want))
			}
			for i, expected := range tt.want {
				if users[i].Email != expected.Email {
					t.Errorf("user %d email = %q, want %q", i, users[i].Email, expected.Email)
				}
			}
		})
	}
}
