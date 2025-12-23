package websocket

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yourusername/sotalk/internal/usecase/message"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// getAllowedOrigins returns list of allowed origins for WebSocket connections
func getAllowedOrigins() []string {
	// Get from environment variable or use defaults
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv != "" {
		return strings.Split(allowedOriginsEnv, ",")
	}

	// Default allowed origins
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		// Production: only allow specific domains
		return []string{
			"https://sotalk.com",
			"https://www.sotalk.com",
			"https://app.sotalk.com",
		}
	}

	// Development: allow localhost and common dev ports
	return []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"http://localhost:5173", // Vite default
		"http://127.0.0.1:3000",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:5173",
	}
}

// checkOrigin validates the WebSocket upgrade request origin
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow requests with no origin (e.g., mobile apps, Postman)
		logger.Debug("WebSocket connection with no origin header")
		return true
	}

	allowedOrigins := getAllowedOrigins()
	for _, allowed := range allowedOrigins {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}

	logger.Warn("WebSocket connection rejected: invalid origin",
		zap.String("origin", origin),
		zap.Strings("allowed_origins", allowedOrigins),
	)
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// Handler handles WebSocket connections
type Handler struct {
	hub            *Hub
	messageService message.Service
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, messageService message.Service) *Handler {
	return &Handler{
		hub:            hub,
		messageService: messageService,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get user ID from auth middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get username (optional, for logging)
	username, _ := c.Get("username")
	usernameStr := ""
	if username != nil {
		usernameStr = username.(string)
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return
	}

	// Create message handler for this client
	messageHandler := NewMessageHandler(h.hub, h.messageService)

	// Create client
	client := NewClient(userID, usernameStr, conn, h.hub, messageHandler)

	// Register client with hub
	h.hub.register <- client

	logger.Info("WebSocket connection established",
		zap.String("user_id", userID.String()),
		zap.String("client_id", client.ID),
		zap.String("username", usernameStr),
	)

	// Start client pumps in goroutines
	go client.WritePump()
	go client.ReadPump()
}

// MessageHandler implements ClientMessageHandler interface
type MessageHandler struct {
	hub            *Hub
	messageService message.Service
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(hub *Hub, messageService message.Service) *MessageHandler {
	return &MessageHandler{
		hub:            hub,
		messageService: messageService,
	}
}

// HandleTyping broadcasts typing indicator to conversation participants
func (m *MessageHandler) HandleTyping(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) {
	event, err := NewEvent(EventTypingStart, TypingPayload{
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
		IsTyping:       true,
	})
	if err != nil {
		logger.Error("Failed to create typing event", zap.Error(err))
		return
	}

	// Broadcast to conversation (excluding self is optional)
	if err := m.hub.BroadcastToConversation(ctx, conversationID, event); err != nil {
		logger.Error("Failed to broadcast typing event",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err),
		)
	}

	logger.Debug("Typing indicator sent",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()),
	)
}

// HandleStopTyping broadcasts stop typing indicator to conversation participants
func (m *MessageHandler) HandleStopTyping(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) {
	event, err := NewEvent(EventTypingStop, TypingPayload{
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
		IsTyping:       false,
	})
	if err != nil {
		logger.Error("Failed to create stop typing event", zap.Error(err))
		return
	}

	// Broadcast to conversation
	if err := m.hub.BroadcastToConversation(ctx, conversationID, event); err != nil {
		logger.Error("Failed to broadcast stop typing event",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err),
		)
	}

	logger.Debug("Stop typing indicator sent",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()),
	)
}

// HandleMessageDelivered marks message as delivered and broadcasts to sender
func (m *MessageHandler) HandleMessageDelivered(ctx context.Context, userID uuid.UUID, messageID uuid.UUID, conversationID uuid.UUID) {
	// Update message status in database
	if err := m.messageService.UpdateMessageStatus(ctx, messageID, "delivered"); err != nil {
		logger.Error("Failed to update message status to delivered",
			zap.String("message_id", messageID.String()),
			zap.Error(err),
		)
		return
	}

	// Broadcast delivery receipt to conversation (sender will receive it)
	event, err := NewEvent(EventMessageDelivered, MessageStatusPayload{
		MessageID:      messageID.String(),
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
		Status:         "delivered",
		Timestamp:      time.Now(),
	})
	if err != nil {
		logger.Error("Failed to create delivered event", zap.Error(err))
		return
	}

	if err := m.hub.BroadcastToConversation(ctx, conversationID, event); err != nil {
		logger.Error("Failed to broadcast delivered event",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err),
		)
	}

	logger.Debug("Message marked as delivered",
		zap.String("message_id", messageID.String()),
		zap.String("user_id", userID.String()),
	)
}

// HandleMessageRead marks message as read and broadcasts to sender
func (m *MessageHandler) HandleMessageRead(ctx context.Context, userID uuid.UUID, messageID uuid.UUID, conversationID uuid.UUID) {
	// Update message status in database
	if err := m.messageService.UpdateMessageStatus(ctx, messageID, "read"); err != nil {
		logger.Error("Failed to update message status to read",
			zap.String("message_id", messageID.String()),
			zap.Error(err),
		)
		return
	}

	// Broadcast read receipt to conversation (sender will receive it)
	event, err := NewEvent(EventMessageRead, MessageStatusPayload{
		MessageID:      messageID.String(),
		ConversationID: conversationID.String(),
		UserID:         userID.String(),
		Status:         "read",
		Timestamp:      time.Now(),
	})
	if err != nil {
		logger.Error("Failed to create read event", zap.Error(err))
		return
	}

	if err := m.hub.BroadcastToConversation(ctx, conversationID, event); err != nil {
		logger.Error("Failed to broadcast read event",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err),
		)
	}

	logger.Debug("Message marked as read",
		zap.String("message_id", messageID.String()),
		zap.String("user_id", userID.String()),
	)
}

// Stats returns WebSocket statistics
func (h *Handler) Stats(c *gin.Context) {
	stats := h.hub.Stats()
	c.JSON(http.StatusOK, stats)
}
