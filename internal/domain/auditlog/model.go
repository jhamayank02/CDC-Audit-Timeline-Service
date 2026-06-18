package auditlog

import (
	"encoding/json"
	"errors"
)

var (
	ErrInvalidOrderBy = errors.New("orderBy must be one of id, table_name, operation, created_at")
)

type Operation string

const (
	OperationCreate  Operation = "create"
	OperationUpdate  Operation = "update"
	OperationDelete  Operation = "delete"
	OperationRead    Operation = "read"
	OperationUnknown Operation = "unknown"
)

type AuditLog struct {
	Id        string          `json:"id"`
	TableName string          `json:"table_name"`
	Operation Operation       `json:"operation"`
	Before    json.RawMessage `json:"before"`
	After     json.RawMessage `json:"after"`
	CreatedAt string          `json:"created_at"`
}
