package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 15 * time.Second // Fast offline detection (force-kill scenarios)

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10 // 13.5 seconds

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a WebSocket client connection
type Client struct {
	// Unique client ID (for multiple devices)
	ID string

	// User ID this client belongs to
	UserID uuid.UUID

	// Username for display purposes
	Username string

	// The websocket connection
	conn *websocket.Conn

	// Hub reference
	hub *Hub

	// Buffered channel of outbound messages
	send chan []byte

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Client message handler
	messageHandler ClientMessageHandler
}

// ClientMessageHandler handles incoming client messages
type ClientMessageHandler interface {
	HandleTyping(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID)
	HandleStopTyping(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID)
	HandleMessageDelivered(ctx context.Context, userID uuid.UUID, messageID uuid.UUID, conversationID uuid.UUID)
	HandleMessageRead(ctx context.Context, userID uuid.UUID, messageID uuid.UUID, conversationID uuid.UUID)
}

// NewClient creates a new Client instance
func NewClient(userID uuid.UUID, username string, conn *websocket.Conn, hub *Hub, messageHandler ClientMessageHandler) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		ID:             uuid.New().String(),
		UserID:         userID,
		Username:       username,
		conn:           conn,
		hub:            hub,
		send:           make(chan []byte, 256),
		ctx:            ctx,
		cancel:         cancel,
		messageHandler: messageHandler,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		c.cancel()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Accept normal closure (1000), going away (1001), and abnormal closure
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error",
					zap.String("user_id", c.UserID.String()),
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
			} else {
				logger.Info("🔌 WebSocket closed gracefully",
					zap.String("user_id", c.UserID.String()),
					zap.String("client_id", c.ID),
				)
			}
			break
		}

		// Handle client message
		c.handleClientMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// handleClientMessage processes incoming messages from the client
func (c *Client) handleClientMessage(message []byte) {
	var clientMsg ClientMessage
	if err := json.Unmarshal(message, &clientMsg); err != nil {
		logger.Error("Failed to unmarshal client message",
			zap.String("user_id", c.UserID.String()),
			zap.Error(err),
		)
		c.sendError("invalid_message", "Failed to parse message")
		return
	}

	// Process the message
	switch clientMsg.Type {
	case ClientMessageTyping:
		c.handleTyping(clientMsg.Payload)

	case ClientMessageStopTyping:
		c.handleStopTyping(clientMsg.Payload)

	case ClientMessageDelivered:
		c.handleMessageDelivered(clientMsg.Payload)

	case ClientMessageRead:
		c.handleMessageRead(clientMsg.Payload)

	case ClientMessagePing:
		c.handlePing()

	default:
		logger.Warn("Unknown client message type",
			zap.String("type", clientMsg.Type),
			zap.String("user_id", c.UserID.String()),
		)
	}
}

// handleTyping handles typing indicator from client
func (c *Client) handleTyping(payload json.RawMessage) {
	var typingPayload ClientTypingPayload
	if err := json.Unmarshal(payload, &typingPayload); err != nil {
		logger.Error("Failed to unmarshal typing payload", zap.Error(err))
		return
	}

	conversationID, err := uuid.Parse(typingPayload.ConversationID)
	if err != nil {
		logger.Error("Invalid conversation ID", zap.Error(err))
		return
	}

	if c.messageHandler != nil {
		c.messageHandler.HandleTyping(c.ctx, c.UserID, conversationID)
	}
}

// handleStopTyping handles stop typing indicator from client
func (c *Client) handleStopTyping(payload json.RawMessage) {
	var typingPayload ClientTypingPayload
	if err := json.Unmarshal(payload, &typingPayload); err != nil {
		logger.Error("Failed to unmarshal stop typing payload", zap.Error(err))
		return
	}

	conversationID, err := uuid.Parse(typingPayload.ConversationID)
	if err != nil {
		logger.Error("Invalid conversation ID", zap.Error(err))
		return
	}

	if c.messageHandler != nil {
		c.messageHandler.HandleStopTyping(c.ctx, c.UserID, conversationID)
	}
}

// handleMessageDelivered handles message delivery acknowledgment
func (c *Client) handleMessageDelivered(payload json.RawMessage) {
	var statusPayload ClientMessageStatusPayload
	if err := json.Unmarshal(payload, &statusPayload); err != nil {
		logger.Error("Failed to unmarshal delivered payload", zap.Error(err))
		return
	}

	messageID, err := uuid.Parse(statusPayload.MessageID)
	if err != nil {
		logger.Error("Invalid message ID", zap.Error(err))
		return
	}

	conversationID, err := uuid.Parse(statusPayload.ConversationID)
	if err != nil {
		logger.Error("Invalid conversation ID", zap.Error(err))
		return
	}

	if c.messageHandler != nil {
		c.messageHandler.HandleMessageDelivered(c.ctx, c.UserID, messageID, conversationID)
	}
}

// handleMessageRead handles message read acknowledgment
func (c *Client) handleMessageRead(payload json.RawMessage) {
	var statusPayload ClientMessageStatusPayload
	if err := json.Unmarshal(payload, &statusPayload); err != nil {
		logger.Error("Failed to unmarshal read payload", zap.Error(err))
		return
	}

	messageID, err := uuid.Parse(statusPayload.MessageID)
	if err != nil {
		logger.Error("Invalid message ID", zap.Error(err))
		return
	}

	conversationID, err := uuid.Parse(statusPayload.ConversationID)
	if err != nil {
		logger.Error("Invalid conversation ID", zap.Error(err))
		return
	}

	if c.messageHandler != nil {
		c.messageHandler.HandleMessageRead(c.ctx, c.UserID, messageID, conversationID)
	}
}

// handlePing responds with pong
func (c *Client) handlePing() {
	event, err := NewEvent(EventPong, map[string]interface{}{
		"timestamp": time.Now(),
	})
	if err != nil {
		logger.Error("Failed to create pong event", zap.Error(err))
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal pong event", zap.Error(err))
		return
	}

	select {
	case c.send <- data:
	default:
		logger.Warn("Failed to send pong, buffer full")
	}
}

// sendError sends an error event to the client
func (c *Client) sendError(code, message string) {
	event, err := NewEvent(EventError, ErrorPayload{
		Code:    code,
		Message: message,
	})
	if err != nil {
		logger.Error("Failed to create error event", zap.Error(err))
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal error event", zap.Error(err))
		return
	}

	select {
	case c.send <- data:
	default:
		logger.Warn("Failed to send error, buffer full")
	}
}

