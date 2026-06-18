package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/config"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/subscription"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/postgres"
	httptransport "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
	subscriptionhttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/subscription"
	userhttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/validation"
)

type API struct {
	config config.HTTPConfig
	db     *sql.DB
	logger *slog.Logger
}

func NewAPI(config config.HTTPConfig, db *sql.DB, logger *slog.Logger) *API {
	return &API{config: config, db: db, logger: logger}
}

func (a *API) Run() error {
	validation.RegisterJSONTagNameFunc()

	userRepo := postgres.NewUserRepository(a.db, a.logger)
	userService := user.NewService(userRepo, a.logger)
	userHandler := userhttp.NewHandler(userService, a.logger)

	subscriptionRepo := postgres.NewSubscriptionRepository(a.db, a.logger)
	subscriptionService := subscription.NewService(subscriptionRepo, userService, a.logger)
	subscriptionHandler := subscriptionhttp.NewHandler(subscriptionService, a.logger)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	requestorMiddleware := httpmiddleware.NewRequestorMiddleware(userService, a.logger)
	httptransport.Register(engine, requestorMiddleware, userHandler, subscriptionHandler)

	server := http.Server{
		Addr:    a.config.Addr,
		Handler: engine,
	}

	return server.ListenAndServe()
}
