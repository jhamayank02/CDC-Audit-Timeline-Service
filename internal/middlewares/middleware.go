package middlewares

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/user"
)

const RequestorIDContextKey = "requestor_id"

type Middleware struct {
	userService user.Service
	logger      *slog.Logger
}

func NewMiddleware(userService user.Service, logger *slog.Logger) *Middleware {
	return &Middleware{
		userService: userService,
		logger:      logger,
	}
}

func (m *Middleware) RequestorContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestorId := c.GetHeader("X-Requestor-Id")
		if requestorId == "" {
			m.logger.Error("[MIDDLEWARE] X-Requestor-Id header is required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "X-Requestor-Id header is required",
			})
			return
		}
		if _, err := uuid.Parse(requestorId); err != nil {
			m.logger.Error("[MIDDLEWARE] invalid requestor id")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid requestor id",
			})
			return
		}
		_, err := m.userService.GetUser(c, requestorId)
		if err != nil {
			m.logger.Error("[MIDDLEWARE] requestor not found", "requestor_id", requestorId, "err", err.Error())
			if errors.Is(err, user.ErrUserNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "requestor not found",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": user.ErrInternalServerError.Error(),
			})
			return
		}
		c.Set(RequestorIDContextKey, requestorId)
		c.Next()
	}
}
