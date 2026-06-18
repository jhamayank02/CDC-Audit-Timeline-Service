package main

import (
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

	server := app.NewAPI(cfg.HTTP, db, log)
	if err := server.Run(); err != nil {
		log.Error("error starting server", "error", err)
		os.Exit(1)
	}
}
