package auditlog

import (
	"context"
)

type Store interface {
	Create(ctx context.Context, auditLog AuditLog) error
	Get(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]AuditLog, int, error)
}
