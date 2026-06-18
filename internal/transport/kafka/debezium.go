package kafka

import (
	"encoding/json"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
)

type DebeziumEvent struct {
	Payload Payload `json:"payload"`
}

type Payload struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	Source Source          `json:"source"`
	Op     string          `json:"op"`
}

type Source struct {
	Table  string `json:"table"`
	Schema string `json:"schema"`
}

func OperationFromDebezium(op string) auditlog.Operation {
	switch op {
	case "c":
		return auditlog.OperationCreate
	case "u":
		return auditlog.OperationUpdate
	case "d":
		return auditlog.OperationDelete
	case "r":
		return auditlog.OperationRead
	default:
		return auditlog.OperationUnknown
	}
}

func (e DebeziumEvent) AuditLog() auditlog.AuditLog {
	return auditlog.AuditLog{
		TableName: e.Payload.Source.Table,
		Operation: OperationFromDebezium(e.Payload.Op),
		Before:    e.Payload.Before,
		After:     e.Payload.After,
	}
}
