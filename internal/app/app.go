package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/router"
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
	logger *slog.Logger
}

func NewApp(config *Config, logger *slog.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (a *App) Run() error {
	// Initialize gin router
	engine := gin.New()
	// Use default logger and panic recovery middleware
	engine.Use(gin.Logger(), gin.Recovery())

	// Regiseter router
	router.Regiser(engine)

	server := http.Server{
		Addr:    a.config.Addr,
		Handler: engine,
	}

	return server.ListenAndServe()
}
