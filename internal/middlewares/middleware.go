package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestorIDContextKey = "requestor_id"

func RequestorContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestorId := c.GetHeader("X-Requestor-Id")
		if requestorId == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "X-Requestor-Id header is required",
			})
			return
		}
		if _, err := uuid.Parse(requestorId); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid requestor id",
			})
			return
		}
		c.Set(RequestorIDContextKey, requestorId)
		c.Next()
	}
}
