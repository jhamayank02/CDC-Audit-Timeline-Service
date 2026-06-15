package main

import (
	"log/slog"
	"os"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/app"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
)

func main() {
	// Initialize logger
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}),
	)

	// Load env variables
	envVars := config.NewEnv(logger)
	if envVars == nil {
		logger.Error("Error loading env variables")
		os.Exit(1)
	}

	// Load address from env
	addr := envVars.GetString("PORT", ":8080", logger)

	cfg := app.NewConfig(addr)
	server := app.NewApp(cfg, logger)

	err := server.Run()
	if err != nil {
		logger.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}
