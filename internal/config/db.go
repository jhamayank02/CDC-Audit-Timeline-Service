package config

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

type DB struct {
	DB        *sql.DB
	envConfig *Env
	logger    *slog.Logger
}

func NewDB(env *Env, logger *slog.Logger) *DB {
	db, err := setupDB(env, logger)
	if err != nil {
		logger.Error("Error setting up db", "error", err)
		return nil
	}

	return &DB{
		DB:        db,
		envConfig: env,
		logger:    logger,
	}
}

func setupDB(env *Env, logger *slog.Logger) (*sql.DB, error) {
	user := env.GetString("DB_USER", "postgres", logger)
	password := env.GetString("DB_PASSWORD", "postgres", logger)
	host := env.GetString("DB_HOST", "localhost", logger)
	port := env.GetString("DB_PORT", "5432", logger)
	dbName := env.GetString("DB_NAME", "postgres", logger)
	sslMode := env.GetString("DB_SSLMODE", "disable", logger)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslMode)

	logger.Debug("Connecting to db", "db_name", dbName, "host", host, "port", port)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Error connecting to db", "error", err)
		return nil, err
	}
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Error("Error pinging db", "error", pingErr)
		return nil, pingErr
	}

	logger.Info("Connected to db", "db_name", dbName, "host", host, "port", port)
	return db, nil
}
