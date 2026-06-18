package auditlog

import "encoding/json"

type Operation string

const (
	OperationCreate  Operation = "create"
	OperationUpdate  Operation = "update"
	OperationDelete  Operation = "delete"
	OperationRead    Operation = "read"
	OperationUnknown Operation = "unknown"
)

type AuditLog struct {
	TableName string
	Operation Operation
	Before    json.RawMessage
	After     json.RawMessage
}
