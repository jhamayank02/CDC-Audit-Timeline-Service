package auditloghttp

import "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"

type AuditLogRes struct {
	Id        string             `json:"id"`
	TableName string             `json:"table_name"`
	Operation auditlog.Operation `json:"operation"`
	Before    string             `json:"before"`
	After     string             `json:"after"`
	CreatedAt string             `json:"created_at"`
}
