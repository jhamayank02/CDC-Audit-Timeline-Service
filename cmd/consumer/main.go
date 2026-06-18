package main

import (
	"context"
	"os"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/app"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/postgres"
)

func main() {
	log := logger.New()

	cfg, err := config.Load(log)
	if err != nil {
		log.Error("error loading config", "error", err)
		os.Exit(1)
	}

	db, err := postgres.NewDB(cfg.DB, log)
	if err != nil {
		log.Error("error initializing db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	consumer := app.NewConsumer(cfg.Kafka, db, log)
	if err := consumer.Run(context.Background()); err != nil {
		log.Error("consumer stopped", "error", err)
		os.Exit(1)
	}
}
