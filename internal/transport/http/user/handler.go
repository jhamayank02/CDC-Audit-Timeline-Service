package userhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
	httpmiddleware "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/transport/http/middleware"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/validation"
)

type Handler struct {
	service user.Service
	logger  *slog.Logger
}

func NewHandler(service user.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, validation.ErrorResponse{Message: "validation failed", Errors: validation.FormatValidationErrors(err)})
		return
	}

	result, err := h.service.Create(c, user.CreateInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		CreatedBy: c.Value(httpmiddleware.RequestorIDContextKey).(string),
	})
	if err != nil {
		h.logger.Error("[HANDLER] failed to create user", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("[HANDLER] failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, validation.ErrorResponse{Message: "validation failed", Errors: validation.FormatValidationErrors(err)})
		return
	}

	if strings.TrimSpace(req.FirstName) == "" && strings.TrimSpace(req.LastName) == "" && strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return
	}

	id := c.Param("id")
	if !isUUID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	result, err := h.service.Update(c, user.UpdateInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		UpdatedBy: c.Value(httpmiddleware.RequestorIDContextKey).(string),
	})
	if err != nil {
		h.logger.Error("[HANDLER] failed to update user", "err", err)
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": user.ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if !isUUID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidID.Error()})
		return
	}

	result, err := h.service.GetByID(c, id)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get user", "err", err)
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": user.ErrUserNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetUsers(c *gin.Context) {
	limit, page, orderBy, sortBy, ok := parseListQuery(c, map[string]bool{
		"id": true, "first_name": true, "last_name": true, "email": true, "created_at": true, "updated_at": true,
	}, user.ErrInvalidOrderBy)
	if !ok {
		return
	}

	offset := (page - 1) * limit
	users, totalCount, err := h.service.List(c, limit, offset, orderBy, sortBy)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get users", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "total_results": totalCount})
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
