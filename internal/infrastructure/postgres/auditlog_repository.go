package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
)

type AuditLogRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAuditLogRepository(db *sql.DB, logger *slog.Logger) *AuditLogRepository {
	return &AuditLogRepository{db: db, logger: logger}
}

const createAuditLogQuery = `
	INSERT INTO audit_logs (id, table_name, operation, before, after, created_at)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
`

func (r *AuditLogRepository) Create(ctx context.Context, logEntry auditlog.AuditLog) error {
	before := normalizeJSON(logEntry.Before)
	after := normalizeJSON(logEntry.After)

	_, err := r.db.ExecContext(ctx, createAuditLogQuery, uuid.New(), logEntry.TableName, string(logEntry.Operation), before, after)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to insert audit log", "err", err)
		return err
	}
	return nil
}

func normalizeJSON(value []byte) string {
	if len(value) == 0 {
		return "null"
	}
	return string(value)
}
