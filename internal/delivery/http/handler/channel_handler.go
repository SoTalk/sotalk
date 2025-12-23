package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/channel"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// ChannelHandler handles channel HTTP requests
type ChannelHandler struct {
	channelService channel.Service
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(channelService channel.Service) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// CreateChannel handles POST /api/v1/channels
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.channelService.CreateChannel(c.Request.Context(), userID, &dto.CreateChannelRequest{
		Name:        req.Name,
		Username:    req.Username,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	})

	if err != nil {
		logger.Error("Failed to create channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_channel_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, response.CreateChannelResponse{
		Channel: mapChannelDTO(result.Channel),
	})
}

// GetChannel handles GET /api/v1/channels/:username
func (h *ChannelHandler) GetChannel(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	username := c.Param("username")

	result, err := h.channelService.GetChannel(c.Request.Context(), userID, username)
	if err != nil {
		logger.Error("Failed to get channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_channel_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.GetChannelResponse{
		Channel:      mapChannelDTO(result.Channel),
		IsSubscribed: result.IsSubscribed,
		IsOwner:      result.IsOwner,
		IsAdmin:      result.IsAdmin,
	})
}

// GetPublicChannels handles GET /api/v1/channels/public
func (h *ChannelHandler) GetPublicChannels(c *gin.Context) {
	var req request.GetChannelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.channelService.GetPublicChannels(c.Request.Context(), req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get public channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_channels_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	channels := make([]response.ChannelDTO, len(result.Channels))
	for i, ch := range result.Channels {
		channels[i] = mapChannelDTO(ch)
	}

	c.JSON(http.StatusOK, response.GetChannelsResponse{
		Channels: channels,
	})
}

// GetUserChannels handles GET /api/v1/channels/owned
func (h *ChannelHandler) GetUserChannels(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.GetChannelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.channelService.GetUserChannels(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get user channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_channels_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	channels := make([]response.ChannelDTO, len(result.Channels))
	for i, ch := range result.Channels {
		channels[i] = mapChannelDTO(ch)
	}

	c.JSON(http.StatusOK, response.GetChannelsResponse{
		Channels: channels,
	})
}

// GetSubscriptions handles GET /api/v1/channels/subscriptions
func (h *ChannelHandler) GetSubscriptions(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.GetChannelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.channelService.GetUserSubscriptions(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_subscriptions_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	channels := make([]response.ChannelDTO, len(result.Channels))
	for i, ch := range result.Channels {
		channels[i] = mapChannelDTO(ch)
	}

	c.JSON(http.StatusOK, response.GetChannelsResponse{
		Channels: channels,
	})
}

// Subscribe handles POST /api/v1/channels/:username/subscribe
func (h *ChannelHandler) Subscribe(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	username := c.Param("username")

	err = h.channelService.Subscribe(c.Request.Context(), userID, username)
	if err != nil {
		logger.Error("Failed to subscribe", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "subscribe_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscribed"})
}

// Unsubscribe handles POST /api/v1/channels/:username/unsubscribe
func (h *ChannelHandler) Unsubscribe(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	username := c.Param("username")

	err = h.channelService.Unsubscribe(c.Request.Context(), userID, username)
	if err != nil {
		logger.Error("Failed to unsubscribe", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "unsubscribe_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
}

// DeleteChannel handles DELETE /api/v1/channels/:id
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_channel_id",
			Message: "Invalid channel ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.channelService.DeleteChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		logger.Error("Failed to delete channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_channel_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "channel deleted"})
}

// Helper function to map channel DTO
func mapChannelDTO(ch dto.ChannelDTO) response.ChannelDTO {
	var settings *response.ChannelSettings
	if ch.Settings != nil {
		settings = &response.ChannelSettings{
			AdminsCanPost:     ch.Settings.AdminsCanPost,
			LinkPreview:       ch.Settings.LinkPreview,
			ForwardingAllowed: ch.Settings.ForwardingAllowed,
		}
	}

	return response.ChannelDTO{
		ID:              ch.ID,
		ConversationID:  ch.ConversationID,
		Name:            ch.Name,
		Username:        ch.Username,
		Description:     ch.Description,
		Avatar:          ch.Avatar,
		OwnerID:         ch.OwnerID,
		IsPublic:        ch.IsPublic,
		SubscriberCount: ch.SubscriberCount,
		Settings:        settings,
		CreatedAt:       ch.CreatedAt,
		UpdatedAt:       ch.UpdatedAt,
	}
}
