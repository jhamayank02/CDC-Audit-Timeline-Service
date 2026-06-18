package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	_ "github.com/lib/pq"
)

func NewDB(cfg config.DBConfig, logger *slog.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	logger.Debug("connecting to db", "db_name", cfg.Name, "host", cfg.Host, "port", cfg.Port)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	logger.Info("connected to db", "db_name", cfg.Name, "host", cfg.Host, "port", cfg.Port)
	return db, nil
}
