package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/privacy"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// PrivacyHandler handles privacy HTTP requests
type PrivacyHandler struct {
	privacyService privacy.Service
}

// NewPrivacyHandler creates a new privacy handler
func NewPrivacyHandler(privacyService privacy.Service) *PrivacyHandler {
	return &PrivacyHandler{
		privacyService: privacyService,
	}
}

// GetPrivacySettings handles GET /api/v1/privacy/settings
func (h *PrivacyHandler) GetPrivacySettings(c *gin.Context) {
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

	settings, err := h.privacyService.GetPrivacySettings(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get privacy settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_settings_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdatePrivacySettings handles PUT /api/v1/privacy/settings
func (h *PrivacyHandler) UpdatePrivacySettings(c *gin.Context) {
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

	var req dto.PrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	settings, err := h.privacyService.UpdatePrivacySettings(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("Failed to update privacy settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_settings_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// BlockUser handles POST /api/v1/privacy/block
func (h *PrivacyHandler) BlockUser(c *gin.Context) {
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

	var req dto.BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.privacyService.BlockUser(c.Request.Context(), userID, &req); err != nil {
		logger.Error("Failed to block user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "block_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user blocked successfully"})
}

// UnblockUser handles DELETE /api/v1/privacy/block/:userId
func (h *PrivacyHandler) UnblockUser(c *gin.Context) {
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

	blockedUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_blocked_user_id",
			Message: "Invalid blocked user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.privacyService.UnblockUser(c.Request.Context(), userID, blockedUserID); err != nil {
		logger.Error("Failed to unblock user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "unblock_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unblocked successfully"})
}

// GetBlockedUsers handles GET /api/v1/privacy/blocked
func (h *PrivacyHandler) GetBlockedUsers(c *gin.Context) {
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

	blocked, err := h.privacyService.GetBlockedUsers(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get blocked users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_blocked_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blocked_users": blocked})
}

// SetDisappearingMessages handles POST /api/v1/privacy/disappearing
func (h *PrivacyHandler) SetDisappearingMessages(c *gin.Context) {
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

	var req dto.DisappearingMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	config, err := h.privacyService.SetDisappearingMessages(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("Failed to set disappearing messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "set_disappearing_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetDisappearingMessagesConfig handles GET /api/v1/privacy/disappearing/:conversationId
func (h *PrivacyHandler) GetDisappearingMessagesConfig(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	conversationID, err := uuid.Parse(c.Param("conversationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	config, err := h.privacyService.GetDisappearingMessagesConfig(c.Request.Context(), conversationID)
	if err != nil {
		logger.Error("Failed to get disappearing messages config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_config_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DisableDisappearingMessages handles DELETE /api/v1/privacy/disappearing/:conversationId
func (h *PrivacyHandler) DisableDisappearingMessages(c *gin.Context) {
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

	conversationID, err := uuid.Parse(c.Param("conversationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.privacyService.DisableDisappearingMessages(c.Request.Context(), userID, conversationID); err != nil {
		logger.Error("Failed to disable disappearing messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "disable_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disappearing messages disabled"})
}

// SetupTwoFactorAuth handles POST /api/v1/privacy/2fa/setup
func (h *PrivacyHandler) SetupTwoFactorAuth(c *gin.Context) {
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

	setup, err := h.privacyService.SetupTwoFactorAuth(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to setup 2FA", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "setup_2fa_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, setup)
}

// EnableTwoFactorAuth handles POST /api/v1/privacy/2fa/enable
func (h *PrivacyHandler) EnableTwoFactorAuth(c *gin.Context) {
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

	var req dto.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.privacyService.VerifyAndEnableTwoFactor(c.Request.Context(), userID, &req); err != nil {
		logger.Error("Failed to enable 2FA", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "enable_2fa_failed",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// DisableTwoFactorAuth handles POST /api/v1/privacy/2fa/disable
func (h *PrivacyHandler) DisableTwoFactorAuth(c *gin.Context) {
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

	var req dto.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.privacyService.DisableTwoFactorAuth(c.Request.Context(), userID, &req); err != nil {
		logger.Error("Failed to disable 2FA", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "disable_2fa_failed",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// GetTwoFactorStatus handles GET /api/v1/privacy/2fa/status
func (h *PrivacyHandler) GetTwoFactorStatus(c *gin.Context) {
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

	status, err := h.privacyService.GetTwoFactorStatus(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get 2FA status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_status_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
