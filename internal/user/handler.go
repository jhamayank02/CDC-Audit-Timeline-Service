package user

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/validation"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

const RequestorIDContextKey = "requestor_id"

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, validation.ErrorResponse{
			Message: "validation failed",
			Errors:  validation.FormatValidationErrors(err),
		})
		return
	}

	req.CreatedBy = c.Value(RequestorIDContextKey).(string)

	h.logger.Info("[HANDLER] creating user", "req", req)

	user, err := h.service.CreateUser(c, &req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to create user", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] user created", "user", user)
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var req UpdateUserReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, validation.ErrorResponse{
			Message: "validation failed",
			Errors:  validation.FormatValidationErrors(err),
		})
		return
	}

	if strings.TrimSpace(req.FirstName) == "" && strings.TrimSpace(req.LastName) == "" && strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
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
	req.UpdatedBy = c.Value(RequestorIDContextKey).(string)

	h.logger.Info("[HANDLER] updating user", "req", req)

	user, err := h.service.UpdateUser(c, &req)
	if err != nil {
		h.logger.Error("[HANDLER] failed to update user", "err", err)
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] user updated", "user", user)
	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUser(c *gin.Context) {
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

	h.logger.Info("[HANDLER] getting user", "id", id)

	user, err := h.service.GetUser(c, id)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get user", "err", err)
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] got user", "user", user)
	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUsers(c *gin.Context) {
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
		"first_name": true,
		"last_name":  true,
		"email":      true,
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

	h.logger.Info("[HANDLER] getting users", "limit", limitInt, "page", pageInt, "offset", offset, "orderBy", orderBy, "sortBy", sortBy)

	users, totalCount, err := h.service.GetUsers(c, int(limitInt), int(offset), strings.ToLower(orderBy), strings.ToLower(sortBy))
	if err != nil {
		h.logger.Error("[HANDLER] failed to get users", "err", err)
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	h.logger.Info("[HANDLER] got users", "users", users)
	c.JSON(http.StatusOK, gin.H{"users": users, "total_results": totalCount})
}
