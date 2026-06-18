package user

import "context"

type Store interface {
	Create(ctx context.Context, input CreateInput) (*User, error)
	Update(ctx context.Context, input UpdateInput) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]User, int, error)
}
