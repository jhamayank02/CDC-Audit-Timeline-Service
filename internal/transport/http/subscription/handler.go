package subscriptionhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/subscription"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/validation"
)

type Handler struct {
	service subscription.Service
	logger  *slog.Logger
}

func NewHandler(service subscription.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, validation.ErrorResponse{Message: "validation failed", Errors: validation.FormatValidationErrors(err)})
		return
	}

	result, err := h.service.Create(c, subscription.CreateInput{
		UserID:    req.UserID,
		PlanName:  req.PlanName,
		Status:    req.Status,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		AutoRenew: req.AutoRenew,
		CreatedBy: c.Value(httpmiddleware.RequestorIDContextKey).(string),
	})
	if err != nil {
		h.logger.Error("[HANDLER] failed to create subscription", "err", err)
		if errors.Is(err, subscription.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": subscription.ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) UpdateSubscription(c *gin.Context) {
	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, validation.ErrorResponse{Message: "validation failed", Errors: validation.FormatValidationErrors(err)})
		return
	}

	id := c.Param("id")
	if !isUUID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	result, err := h.service.Update(c, subscription.UpdateInput{
		ID:        id,
		Status:    req.Status,
		AutoRenew: req.AutoRenew,
		UpdatedBy: c.Value(httpmiddleware.RequestorIDContextKey).(string),
	})
	if err != nil {
		h.logger.Error("[HANDLER] failed to update subscription", "err", err)
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": subscription.ErrSubscriptionNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetSubscription(c *gin.Context) {
	id := c.Param("id")
	if !isUUID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	result, err := h.service.GetByID(c, id)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get subscription", "err", err)
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": subscription.ErrSubscriptionNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetSubscriptions(c *gin.Context) {
	limit, page, orderBy, sortBy, ok := parseListQuery(c, map[string]bool{
		"id": true, "user_id": true, "plan_name": true, "status": true, "start_date": true, "end_date": true, "auto_renew": true, "created_at": true, "updated_at": true,
	}, subscription.ErrInvalidOrderBy)
	if !ok {
		return
	}

	offset := (page - 1) * limit
	subscriptions, totalCount, err := h.service.List(c, limit, offset, orderBy, sortBy)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get subscriptions", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions, "total_results": totalCount})
}

func parseListQuery(c *gin.Context, allowedOrderBy map[string]bool, orderByErr error) (int, int, string, string, bool) {
	limitValue := defaultString(c.Query("limit"), "10")
	pageValue := defaultString(c.Query("page"), "1")
	orderBy := strings.ToLower(defaultString(c.Query("orderBy"), "created_at"))
	sortBy := strings.ToLower(defaultString(c.Query("sortBy"), "asc"))

	limit, err := strconv.Atoi(limitValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return 0, 0, "", "", false
	}
	page, err := strconv.Atoi(pageValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return 0, 0, "", "", false
	}
	if !allowedOrderBy[orderBy] {
		c.JSON(http.StatusBadRequest, gin.H{"error": orderByErr.Error()})
		return 0, 0, "", "", false
	}
	if sortBy != "asc" && sortBy != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidSortBy.Error()})
		return 0, 0, "", "", false
	}

	return limit, page, orderBy, sortBy, true
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func isUUID(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	_, err := uuid.Parse(value)
	return err == nil
}
