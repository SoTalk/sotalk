package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/domain/contact"
	"github.com/yourusername/sotalk/internal/domain/privacy"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/message"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// MessageHandler handles message HTTP requests
type MessageHandler struct {
	messageService message.Service
	privacyRepo    privacy.Repository
	contactRepo    contact.Repository
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService message.Service, privacyRepo privacy.Repository, contactRepo contact.Repository) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		privacyRepo:    privacyRepo,
		contactRepo:    contactRepo,
	}
}

// SendMessage handles POST /api/v1/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	senderID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse recipient ID
	recipientID, err := req.ParseRecipientID()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_recipient_id",
			Message: "Invalid recipient ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse reply to ID
	replyToID, err := req.ParseReplyToID()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_reply_to_id",
			Message: "Invalid reply to ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.messageService.SendMessage(c.Request.Context(), senderID, &dto.SendMessageRequest{
		RecipientID: recipientID,
		Content:     req.Content, // Plain text, no encryption
		ContentType: req.ContentType,
		Signature:   req.Signature,
		ReplyToID:   replyToID,
	})

	if err != nil {
		logger.Error("Failed to send message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "send_message_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}


	c.JSON(http.StatusOK, response.SendMessageResponse{
		Message: mapMessageDTO(result.Message),
	})
}

// GetMessages handles GET /api/v1/messages
func (h *MessageHandler) GetMessages(c *gin.Context) {
	// Get user ID from context
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

	var req request.GetMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse conversation ID
	conversationID, err := req.ParseConversationID()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.messageService.GetMessages(c.Request.Context(), userID, &dto.GetMessagesRequest{
		ConversationID: conversationID,
		Limit:          req.Limit,
		Offset:         req.Offset,
	})

	if err != nil {
		logger.Error("Failed to get messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_messages_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	messages := make([]response.MessageDTO, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = mapMessageDTO(msg)
	}

	c.JSON(http.StatusOK, response.GetMessagesResponse{
		Messages: messages,
		Total:    result.Total,
	})
}

// GetConversations handles GET /api/v1/conversations
func (h *MessageHandler) GetConversations(c *gin.Context) {
	// Get user ID from context
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

	var req request.GetConversationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.messageService.GetConversations(c.Request.Context(), userID, &dto.GetConversationsRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	})

	if err != nil {
		logger.Error("Failed to get conversations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_conversations_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	conversations := make([]response.ConversationDTO, len(result.Conversations))
	for i, conv := range result.Conversations {
		conversations[i] = h.mapConversationDTOWithPrivacy(c.Request.Context(), conv, userID)
	}

	c.JSON(http.StatusOK, response.GetConversationsResponse{
		Conversations: conversations,
	})
}

// CreateConversation handles POST /api/v1/conversations
func (h *MessageHandler) CreateConversation(c *gin.Context) {
	// Get user ID from context
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

	var req request.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate participant_id is not empty
	if req.ParticipantID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_participant_id",
			Message: "Participant ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Default to direct conversation
	conversationType := "direct"
	if req.Type != nil && *req.Type != "" {
		conversationType = *req.Type
	}

	// Call use case (service will handle UUID or wallet address)
	result, err := h.messageService.CreateConversation(c.Request.Context(), userID, &dto.CreateConversationRequest{
		ParticipantID: req.ParticipantID, // Pass as string (UUID or wallet address)
		Type:          conversationType,
	})

	if err != nil {
		logger.Error("Failed to create conversation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_conversation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.ConversationResponse{
		Conversation: h.mapConversationDTOWithPrivacy(c.Request.Context(), result.Conversation, userID),
	})
}

// MarkAsRead handles POST /api/v1/messages/read
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	// Get user ID from context
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

	var req request.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse conversation ID
	conversationID, err := req.ParseConversationID()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	err = h.messageService.MarkAsRead(c.Request.Context(), userID, &dto.MarkAsReadRequest{
		ConversationID: conversationID,
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

	// NOTE: WebSocket message.read events are handled by the WebSocket client handler
	// when the client explicitly sends message IDs. The HTTP endpoint doesn't broadcast
	// because it doesn't know which specific messages were marked as read.

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

// DeleteMessage handles DELETE /api/v1/messages/:id
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	// Get user ID from context
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

	// Get message ID from URL
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	err = h.messageService.DeleteMessage(c.Request.Context(), userID, messageID)
	if err != nil {
		logger.Error("Failed to delete message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_message_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}

// EditMessage handles PUT /api/v1/messages/:id
func (h *MessageHandler) EditMessage(c *gin.Context) {
	// Get user ID from context
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

	// Get message ID from URL
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse request body
	var req request.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.messageService.EditMessage(c.Request.Context(), userID, messageID, req.Content)
	if err != nil {
		logger.Error("Failed to edit message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "edit_message_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SendMessageResponse{
		Message: mapMessageDTO(*result),
	})
}

// Helper functions to map DTOs
func mapMessageDTO(msg dto.MessageDTO) response.MessageDTO {
	var sender *response.UserDTO
	if msg.Sender != nil {
		sender = &response.UserDTO{
			ID:            msg.Sender.ID,
			WalletAddress: msg.Sender.WalletAddress,
			Username:      msg.Sender.Username,
			Avatar:        msg.Sender.Avatar,
			Status:        msg.Sender.Status,
		}
	}

	return response.MessageDTO{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		ContentType:    msg.ContentType,
		Signature:      msg.Signature,
		ReplyToID:      msg.ReplyToID,
		Status:         msg.Status,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
		Sender:         sender,
		Reactions:      msg.Reactions,
		IsPinned:       msg.IsPinned,
		PinnedBy:       msg.PinnedBy,
	}
}

// mapConversationDTOWithPrivacy maps conversation DTO with privacy checking
func (h *MessageHandler) mapConversationDTOWithPrivacy(ctx context.Context, conv dto.ConversationDTO, currentUserID uuid.UUID) response.ConversationDTO {
	participants := make([]response.UserDTO, len(conv.Participants))
	for i, p := range conv.Participants {
		// Default to hiding everything
		canSeeAvatar := false
		canSeeLastSeen := false
		canSeeStatus := false

		// Parse target user ID
		targetUserID, err := uuid.Parse(p.ID)
		if err == nil && targetUserID != currentUserID {
			// Get target user's privacy settings
			privacySettings, err := h.privacyRepo.GetPrivacySettings(ctx, targetUserID)
			if err == nil && privacySettings != nil {
				// Determine relationship between current user and target user
				relationship := "stranger"
				isContact, err := h.contactRepo.IsContact(ctx, currentUserID, targetUserID)
				if err == nil && isContact {
					relationship = "contact"
				}

				// Calculate privacy flags based on target user's settings
				canSeeAvatar = privacySettings.CanSeeProfilePhoto(relationship)
				canSeeLastSeen = privacySettings.CanSeeLastSeen(relationship)
				canSeeStatus = privacySettings.CanSeeStatus(relationship)
			} else {
				// If no privacy settings found, default to everyone can see (for backwards compatibility)
				canSeeAvatar = true
				canSeeLastSeen = true
				canSeeStatus = true
			}
		} else if targetUserID == currentUserID {
			// User can always see their own information
			canSeeAvatar = true
			canSeeLastSeen = true
			canSeeStatus = true
		}

		participants[i] = response.UserDTO{
			ID:             p.ID,
			WalletAddress:  p.WalletAddress,
			Username:       p.Username,
			Avatar:         p.Avatar,
			Status:         p.Status,
			IsOnline:       p.IsOnline,
			LastSeen:       &p.LastSeen,
			CanSeeAvatar:   &canSeeAvatar,
			CanSeeLastSeen: &canSeeLastSeen,
			CanSeeStatus:   &canSeeStatus,
		}
	}

	var lastMessage *response.MessageDTO
	if conv.LastMessage != nil {
		msg := mapMessageDTO(*conv.LastMessage)
		lastMessage = &msg
	}

	return response.ConversationDTO{
		ID:           conv.ID,
		Type:         conv.Type,
		Participants: participants,
		LastMessage:  lastMessage,
		UnreadCount:  conv.UnreadCount,
		CreatedAt:    conv.CreatedAt,
		UpdatedAt:    conv.UpdatedAt,
	}
}

// AddReaction handles POST /api/v1/messages/:id/react (Day 13)
func (h *MessageHandler) AddReaction(c *gin.Context) {
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

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req struct {
		Emoji string `json:"emoji" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.messageService.AddReaction(c.Request.Context(), userID, messageID, req.Emoji); err != nil {
		logger.Error("Failed to add reaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "add_reaction_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// WebSocket broadcast is handled by the service layer
	// Removed duplicate broadcast to prevent double events

	c.JSON(http.StatusOK, gin.H{"message": "reaction added"})
}

// RemoveReaction handles DELETE /api/v1/messages/:id/react (Day 13)
func (h *MessageHandler) RemoveReaction(c *gin.Context) {
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

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	emoji := c.Query("emoji")
	if emoji == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "missing_emoji",
			Message: "Emoji parameter is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.messageService.RemoveReaction(c.Request.Context(), userID, messageID, emoji); err != nil {
		logger.Error("Failed to remove reaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "remove_reaction_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	logger.Info("Reaction removed successfully",
		zap.String("message_id", messageID.String()),
		zap.String("emoji", emoji),
		zap.String("user_id", userID.String()),
	)

	// WebSocket broadcast is handled by the service layer
	// Removed duplicate broadcast to prevent double events

	c.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
}

// GetMessageReactions handles GET /api/v1/messages/:id/reactions (Day 13)
func (h *MessageHandler) GetMessageReactions(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	reactions, err := h.messageService.GetMessageReactions(c.Request.Context(), messageID)
	if err != nil {
		logger.Error("Failed to get reactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_reactions_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, reactions)
}

// PinMessage handles POST /api/v1/messages/:id/pin (Day 13)
func (h *MessageHandler) PinMessage(c *gin.Context) {
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

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.messageService.PinMessage(c.Request.Context(), userID, conversationID, messageID); err != nil {
		logger.Error("Failed to pin message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "pin_message_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message pinned"})
}

// UnpinMessage handles DELETE /api/v1/messages/:id/pin (Day 13)
func (h *MessageHandler) UnpinMessage(c *gin.Context) {
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

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	conversationID, err := uuid.Parse(c.Query("conversation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "conversation_id query parameter required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.messageService.UnpinMessage(c.Request.Context(), userID, conversationID, messageID); err != nil {
		logger.Error("Failed to unpin message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "unpin_message_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message unpinned"})
}

// GetPinnedMessages handles GET /api/v1/conversations/:id/pinned (Day 13)
func (h *MessageHandler) GetPinnedMessages(c *gin.Context) {
	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	pinnedMessages, err := h.messageService.GetPinnedMessages(c.Request.Context(), conversationID)
	if err != nil {
		logger.Error("Failed to get pinned messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_pinned_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, pinnedMessages)
}

// ForwardMessage handles POST /api/v1/messages/:id/forward (Day 13)
func (h *MessageHandler) ForwardMessage(c *gin.Context) {
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

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_message_id",
			Message: "Invalid message ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.ForwardMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	targetConversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_target_conversation_id",
			Message: "Invalid target conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.messageService.ForwardMessage(c.Request.Context(), userID, messageID, targetConversationID)
	if err != nil {
		logger.Error("Failed to forward message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "forward_message_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SearchMessages handles POST /api/v1/messages/search (Day 13)
func (h *MessageHandler) SearchMessages(c *gin.Context) {
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

	var req dto.SearchMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.messageService.SearchMessages(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("Failed to search messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "search_messages_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ArchiveConversation handles PUT /api/v1/conversations/:id/archive
func (h *MessageHandler) ArchiveConversation(c *gin.Context) {
	// Get user ID from context
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

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Archive the conversation
	if err := h.messageService.ArchiveConversation(c.Request.Context(), userID, conversationID); err != nil {
		logger.Error("Failed to archive conversation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "archive_conversation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversation archived successfully",
	})
}

// UnarchiveConversation handles PUT /api/v1/conversations/:id/unarchive
func (h *MessageHandler) UnarchiveConversation(c *gin.Context) {
	// Get user ID from context
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

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Unarchive the conversation
	if err := h.messageService.UnarchiveConversation(c.Request.Context(), userID, conversationID); err != nil {
		logger.Error("Failed to unarchive conversation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "unarchive_conversation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversation unarchived successfully",
	})
}

// DeleteConversation handles DELETE /api/v1/conversations/:id
func (h *MessageHandler) DeleteConversation(c *gin.Context) {
	// Get user ID from context
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

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "Invalid conversation ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Delete the conversation
	if err := h.messageService.DeleteConversation(c.Request.Context(), userID, conversationID); err != nil {
		logger.Error("Failed to delete conversation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_conversation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversation deleted successfully",
	})
}
