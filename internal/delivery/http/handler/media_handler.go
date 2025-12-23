package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/media"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// MediaHandler handles media HTTP requests
type MediaHandler struct {
	mediaService media.Service
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(mediaService media.Service) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// UploadMedia handles POST /api/v1/media/upload
func (h *MediaHandler) UploadMedia(c *gin.Context) {
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

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to get file", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_file",
			Message: "No file provided",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Upload media
	result, err := h.mediaService.UploadMedia(c.Request.Context(), userID, file)
	if err != nil {
		logger.Error("Failed to upload media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, response.UploadMediaResponse{
		Media: mapMediaDTO(result.Media),
	})
}

// GetMedia handles GET /api/v1/media/:id
func (h *MediaHandler) GetMedia(c *gin.Context) {
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

	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_media_id",
			Message: "Invalid media ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.mediaService.GetMedia(c.Request.Context(), userID, mediaID)
	if err != nil {
		logger.Error("Failed to get media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_media_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.GetMediaResponse{
		Media: mapMediaDTO(result.Media),
	})
}

// GetUserMedia handles GET /api/v1/media
func (h *MediaHandler) GetUserMedia(c *gin.Context) {
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

	var req request.GetMediaListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.mediaService.GetUserMedia(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get user media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_media_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	mediaList := make([]response.MediaDTO, len(result.Media))
	for i, m := range result.Media {
		mediaList[i] = mapMediaDTO(m)
	}

	c.JSON(http.StatusOK, response.GetMediaListResponse{
		Media: mediaList,
		Total: result.Total,
	})
}

// DeleteMedia handles DELETE /api/v1/media/:id
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
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

	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_media_id",
			Message: "Invalid media ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.mediaService.DeleteMedia(c.Request.Context(), userID, mediaID)
	if err != nil {
		logger.Error("Failed to delete media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_media_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "media deleted"})
}

// UploadVoice handles POST /api/v1/media/voice
func (h *MediaHandler) UploadVoice(c *gin.Context) {
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

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to get voice file", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_file",
			Message: "No voice file provided",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Upload voice message
	result, err := h.mediaService.UploadMedia(c.Request.Context(), userID, file)
	if err != nil {
		logger.Error("Failed to upload voice", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, response.UploadMediaResponse{
		Media: mapMediaDTO(result.Media),
	})
}

// UploadFile handles POST /api/v1/media/file
func (h *MediaHandler) UploadFile(c *gin.Context) {
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

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to get file", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_file",
			Message: "No file provided",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Upload file
	result, err := h.mediaService.UploadMedia(c.Request.Context(), userID, file)
	if err != nil {
		logger.Error("Failed to upload file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, response.UploadMediaResponse{
		Media: mapMediaDTO(result.Media),
	})
}

// GetConversationMedia handles GET /api/v1/conversations/:id/media
func (h *MediaHandler) GetConversationMedia(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	_, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.GetMediaListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// In a full implementation, this would query media from the conversation
	// For now, log the conversation ID and return success with empty data
	logger.Debug("Getting conversation media",
		zap.String("conversation_id", conversationID.String()),
		zap.Int("limit", req.Limit),
		zap.Int("offset", req.Offset),
	)

	c.JSON(http.StatusOK, response.GetMediaListResponse{
		Media: []response.MediaDTO{},
		Total: 0,
	})
}

// GetStorageInfo handles GET /api/v1/media/storage
func (h *MediaHandler) GetStorageInfo(c *gin.Context) {
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

	// Get user storage size
	totalSize, err := h.mediaService.GetUserStorageSize(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get storage info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_storage_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Calculate storage in MB and GB
	totalSizeMB := float64(totalSize) / (1024 * 1024)
	totalSizeGB := float64(totalSize) / (1024 * 1024 * 1024)
	maxStorageGB := float64(1) // 1GB max
	percentUsed := (totalSizeGB / maxStorageGB) * 100

	c.JSON(http.StatusOK, gin.H{
		"total_bytes":   totalSize,
		"total_mb":      totalSizeMB,
		"total_gb":      totalSizeGB,
		"max_gb":        maxStorageGB,
		"percent_used":  percentUsed,
		"available_gb":  maxStorageGB - totalSizeGB,
	})
}

// Helper function to map media DTO
func mapMediaDTO(m dto.MediaDTO) response.MediaDTO {
	var metadata *response.MediaMetadata
	if m.Metadata != nil {
		metadata = &response.MediaMetadata{
			Width:      m.Metadata.Width,
			Height:     m.Metadata.Height,
			Duration:   m.Metadata.Duration,
			Waveform:   m.Metadata.Waveform,
			Format:     m.Metadata.Format,
			Resolution: m.Metadata.Resolution,
		}
	}

	return response.MediaDTO{
		ID:           m.ID,
		UserID:       m.UserID,
		MessageID:    m.MessageID,
		Type:         m.Type,
		FileName:     m.FileName,
		FileSize:     m.FileSize,
		FileSizeMB:   m.FileSizeMB,
		MimeType:     m.MimeType,
		URL:          m.URL,
		ThumbnailURL: m.ThumbnailURL,
		Metadata:     metadata,
		CreatedAt:    m.CreatedAt,
	}
}
