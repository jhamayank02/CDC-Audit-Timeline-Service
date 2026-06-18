package auditlog

import "context"

type Store interface {
	Create(ctx context.Context, auditLog AuditLog) error
}
