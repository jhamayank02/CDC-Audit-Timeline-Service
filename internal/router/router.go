package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/middlewares"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/subscription"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/user"
)

func Regiser(r *gin.Engine, middleware *middlewares.Middleware, userHandler *user.Handler, subscriptionHandler *subscription.Handler) {
	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	api.Use(middleware.RequestorContext())

	userApi := api.Group("/users")
	userApi.POST("/", userHandler.CreateUser)
	userApi.PUT("/:id", userHandler.UpdateUser)
	userApi.GET("/:id", userHandler.GetUser)
	userApi.GET("/", userHandler.GetUsers)

	subscriptionApi := api.Group("/subscriptions")
	subscriptionApi.POST("/", subscriptionHandler.CreateSubscription)
	subscriptionApi.PUT("/:id", subscriptionHandler.UpdateSubscription)
	subscriptionApi.GET("/:id", subscriptionHandler.GetSubscription)
	subscriptionApi.GET("/", subscriptionHandler.GetSubscriptions)
}
