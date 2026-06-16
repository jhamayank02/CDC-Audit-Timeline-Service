package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/user"
)

func Regiser(r *gin.Engine, userHandler *user.Handler) {
	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	userApi := api.Group("/users")
	userApi.POST("/", userHandler.CreateUser)
	userApi.PUT("/:id", userHandler.UpdateUser)
	userApi.GET("/:id", userHandler.GetUser)
	userApi.GET("/", userHandler.GetUsers)
}
