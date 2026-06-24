package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

// NewTestEnv loads the PostgreSQL test-container settings used by integration
// tests. It searches parent directories because Go executes package tests from
// the package directory, rather than necessarily from the repository root.
// A local postgres.test.env takes precedence; the committed example supplies
// safe defaults when no local override exists.
func NewTestEnv(logger *slog.Logger) (*Env, error) {
	envPath, err := findTestEnvFile()
	if err != nil {
		return nil, err
	}
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("load test environment file %q: %w", envPath, err)
	}

	return &Env{
		logger: logger,
	}, nil
}

func findTestEnvFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		for _, name := range []string{"postgres.test.env", "postgres.test.example.env"} {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("check test environment file %q: %w", path, err)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("postgres.test.env or postgres.test.example.env not found")
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
