package config

import (
	"database/sql"
	"log/slog"

	"github.com/go-sql-driver/mysql"
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
	cfg := mysql.NewConfig()

	cfg.User = env.GetString("DB_USER", "root", logger)
	cfg.Passwd = env.GetString("DB_PASSWORD", "root", logger)
	cfg.Net = env.GetString("DB_PROTOCOL", "tcp", logger)
	cfg.Addr = env.GetString("DB_HOST", "localhost:3306", logger)
	cfg.DBName = env.GetString("DB_NAME", "test", logger)

	logger.Debug("Connecting to db", "db_name", cfg.DBName, "host", cfg.Addr)

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		logger.Error("Error connecting to db", "error", err)
		return nil, err
	}
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Error("Error pinging db", "error", pingErr)
		return nil, pingErr
	}

	logger.Info("Connected to db", "db_name", cfg.DBName, "host", cfg.Addr)
	return db, nil
}
