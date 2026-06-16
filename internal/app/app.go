package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/router"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/subscription"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/validation"
)

type Config struct {
	Addr string
}

func NewConfig(addr string) *Config {
	return &Config{
		Addr: addr,
	}
}

type App struct {
	config *Config
	db     *sql.DB
	logger *slog.Logger
}

func NewApp(config *Config, db *sql.DB, logger *slog.Logger) *App {
	return &App{
		config: config,
		db:     db,
		logger: logger,
	}
}

func (a *App) Run() error {
	validation.RegisterJSONTagNameFunc()

	userRepo := user.NewRepository(a.db, a.logger)
	userService := user.NewService(userRepo, a.logger)
	userHandler := user.NewHandler(userService, a.logger)

	subscriptionRepo := subscription.NewRepository(a.db, a.logger)
	subscriptionService := subscription.NewService(subscriptionRepo, userService, a.logger)
	subscriptionHandler := subscription.NewHandler(subscriptionService, a.logger)

	// Initialize gin router
	engine := gin.New()
	// Use default logger and panic recovery middleware
	engine.Use(gin.Logger(), gin.Recovery())

	// Regiseter router
	router.Regiser(engine, userHandler, subscriptionHandler)

	server := http.Server{
		Addr:    a.config.Addr,
		Handler: engine,
	}

	return server.ListenAndServe()
}
