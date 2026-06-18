package auditloghttp

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/auditlog"
	apperrors "github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/errors"
)

type Handler struct {
	service auditlog.Service
	logger  *slog.Logger
}

func NewHandler(service auditlog.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) GetAuditLogs(c *gin.Context) {
	limit, page, orderBy, sortBy, ok := parseListQuery(c, map[string]bool{
		"id": true, "table_name": true, "operation": true, "created_at": true,
	}, auditlog.ErrInvalidOrderBy)
	if !ok {
		return
	}

	offset := (page - 1) * limit
	auditLogs, totalCount, err := h.service.GetAuditLogs(c, limit, offset, orderBy, sortBy)
	if err != nil {
		h.logger.Error("[HANDLER] failed to get audit logs", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": auditLogs, "total_results": totalCount})
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
