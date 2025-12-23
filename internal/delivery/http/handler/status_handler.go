package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/status"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// StatusHandler handles status/stories HTTP requests
type StatusHandler struct {
	statusService status.Service
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(statusService status.Service) *StatusHandler {
	return &StatusHandler{
		statusService: statusService,
	}
}

// CreateStatus handles POST /api/v1/statuses
func (h *StatusHandler) CreateStatus(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.CreateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.statusService.CreateStatus(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("Failed to create status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create status",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetUserStatuses handles GET /api/v1/statuses/user/:userId
func (h *StatusHandler) GetUserStatuses(c *gin.Context) {
	viewerIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	viewerID, err := uuid.Parse(viewerIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid viewer ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	statuses, err := h.statusService.GetUserStatuses(c.Request.Context(), viewerID, targetUserID)
	if err != nil {
		logger.Error("Failed to get user statuses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get statuses",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}

// GetStatusFeed handles GET /api/v1/statuses/feed
func (h *StatusHandler) GetStatusFeed(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse pagination parameters
	var limit, offset int
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	statuses, err := h.statusService.GetStatusFeed(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get status feed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get status feed",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}

// ViewStatus handles POST /api/v1/statuses/:id/view
func (h *StatusHandler) ViewStatus(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	viewerID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid viewer ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid status ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.statusService.ViewStatus(c.Request.Context(), statusID, viewerID); err != nil {
		logger.Error("Failed to view status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to view status",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status viewed"})
}

// GetStatusViews handles GET /api/v1/statuses/:id/views
func (h *StatusHandler) GetStatusViews(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid status ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	views, err := h.statusService.GetStatusViews(c.Request.Context(), userID, statusID)
	if err != nil {
		logger.Error("Failed to get status views", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get status views",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"views": views})
}

// DeleteStatus handles DELETE /api/v1/statuses/:id
func (h *StatusHandler) DeleteStatus(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid status ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.statusService.DeleteStatus(c.Request.Context(), userID, statusID); err != nil {
		logger.Error("Failed to delete status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete status",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status deleted"})
}
