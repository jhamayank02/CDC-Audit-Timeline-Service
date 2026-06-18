package subscription

import "context"

type Store interface {
	Create(ctx context.Context, input CreateInput) (*Subscription, error)
	Update(ctx context.Context, input UpdateInput) (*Subscription, error)
	GetByID(ctx context.Context, id string) (*Subscription, error)
	List(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]Subscription, int, error)
}
