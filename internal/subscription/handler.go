package subscription

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidUserId.Error()})
		return
	}
	h.logger.Info("[HANDLER] creating subscription", "req", req)

	subscription, err := h.service.CreateSubscription(c, &req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to create subscription", "err", err)
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] subscription created", "subscription", subscription)
	c.JSON(http.StatusCreated, subscription)
}

func (h *Handler) UpdateSubscription(c *gin.Context) {
	var req UpdateSubscriptionReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if len(strings.TrimSpace(id)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidId.Error()})
		return
	}
	_, err = uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidId.Error()})
		return
	}
	req.Id = id
	h.logger.Info("[HANDLER] updating subscription", "req", req)

	subscription, err := h.service.UpdateSubscription(c, &req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to update subscription", "err", err)
		if errors.Is(err, ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrSubscriptionNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] subscription updated", "subscription", subscription)
	c.JSON(http.StatusOK, subscription)
}

func (h *Handler) GetSubscription(c *gin.Context) {
	id := c.Param("id")
	if len(strings.TrimSpace(id)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidId.Error()})
		return
	}
	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidId.Error()})
		return
	}

	h.logger.Info("[HANDLER] getting subscription", "id", id)

	subscription, err := h.service.GetSubscription(c, id)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get subscription", "err", err)
		if errors.Is(err, ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrSubscriptionNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] got subscription", "subscription", subscription)
	c.JSON(http.StatusOK, subscription)
}

func (h *Handler) GetSubscriptions(c *gin.Context) {
	limit := c.Query("limit")
	page := c.Query("page")
	orderBy := c.Query("orderBy")
	sortBy := c.Query("sortBy")

	if limit == "" {
		limit = "10"
	}
	if page == "" {
		page = "1"
	}
	if orderBy == "" {
		orderBy = "created_at"
	}
	if sortBy == "" {
		sortBy = "asc"
	}

	limitInt, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		h.logger.Info("[HANDLER] invalid limit", "limit", limit)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}
	pageInt, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		h.logger.Info("[HANDLER] invalid page", "page", page)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	validOrderByFields := map[string]bool{
		"id":         true,
		"user_id":    true,
		"plan_name":  true,
		"status":     true,
		"start_date": true,
		"end_date":   true,
		"auto_renew": true,
		"created_at": true,
		"updated_at": true,
	}
	if !validOrderByFields[strings.ToLower(orderBy)] {
		h.logger.Info("[HANDLER] invalid orderBy", "orderBy", orderBy)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidOrderBy.Error()})
		return
	}

	validSortByFields := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if !validSortByFields[strings.ToLower(sortBy)] {
		h.logger.Info("[HANDLER] invalid sortBy", "sortBy", sortBy)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidSortBy.Error()})
		return
	}
	offset := (pageInt - 1) * limitInt

	h.logger.Info("[HANDLER] getting subscriptions", "limit", limitInt, "page", pageInt, "offset", offset, "orderBy", orderBy, "sortBy", sortBy)

	subscriptions, totalCount, err := h.service.GetSubscriptions(c, int(limitInt), int(offset), strings.ToLower(orderBy), strings.ToLower(sortBy))
	if err != nil {
		h.logger.Error("[HANDLER] failed to get subscriptions", "err", err)
		if errors.Is(err, ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrSubscriptionNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] got subscriptions", "subscriptions", subscriptions)
	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions, "total_results": totalCount})
}
