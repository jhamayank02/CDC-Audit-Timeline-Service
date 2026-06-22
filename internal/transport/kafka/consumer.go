package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafkago.Reader
	auditor auditlog.Service
	logger  *slog.Logger
}

func NewConsumer(cfg config.KafkaConfig, auditor auditlog.Service, logger *slog.Logger) *Consumer {
	readerConfig := kafkago.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
	}

	if len(cfg.Topics) > 1 {
		readerConfig.GroupTopics = cfg.Topics
	} else if len(cfg.Topics) == 1 {
		readerConfig.Topic = cfg.Topics[0]
	} else {
		readerConfig.Topic = cfg.Topic
	}

	reader := kafkago.NewReader(readerConfig)

	return &Consumer{
		reader:  reader,
		auditor: auditor,
		logger:  logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Error("failed to read message", "err", err)
			continue
		}

		c.processMessage(ctx, msg)
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafkago.Message) {
	var event DebeziumEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("failed to parse event", "err", err)
		return
	}

	c.logger.Info("event received", "topic", msg.Topic, "table", event.Payload.Source.Table, "operation", event.Payload.Op)
	if event.Payload.Op == "r" {
		c.logger.Info("skipping read event")
		return
	}
	if err := c.auditor.Record(ctx, event.AuditLog()); err != nil {
		return
	}
	c.logger.Info("audit log inserted successfully")
}
