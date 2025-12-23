package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/conversation"
	"github.com/yourusername/sotalk/internal/domain/message"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// WSBroadcaster defines the interface for WebSocket broadcasting
type WSBroadcaster interface {
	BroadcastNewMessage(ctx context.Context, conversationID uuid.UUID, msg dto.MessageDTO) error
	BroadcastMessageDeleted(ctx context.Context, conversationID uuid.UUID, messageID string) error
	BroadcastMessageUpdated(ctx context.Context, conversationID uuid.UUID, msg dto.MessageDTO) error
	BroadcastReactionAdded(ctx context.Context, conversationID uuid.UUID, messageID, userID, emoji string) error
	BroadcastReactionRemoved(ctx context.Context, conversationID uuid.UUID, messageID, userID, emoji string) error
	BroadcastMessagePinned(ctx context.Context, conversationID uuid.UUID, messageID, userID string) error
	BroadcastMessageUnpinned(ctx context.Context, conversationID uuid.UUID, messageID, userID string) error
}

// service implements the Service interface
type service struct {
	messageRepo      message.Repository
	conversationRepo conversation.Repository
	userRepo         user.Repository
	wsBroadcaster    WSBroadcaster
}

// NewService creates a new messaging service
func NewService(
	messageRepo message.Repository,
	conversationRepo conversation.Repository,
	userRepo user.Repository,
	wsBroadcaster WSBroadcaster,
) Service {
	return &service{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		wsBroadcaster:    wsBroadcaster,
	}
}

// SendMessage sends a message to a recipient
func (s *service) SendMessage(ctx context.Context, senderID uuid.UUID, req *dto.SendMessageRequest) (*dto.SendMessageResponse, error) {
	// Validate sender exists
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	// Validate recipient exists
	recipient, err := s.userRepo.FindByID(ctx, req.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("recipient not found: %w", err)
	}

	// Get or create conversation
	conversationID, err := s.GetOrCreateDirectConversation(ctx, senderID, req.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create conversation: %w", err)
	}

	// Create message
	msg := message.NewMessage(
		conversationID,
		senderID,
		req.Content,
		message.ContentType(req.ContentType),
	)

	// Set optional fields
	if req.Signature != "" {
		msg.SetSignature(req.Signature)
	}
	if req.ReplyToID != nil {
		msg.SetReplyTo(*req.ReplyToID)
	}

	// Save message
	msg.MarkAsSent()
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation last message
	conv, err := s.conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}
	conv.UpdateLastMessage(msg.ID)
	if err := s.conversationRepo.Update(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	// Map to DTO
	messageDTO := toMessageDTO(msg, sender, recipient)

	// Broadcast new message via WebSocket (async to avoid blocking HTTP response)
	if s.wsBroadcaster != nil {
		logger.Info("🔔 Starting async broadcast",
			zap.String("message_id", msg.ID.String()),
			zap.String("conversation_id", conversationID.String()),
		)
		go func() {
			logger.Info("🔔 Inside goroutine, calling BroadcastNewMessage",
				zap.String("message_id", msg.ID.String()),
			)
			if err := s.wsBroadcaster.BroadcastNewMessage(context.Background(), conversationID, messageDTO); err != nil {
				logger.Error("❌ Failed to broadcast new message",
					zap.String("message_id", msg.ID.String()),
					zap.String("conversation_id", conversationID.String()),
					zap.Error(err),
				)
			} else {
				logger.Info("✅ Broadcast successful",
					zap.String("message_id", msg.ID.String()),
				)
			}
		}()
	} else {
		logger.Warn("⚠️ wsBroadcaster is nil, cannot broadcast")
	}

	return &dto.SendMessageResponse{
		Message: messageDTO,
	}, nil
}

// GetMessages gets messages from a conversation
func (s *service) GetMessages(ctx context.Context, userID uuid.UUID, req *dto.GetMessagesRequest) (*dto.GetMessagesResponse, error) {
	// Check if user is participant
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, req.ConversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant: %w", err)
	}
	if !isParticipant {
		return nil, conversation.ErrNotParticipant
	}

	// Set defaults
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Get messages (isPinned will only be true for messages pinned by this user)
	messages, err := s.messageRepo.FindByConversationID(ctx, req.ConversationID, limit, req.Offset, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Get total count
	total, err := s.messageRepo.CountByConversationID(ctx, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	// Get senders info (cache user lookups)
	userCache := make(map[uuid.UUID]*user.User)
	messageDTOs := make([]dto.MessageDTO, len(messages))

	for i, msg := range messages {
		var sender *user.User
		if cachedUser, exists := userCache[msg.SenderID]; exists {
			sender = cachedUser
		} else {
			sender, err = s.userRepo.FindByID(ctx, msg.SenderID)
			if err == nil {
				userCache[msg.SenderID] = sender
			}
		}

		messageDTOs[i] = toMessageDTO(msg, sender, nil)
	}

	return &dto.GetMessagesResponse{
		Messages: messageDTOs,
		Total:    total,
	}, nil
}

// GetConversations gets user's conversations
func (s *service) GetConversations(ctx context.Context, userID uuid.UUID, req *dto.GetConversationsRequest) (*dto.GetConversationsResponse, error) {
	// Set defaults
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Get conversations
	conversations, err := s.conversationRepo.FindByUserID(ctx, userID, limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Map to DTOs
	conversationDTOs := make([]dto.ConversationDTO, len(conversations))
	for i, conv := range conversations {
		// Get participants
		participants, err := s.conversationRepo.FindParticipants(ctx, conv.ID)
		if err != nil {
			continue
		}

		// Get participant users
		participantDTOs := make([]dto.UserDTO, 0, len(participants))
		for _, p := range participants {
			u, err := s.userRepo.FindByID(ctx, p.UserID)
			if err != nil {
				continue
			}
			participantDTOs = append(participantDTOs, dto.UserDTO{
				ID:            u.ID.String(),
				WalletAddress: u.WalletAddress,
				Username:      u.Username,
				Avatar:        u.Avatar,
				Status:        string(u.Status),
				IsOnline:      p.IsOnline, // Get online status from participant (checked via Hub)
			})
		}

		// Get last message if exists
		var lastMessageDTO *dto.MessageDTO
		if conv.LastMessageID != nil {
			lastMsg, err := s.messageRepo.FindByID(ctx, *conv.LastMessageID)
			if err == nil {
				sender, _ := s.userRepo.FindByID(ctx, lastMsg.SenderID)
				msgDTO := toMessageDTO(lastMsg, sender, nil)
				lastMessageDTO = &msgDTO
			}
		}

		// Count unread messages
		unreadCount, err := s.messageRepo.CountUnreadByConversationID(ctx, conv.ID, userID)
		if err != nil {
			unreadCount = 0 // Default to 0 if error
		}

		conversationDTOs[i] = dto.ConversationDTO{
			ID:           conv.ID.String(),
			Type:         string(conv.Type),
			Participants: participantDTOs,
			LastMessage:  lastMessageDTO,
			UnreadCount:  int(unreadCount),
			CreatedAt:    conv.CreatedAt,
			UpdatedAt:    conv.UpdatedAt,
		}
	}

	return &dto.GetConversationsResponse{
		Conversations: conversationDTOs,
	}, nil
}

// CreateConversation creates a new conversation
func (s *service) CreateConversation(ctx context.Context, userID uuid.UUID, req *dto.CreateConversationRequest) (*dto.CreateConversationResponse, error) {
	// Resolve participant_id (can be UUID or wallet address)
	participantID, err := s.resolveParticipantID(ctx, req.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve participant: %w", err)
	}

	// Check if direct conversation already exists
	if req.Type == "direct" {
		existingConv, err := s.conversationRepo.FindDirectConversation(ctx, userID, participantID)
		if err == nil {
			// Conversation already exists, return it
			participants, _ := s.conversationRepo.FindParticipants(ctx, existingConv.ID)
			participantDTOs := make([]dto.UserDTO, 0, len(participants))
			for _, p := range participants {
				u, err := s.userRepo.FindByID(ctx, p.UserID)
				if err != nil {
					continue
				}
				participantDTOs = append(participantDTOs, dto.UserDTO{
					ID:            u.ID.String(),
					WalletAddress: u.WalletAddress,
					Username:      u.Username,
					Avatar:        u.Avatar,
					Status:        string(u.Status),
					IsOnline:      p.IsOnline, // Get online status from participant (checked via Hub)
				})
			}

			return &dto.CreateConversationResponse{
				Conversation: dto.ConversationDTO{
					ID:           existingConv.ID.String(),
					Type:         string(existingConv.Type),
					Participants: participantDTOs,
					UnreadCount:  0,
					CreatedAt:    existingConv.CreatedAt,
					UpdatedAt:    existingConv.UpdatedAt,
				},
			}, nil
		}
	}

	// Create new conversation
	conv := conversation.NewDirectConversation()
	if err := s.conversationRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Add current user as participant
	participant1 := conversation.NewParticipant(conv.ID, userID, conversation.RoleMember)
	if err := s.conversationRepo.AddParticipant(ctx, participant1); err != nil {
		return nil, fmt.Errorf("failed to add current user: %w", err)
	}

	// Add other participant
	participant2 := conversation.NewParticipant(conv.ID, participantID, conversation.RoleMember)
	if err := s.conversationRepo.AddParticipant(ctx, participant2); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	// Get participants for response
	participants, _ := s.conversationRepo.FindParticipants(ctx, conv.ID)
	participantDTOs := make([]dto.UserDTO, 0, len(participants))
	for _, p := range participants {
		u, err := s.userRepo.FindByID(ctx, p.UserID)
		if err != nil {
			continue
		}
		participantDTOs = append(participantDTOs, dto.UserDTO{
			ID:            u.ID.String(),
			WalletAddress: u.WalletAddress,
			Username:      u.Username,
			Avatar:        u.Avatar,
			Status:        string(u.Status),
			IsOnline:      p.IsOnline, // Get online status from participant (checked via Hub)
		})
	}

	return &dto.CreateConversationResponse{
		Conversation: dto.ConversationDTO{
			ID:           conv.ID.String(),
			Type:         string(conv.Type),
			Participants: participantDTOs,
			UnreadCount:  0,
			CreatedAt:    conv.CreatedAt,
			UpdatedAt:    conv.UpdatedAt,
		},
	}, nil
}

// GetOrCreateDirectConversation gets or creates a direct conversation
func (s *service) GetOrCreateDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (uuid.UUID, error) {
	// Try to find existing conversation
	existingConv, err := s.conversationRepo.FindDirectConversation(ctx, user1ID, user2ID)
	if err == nil {
		return existingConv.ID, nil
	}

	if err != conversation.ErrConversationNotFound {
		return uuid.Nil, fmt.Errorf("failed to find conversation: %w", err)
	}

	// Create new conversation
	conv := conversation.NewDirectConversation()
	if err := s.conversationRepo.Create(ctx, conv); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Add participants
	participant1 := conversation.NewParticipant(conv.ID, user1ID, conversation.RoleMember)
	if err := s.conversationRepo.AddParticipant(ctx, participant1); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add participant 1: %w", err)
	}

	participant2 := conversation.NewParticipant(conv.ID, user2ID, conversation.RoleMember)
	if err := s.conversationRepo.AddParticipant(ctx, participant2); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add participant 2: %w", err)
	}

	return conv.ID, nil
}

// MarkAsRead marks messages in a conversation as read
func (s *service) MarkAsRead(ctx context.Context, userID uuid.UUID, req *dto.MarkAsReadRequest) error {
	// Check if user is participant
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, req.ConversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to check participant: %w", err)
	}
	if !isParticipant {
		return conversation.ErrNotParticipant
	}

	// Get all messages in conversation (isPinned will only be true for messages pinned by this user)
	messages, err := s.messageRepo.FindByConversationID(ctx, req.ConversationID, 1000, 0, &userID)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	// Update status to "read" for all messages not sent by this user
	var messageIDs []string
	for _, msg := range messages {
		// Only mark messages from other users as read
		if msg.SenderID != userID && msg.Status != message.StatusRead {
			if err := s.messageRepo.UpdateStatus(ctx, msg.ID, message.StatusRead); err != nil {
				logger.Error("Failed to update message status", zap.Error(err), zap.String("message_id", msg.ID.String()))
				continue
			}
			messageIDs = append(messageIDs, msg.ID.String())
		}
	}

	// Update last read timestamp
	if err := s.conversationRepo.UpdateParticipantLastRead(ctx, req.ConversationID, userID); err != nil {
		return fmt.Errorf("failed to update last read: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func (s *service) DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error {
	// Get message
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	// Check if user is the sender
	if msg.SenderID != userID {
		return message.ErrUnauthorized
	}

	// Delete message
	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// EditMessage edits the content of a message
func (s *service) EditMessage(ctx context.Context, userID, messageID uuid.UUID, newContent string) (*dto.MessageDTO, error) {
	// Get message
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	// Check if user is the sender
	if msg.SenderID != userID {
		return nil, message.ErrUnauthorized
	}

	// Update message content
	msg.Content = newContent
	msg.UpdatedAt = time.Now()

	// Save updated message
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Get sender for DTO
	sender, err := s.userRepo.FindByID(ctx, msg.SenderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	// Convert to DTO
	messageDTO := toMessageDTO(msg, sender, nil)

	// Broadcast message updated via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastMessageUpdated(
			ctx,
			msg.ConversationID,
			messageDTO,
		); err != nil {
			logger.Error("Failed to broadcast message updated",
				zap.String("message_id", messageID.String()),
				zap.Error(err),
			)
		}
	}

	return &messageDTO, nil
}

// GetMessageByID gets a message by ID
func (s *service) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*dto.MessageDTO, error) {
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	// Get sender
	sender, err := s.userRepo.FindByID(ctx, msg.SenderID)
	if err != nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	messageDTO := toMessageDTO(msg, sender, nil)
	return &messageDTO, nil
}

// GetConversationParticipants gets participant IDs for a conversation
func (s *service) GetConversationParticipants(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	participants, err := s.conversationRepo.FindParticipants(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	participantIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		participantIDs[i] = p.UserID
	}

	return participantIDs, nil
}

// UpdateMessageStatus updates the status of a message (delivered/read)
func (s *service) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) error {
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Update status based on string value
	switch status {
	case "delivered":
		msg.MarkAsDelivered()
	case "read":
		msg.MarkAsRead()
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	// Save updated message
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return err
	}

	return nil
}

// Helper function to map domain message to DTO
func toMessageDTO(msg *message.Message, sender *user.User, recipient *user.User) dto.MessageDTO {
	var replyToID *string
	if msg.ReplyToID != nil {
		id := msg.ReplyToID.String()
		replyToID = &id
	}

	msgDTO := dto.MessageDTO{
		ID:             msg.ID.String(),
		ConversationID: msg.ConversationID.String(),
		SenderID:       msg.SenderID.String(),
		Content:        msg.Content,
		ContentType:    string(msg.ContentType),
		Signature:      msg.Signature,
		ReplyToID:      replyToID,
		Status:         string(msg.Status),
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
		Reactions:      msg.Reactions,
		IsPinned:       msg.IsPinned,
		PinnedBy:       msg.PinnedBy,
	}

	if sender != nil {
		msgDTO.Sender = &dto.UserDTO{
			ID:            sender.ID.String(),
			WalletAddress: sender.WalletAddress,
			Username:      sender.Username,
			Avatar:        sender.Avatar,
			Status:        string(sender.Status),
		}
	}

	return msgDTO
}

// Message Reactions (Day 13)

func (s *service) AddReaction(ctx context.Context, userID, messageID uuid.UUID, emoji string) error {
	// Get message to find conversation ID
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	if err := s.messageRepo.AddReaction(ctx, messageID, userID, emoji); err != nil {
		return err
	}

	// Broadcast reaction added via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastReactionAdded(
			ctx,
			msg.ConversationID,
			messageID.String(),
			userID.String(),
			emoji,
		); err != nil {
			logger.Error("Failed to broadcast reaction added",
				zap.String("message_id", messageID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *service) RemoveReaction(ctx context.Context, userID, messageID uuid.UUID, emoji string) error {
	logger.Info("🗑️ RemoveReaction called",
		zap.String("messageID", messageID.String()),
		zap.String("userID", userID.String()),
		zap.String("emoji", emoji),
	)

	// Get message to find conversation ID
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		logger.Error("Failed to find message", zap.Error(err))
		return err
	}

	logger.Info("📝 Message found, removing reaction from DB",
		zap.String("conversationID", msg.ConversationID.String()),
	)

	if err := s.messageRepo.RemoveReaction(ctx, messageID, userID, emoji); err != nil {
		logger.Error("Failed to remove reaction from DB", zap.Error(err))
		return err
	}

	logger.Info("✅ Reaction removed from DB")

	// Broadcast reaction removed via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastReactionRemoved(
			ctx,
			msg.ConversationID,
			messageID.String(),
			userID.String(),
			emoji,
		); err != nil {
			logger.Error("Failed to broadcast reaction removed",
				zap.String("message_id", messageID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *service) GetMessageReactions(ctx context.Context, messageID uuid.UUID) (*dto.MessageReactionsResponse, error) {
	reactions, err := s.messageRepo.GetMessageReactions(ctx, messageID)
	if err != nil {
		return nil, err
	}

	reactionDTOs := make([]dto.ReactionResponse, len(reactions))
	for i, r := range reactions {
		reactionDTOs[i] = dto.ReactionResponse{
			ID:        r.ID.String(),
			MessageID: r.MessageID.String(),
			UserID:    r.UserID.String(),
			Emoji:     r.Emoji,
			CreatedAt: r.CreatedAt,
		}
	}

	return &dto.MessageReactionsResponse{
		MessageID: messageID.String(),
		Reactions: reactionDTOs,
	}, nil
}

// Pinned Messages (Day 13)

func (s *service) PinMessage(ctx context.Context, userID, conversationID, messageID uuid.UUID) error {
	// Verify user is participant of conversation
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to check participant: %w", err)
	}
	if !isParticipant {
		return conversation.ErrNotParticipant
	}

	// Verify message exists and belongs to conversation
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if msg.ConversationID != conversationID {
		return message.ErrMessageNotFound
	}

	if err := s.messageRepo.PinMessage(ctx, conversationID, messageID, userID); err != nil {
		return fmt.Errorf("failed to pin message: %w", err)
	}

	// Broadcast message pinned via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastMessagePinned(
			ctx,
			conversationID,
			messageID.String(),
			userID.String(),
		); err != nil {
			logger.Error("Failed to broadcast message pinned",
				zap.String("message_id", messageID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *service) UnpinMessage(ctx context.Context, userID, conversationID, messageID uuid.UUID) error {
	// Verify user is participant of conversation
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to check participant: %w", err)
	}
	if !isParticipant {
		return conversation.ErrNotParticipant
	}

	// Verify message exists and belongs to conversation
	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if msg.ConversationID != conversationID {
		return message.ErrMessageNotFound
	}

	if err := s.messageRepo.UnpinMessage(ctx, conversationID, messageID, userID); err != nil {
		return fmt.Errorf("failed to unpin message: %w", err)
	}

	// Broadcast message unpinned via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastMessageUnpinned(
			ctx,
			conversationID,
			messageID.String(),
			userID.String(),
		); err != nil {
			logger.Error("Failed to broadcast message unpinned",
				zap.String("message_id", messageID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *service) GetPinnedMessages(ctx context.Context, conversationID uuid.UUID) (*dto.PinnedMessagesResponse, error) {
	pinned, err := s.messageRepo.GetPinnedMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	pinnedDTOs := make([]dto.PinnedMessageResponse, len(pinned))
	for i, p := range pinned {
		pinnedDTOs[i] = dto.PinnedMessageResponse{
			ID:             p.ID.String(),
			ConversationID: p.ConversationID.String(),
			MessageID:      p.MessageID.String(),
			PinnedBy:       p.PinnedBy.String(),
			PinnedAt:       p.PinnedAt,
		}
	}

	return &dto.PinnedMessagesResponse{
		ConversationID: conversationID.String(),
		PinnedMessages: pinnedDTOs,
	}, nil
}

// Message Forwarding (Day 13)

func (s *service) ForwardMessage(ctx context.Context, userID, messageID, targetConversationID uuid.UUID) (*dto.SendMessageResponse, error) {
	// Get original message
	originalMsg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("original message not found: %w", err)
	}

	// Verify user has access to the original message (is participant in source conversation)
	isSourceParticipant, err := s.conversationRepo.IsParticipant(ctx, originalMsg.ConversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check source participant: %w", err)
	}
	if !isSourceParticipant {
		return nil, conversation.ErrNotParticipant
	}

	// Verify user is participant of target conversation
	isTargetParticipant, err := s.conversationRepo.IsParticipant(ctx, targetConversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check target participant: %w", err)
	}
	if !isTargetParticipant {
		return nil, conversation.ErrNotParticipant
	}

	// Create new message with same content
	newMsg := message.NewMessage(
		targetConversationID,
		userID,
		originalMsg.Content,
		originalMsg.ContentType,
	)

	// Mark as sent
	newMsg.MarkAsSent()

	// Save new message
	if err := s.messageRepo.Create(ctx, newMsg); err != nil {
		return nil, fmt.Errorf("failed to create forwarded message: %w", err)
	}

	// Record forward relationship
	if err := s.messageRepo.RecordForward(ctx, originalMsg.ID, newMsg.ID, userID, targetConversationID); err != nil {
		logger.Warn("Failed to record forward", zap.Error(err))
	}

	// Update conversation's last message
	conv, err := s.conversationRepo.FindByID(ctx, targetConversationID)
	if err == nil {
		conv.UpdateLastMessage(newMsg.ID)
		if err := s.conversationRepo.Update(ctx, conv); err != nil {
			logger.Warn("Failed to update conversation last message", zap.Error(err))
		}
	}

	// Get sender info for response
	sender, _ := s.userRepo.FindByID(ctx, userID)
	messageDTO := toMessageDTO(newMsg, sender, nil)

	// Broadcast forwarded message via WebSocket
	if s.wsBroadcaster != nil {
		go func() {
			if err := s.wsBroadcaster.BroadcastNewMessage(context.Background(), targetConversationID, messageDTO); err != nil {
				logger.Error("Failed to broadcast forwarded message",
					zap.String("message_id", newMsg.ID.String()),
					zap.String("conversation_id", targetConversationID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	logger.Info("Message forwarded",
		zap.String("original_message_id", originalMsg.ID.String()),
		zap.String("new_message_id", newMsg.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("target_conversation_id", targetConversationID.String()),
	)

	return &dto.SendMessageResponse{
		Message: messageDTO,
	}, nil
}

// Message Search (Day 13)

func (s *service) SearchMessages(ctx context.Context, userID uuid.UUID, req *dto.SearchMessagesRequest) (*dto.SearchMessagesResponse, error) {
	var conversationID *uuid.UUID
	if req.ConversationID != nil {
		id, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			return nil, err
		}
		conversationID = &id
	}

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	messages, err := s.messageRepo.SearchMessages(ctx, userID, req.Query, conversationID, limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	messageDTOs := make([]*dto.MessageDTO, len(messages))
	for i, msg := range messages {
		sender, _ := s.userRepo.FindByID(ctx, msg.SenderID)
		msgDTO := toMessageDTO(msg, sender, nil)
		messageDTOs[i] = &msgDTO
	}

	return &dto.SearchMessagesResponse{
		Messages: messageDTOs,
		Total:    len(messageDTOs),
	}, nil
}

// resolveParticipantID resolves a participant identifier to a user UUID
// Accepts either a UUID string or a Solana wallet address
func (s *service) resolveParticipantID(ctx context.Context, participantID string) (uuid.UUID, error) {
	// Try to parse as UUID first
	userUUID, err := uuid.Parse(participantID)
	if err == nil {
		// It's a valid UUID, verify user exists
		_, err := s.userRepo.FindByID(ctx, userUUID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("user not found with ID: %s", participantID)
		}
		return userUUID, nil
	}

	// Not a UUID, treat as wallet address
	user, err := s.userRepo.FindByWalletAddress(ctx, participantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user not found with wallet address: %s", participantID)
	}

	return user.ID, nil
}

// ArchiveConversation archives a conversation for a user
func (s *service) ArchiveConversation(ctx context.Context, userID, conversationID uuid.UUID) error {
	// Verify user is a participant
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		logger.Error("Failed to check participant status", zap.Error(err))
		return fmt.Errorf("failed to check participant status: %w", err)
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant of this conversation")
	}

	// Archive the conversation
	if err := s.conversationRepo.ArchiveConversation(ctx, conversationID, userID); err != nil {
		logger.Error("Failed to archive conversation", zap.Error(err))
		return fmt.Errorf("failed to archive conversation: %w", err)
	}

	logger.Info("Conversation archived",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()))

	return nil
}

// UnarchiveConversation unarchives a conversation for a user
func (s *service) UnarchiveConversation(ctx context.Context, userID, conversationID uuid.UUID) error {
	// Verify user is a participant
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		logger.Error("Failed to check participant status", zap.Error(err))
		return fmt.Errorf("failed to check participant status: %w", err)
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant of this conversation")
	}

	// Unarchive the conversation
	if err := s.conversationRepo.UnarchiveConversation(ctx, conversationID, userID); err != nil {
		logger.Error("Failed to unarchive conversation", zap.Error(err))
		return fmt.Errorf("failed to unarchive conversation: %w", err)
	}

	logger.Info("Conversation unarchived",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()))

	return nil
}

// DeleteConversation permanently deletes a conversation (only if user is owner/admin)
func (s *service) DeleteConversation(ctx context.Context, userID, conversationID uuid.UUID) error {
	// Get conversation details
	conv, err := s.conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		logger.Error("Failed to find conversation", zap.Error(err))
		return fmt.Errorf("conversation not found: %w", err)
	}

	// Verify user is a participant
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		logger.Error("Failed to check participant status", zap.Error(err))
		return fmt.Errorf("failed to check participant status: %w", err)
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant of this conversation")
	}

	// For direct conversations, allow any participant to delete
	// For group/channel conversations, only allow owner/admin to delete
	if conv.Type != conversation.TypeDirect {
		participants, err := s.conversationRepo.FindParticipants(ctx, conversationID)
		if err != nil {
			logger.Error("Failed to find participants", zap.Error(err))
			return fmt.Errorf("failed to check permissions: %w", err)
		}

		// Check if user is owner or admin
		hasPermission := false
		for _, p := range participants {
			if p.UserID == userID && (p.Role == conversation.RoleOwner || p.Role == conversation.RoleAdmin) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return fmt.Errorf("only owner or admin can delete group/channel conversations")
		}
	}

	// Delete the conversation
	if err := s.conversationRepo.Delete(ctx, conversationID); err != nil {
		logger.Error("Failed to delete conversation", zap.Error(err))
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	logger.Info("Conversation deleted",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()))

	return nil
}
