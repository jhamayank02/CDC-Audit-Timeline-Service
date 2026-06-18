package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
)

const RequestorIDContextKey = "requestor_id"

type RequestorMiddleware struct {
	userService user.Service
	logger      *slog.Logger
}

func NewRequestorMiddleware(userService user.Service, logger *slog.Logger) *RequestorMiddleware {
	return &RequestorMiddleware{
		userService: userService,
		logger:      logger,
	}
}

func (m *RequestorMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestorID := c.GetHeader("X-Requestor-Id")
		if requestorID == "" {
			m.logger.Error("[MIDDLEWARE] X-Requestor-Id header is required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "X-Requestor-Id header is required"})
			return
		}
		if _, err := uuid.Parse(requestorID); err != nil {
			m.logger.Error("[MIDDLEWARE] invalid requestor id")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid requestor id"})
			return
		}
		_, err := m.userService.GetByID(c, requestorID)
		if err != nil {
			m.logger.Error("[MIDDLEWARE] requestor not found", "requestor_id", requestorID, "err", err.Error())
			if errors.Is(err, user.ErrUserNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "requestor not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": apperrors.ErrInternalServerError.Error()})
			return
		}
		c.Set(RequestorIDContextKey, requestorID)
		c.Next()
	}
}
