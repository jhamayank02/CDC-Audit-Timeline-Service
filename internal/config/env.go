package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	logger *slog.Logger
}

func NewEnv(logger *slog.Logger) *Env {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
		return nil
	}

	return &Env{
		logger: logger,
	}
}

func (e *Env) GetString(key string, fallback string, logger *slog.Logger) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		logger.Debug("environment variable not found, using fallback", "key", key, "fallback", fallback)
		return fallback
	}

	return value
}

func (e *Env) GetInt(key string, fallback int, logger *slog.Logger) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		logger.Debug("environment variable not found, using fallback", "key", key, "fallback", fallback)
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		logger.Warn("invalid integer environment variable, using fallback", "key", key, "value", value, "fallback", fallback, "error", err)
		return fallback
	}

	return result
}

func (e *Env) GetBool(key string, fallback bool, logger *slog.Logger) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		logger.Debug("environment variable not found, using fallback", "key", key, "fallback", fallback)
		return fallback
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		logger.Warn("invalid boolean environment variable, using fallback", "key", key, "value", value, "fallback", fallback, "error", err)
		return fallback
	}

	return result
}
