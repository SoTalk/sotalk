package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/notification"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	notificationService notification.Service
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService notification.Service) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications handles GET /api/v1/notifications
// @Summary Get user notifications
// @Description Gets user notifications with pagination
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.NotificationsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.notificationService.GetNotifications(c.Request.Context(), &dto.GetNotificationsRequest{
		UserID: userID.(string),
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_notifications_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	notifications := make([]response.NotificationDTO, len(result.Notifications))
	for i, n := range result.Notifications {
		notifications[i] = response.NotificationDTO{
			ID:        n.ID,
			Type:      n.Type,
			Title:     n.Title,
			Body:      n.Body,
			Data:      n.Data,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response.NotificationsResponse{
		Notifications: notifications,
		Total:         result.Total,
		Unread:        result.Unread,
	})
}

// GetUnreadCount handles GET /api/v1/notifications/unread
// @Summary Get unread notification count
// @Description Gets the count of unread notifications
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.UnreadCountResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/unread [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get unread count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_unread_count_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.UnreadCountResponse{
		Count: count,
	})
}

// MarkAsRead handles POST /api/v1/notifications/:id/read
// @Summary Mark notification as read
// @Description Marks a specific notification as read
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Notification ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err := h.notificationService.MarkAsRead(c.Request.Context(), &dto.MarkNotificationAsReadRequest{
		UserID:         userID.(string),
		NotificationID: notificationID,
	})

	if err != nil {
		logger.Error("Failed to mark as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "mark_as_read_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Notification marked as read",
	})
}

// MarkAllAsRead handles POST /api/v1/notifications/read-all
// @Summary Mark all notifications as read
// @Description Marks all notifications as read for the current user
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to mark all as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "mark_all_as_read_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "All notifications marked as read",
	})
}

// GetSettings handles GET /api/v1/notifications/settings
// @Summary Get notification settings
// @Description Gets user notification preferences
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.NotificationSettingsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/settings [get]
func (h *NotificationHandler) GetSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	result, err := h.notificationService.GetSettings(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get notification settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_settings_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.NotificationSettingsResponse{
		MessagesEnabled:  result.MessagesEnabled,
		GroupsEnabled:    result.GroupsEnabled,
		ChannelsEnabled:  result.ChannelsEnabled,
		PaymentsEnabled:  result.PaymentsEnabled,
		MentionsEnabled:  result.MentionsEnabled,
		ReactionsEnabled: result.ReactionsEnabled,
		SoundEnabled:     result.SoundEnabled,
		VibrationEnabled: result.VibrationEnabled,
	})
}

// UpdateSettings handles PUT /api/v1/notifications/settings
// @Summary Update notification settings
// @Description Updates user notification preferences
// @Tags notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdateNotificationSettingsRequest true "Update Settings Request"
// @Success 200 {object} response.NotificationSettingsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/settings [put]
func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req request.UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.notificationService.UpdateSettings(c.Request.Context(), &dto.UpdateNotificationSettingsRequest{
		UserID:           userID.(string),
		MessagesEnabled:  req.MessagesEnabled,
		GroupsEnabled:    req.GroupsEnabled,
		ChannelsEnabled:  req.ChannelsEnabled,
		PaymentsEnabled:  req.PaymentsEnabled,
		MentionsEnabled:  req.MentionsEnabled,
		ReactionsEnabled: req.ReactionsEnabled,
		SoundEnabled:     req.SoundEnabled,
		VibrationEnabled: req.VibrationEnabled,
	})

	if err != nil {
		logger.Error("Failed to update notification settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_settings_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.NotificationSettingsResponse{
		MessagesEnabled:  result.MessagesEnabled,
		GroupsEnabled:    result.GroupsEnabled,
		ChannelsEnabled:  result.ChannelsEnabled,
		PaymentsEnabled:  result.PaymentsEnabled,
		MentionsEnabled:  result.MentionsEnabled,
		ReactionsEnabled: result.ReactionsEnabled,
		SoundEnabled:     result.SoundEnabled,
		VibrationEnabled: result.VibrationEnabled,
	})
}

// DeleteNotification handles DELETE /api/v1/notifications/:id
// @Summary Delete notification
// @Description Deletes a specific notification
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Notification ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err := h.notificationService.DeleteNotification(c.Request.Context(), &dto.DeleteNotificationRequest{
		UserID:         userID.(string),
		NotificationID: notificationID,
	})

	if err != nil {
		logger.Error("Failed to delete notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_notification_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Notification deleted successfully",
	})
}
