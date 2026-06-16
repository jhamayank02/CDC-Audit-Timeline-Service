package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidId           = errors.New("id must be in valid uuid format")
	ErrInvalidSortBy       = errors.New("sortBy must be asc or desc")
	ErrInvalidOrderBy      = errors.New("orderBy must be one of id, first_name, last_name, email, created_at, updated_at")
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type CreateUserReq struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	CreatedBy string `json:"created_by"`
}

type UpdateUserReq struct {
	Id        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	UpdatedBy string `json:"updated_by"`
}

type Repository interface {
	CreateUser(ctx context.Context, req *CreateUserReq) (*User, error)
	UpdateUser(ctx context.Context, req *UpdateUserReq) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUsers(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error)
}

type Service interface {
	CreateUser(ctx context.Context, req *CreateUserReq) (*User, error)
	UpdateUser(ctx context.Context, req *UpdateUserReq) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUsers(ctx context.Context, limit int, offset int, orderBy, sortBy string) ([]User, int, error)
}
