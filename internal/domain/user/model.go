package user

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidOrderBy = errors.New("orderBy must be one of id, first_name, last_name, email, created_at, updated_at")
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

type CreateInput struct {
	FirstName string
	LastName  string
	Email     string
	CreatedBy string
}

type UpdateInput struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	UpdatedBy string
}
