package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	logger *slog.Logger
}

func NewEnv(logger *slog.Logger) (*Env, error) {
	err := godotenv.Load()
	if err != nil {
		logger.Warn("failed to load .env file, continuing with process environment", "error", err)
	}

	return &Env{
		logger: logger,
	}, nil
}

func (e *Env) GetString(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		e.logger.Debug("environment variable not found, using fallback", "key", key, "fallback", fallback)
		return fallback
	}

	return value
}

func (e *Env) GetStringSlice(key string, fallback []string) []string {
	value := e.GetString(key, "")
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func (e *Env) GetInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		e.logger.Debug("environment variable not found, using fallback", "key", key, "fallback", fallback)
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		e.logger.Warn("invalid integer environment variable, using fallback", "key", key, "value", value, "fallback", fallback, "error", err)
		return fallback
	}

	return result
}

func (e *Env) GetBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		e.logger.Debug("environment variable not found, using fallback", "key", key, "fallback", fallback)
		return fallback
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		e.logger.Warn("invalid boolean environment variable, using fallback", "key", key, "value", value, "fallback", fallback, "error", err)
		return fallback
	}

	return result
}
