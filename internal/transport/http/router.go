package http

import (
	"github.com/gin-gonic/gin"
	auditloghttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/auditlog"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
	subscriptionhttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/subscription"
	userhttp "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/user"
)

func Register(r *gin.Engine, requestorMiddleware *httpmiddleware.RequestorMiddleware, userHandler *userhttp.Handler, subscriptionHandler *subscriptionhttp.Handler, auditlogHandler *auditloghttp.Handler) {
	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	api.Use(requestorMiddleware.Handler())

	userAPI := api.Group("/users")
	userAPI.POST("/", userHandler.CreateUser)
	userAPI.PUT("/:id", userHandler.UpdateUser)
	userAPI.GET("/:id", userHandler.GetUser)
	userAPI.GET("/", userHandler.GetUsers)

	subscriptionAPI := api.Group("/subscriptions")
	subscriptionAPI.POST("/", subscriptionHandler.CreateSubscription)
	subscriptionAPI.PUT("/:id", subscriptionHandler.UpdateSubscription)
	subscriptionAPI.GET("/:id", subscriptionHandler.GetSubscription)
	subscriptionAPI.GET("/", subscriptionHandler.GetSubscriptions)

	auditlogAPI := api.Group("/audit-logs")
	auditlogAPI.GET("/", auditlogHandler.GetAuditLogs)
}
