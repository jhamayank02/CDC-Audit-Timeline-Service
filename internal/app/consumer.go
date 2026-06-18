package app

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/postgres"
	kafkatransport "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/kafka"
)

type Consumer struct {
	config config.KafkaConfig
	db     *sql.DB
	logger *slog.Logger
}

func NewConsumer(config config.KafkaConfig, db *sql.DB, logger *slog.Logger) *Consumer {
	return &Consumer{config: config, db: db, logger: logger}
}

func (c *Consumer) Run(ctx context.Context) error {
	auditLogRepo := postgres.NewAuditLogRepository(c.db, c.logger)
	auditLogService := auditlog.NewService(auditLogRepo, c.logger)
	consumer := kafkatransport.NewConsumer(c.config, auditLogService, c.logger)

	c.logger.Info("consumer process started", "brokers", c.config.Brokers, "topics", c.config.Topics, "group_id", c.config.GroupID)
	return consumer.Run(ctx)
}
