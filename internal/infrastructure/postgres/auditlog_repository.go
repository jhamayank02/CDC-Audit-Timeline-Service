package postgres

import (
	"context"
	"database/sql"
	"fmt"
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

const (
	createAuditLogQuery = `
		INSERT INTO audit_logs (id, table_name, operation, before, after, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`
	getAuditLogsBaseQuery = `
		SELECT id, table_name, operation, before, after, created_at, COUNT(*) OVER() AS total_count
		FROM audit_logs
	`
)

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

func (r *AuditLogRepository) Get(ctx context.Context, limit, offset int, orderBy, sortBy string) ([]auditlog.AuditLog, int, error) {
	query := fmt.Sprintf("%s ORDER BY %s %s OFFSET $1 LIMIT $2", getAuditLogsBaseQuery, orderBy, sortBy)

	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		r.logger.Error("[REPOSITORY] failed to get audit logs", "err", err)
		return nil, 0, err
	}
	defer rows.Close()

	auditLogs := make([]auditlog.AuditLog, 0)
	totalCount := 0
	for rows.Next() {
		var auditLog auditlog.AuditLog
		err := rows.Scan(&auditLog.Id, &auditLog.TableName, &auditLog.Operation, &auditLog.Before, &auditLog.After, &auditLog.CreatedAt, &totalCount)
		if err != nil {
			r.logger.Error("[REPOSITORY] failed to scan audit log", "err", err)
			return nil, 0, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	return auditLogs, totalCount, rows.Err()
}

func normalizeJSON(value []byte) string {
	if len(value) == 0 {
		return "null"
	}
	return string(value)
}
