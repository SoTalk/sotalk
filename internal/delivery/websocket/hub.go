package websocket

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/conversation"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// Hub maintains active WebSocket connections and handles broadcasting
type Hub struct {
	// Registered clients by user ID
	// Map[UserID]Map[ClientID]*Client - supports multiple devices per user
	clients map[uuid.UUID]map[string]*Client

	// Conversation repository for participant lookups
	conversationRepo conversation.Repository

	// User repository for updating online status and last seen
	userRepo user.Repository

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub(conversationRepo conversation.Repository, userRepo user.Repository) *Hub {
	return &Hub{
		clients:          make(map[uuid.UUID]map[string]*Client),
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		register:         make(chan *Client, 256),
		unregister:       make(chan *Client, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	// Initialize user's client map if not exists
	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[string]*Client)
	}

	isFirstDevice := len(h.clients[client.UserID]) == 0

	// Add client with unique ID
	h.clients[client.UserID][client.ID] = client

	logger.Info("🟢 Client connected",
		zap.String("user_id", client.UserID.String()),
		zap.String("client_id", client.ID),
		zap.Int("total_devices", len(h.clients[client.UserID])),
		zap.Bool("is_first_device", isFirstDevice),
	)

	// Broadcast user online status IMMEDIATELY if this is their first device
	if isFirstDevice {
		logger.Info("📢 Broadcasting user ONLINE immediately",
			zap.String("user_id", client.UserID.String()),
		)
		h.broadcastUserStatusSync(client.UserID, true)

		// Update user status to online in database (asynchronous to avoid blocking)
		go func(userID uuid.UUID) {
			ctx := context.Background()
			userEntity, err := h.userRepo.FindByID(ctx, userID)
			if err != nil {
				logger.Error("❌ Failed to fetch user for status update",
					zap.String("user_id", userID.String()),
					zap.Error(err),
				)
				return
			}

			userEntity.UpdateStatus(user.StatusOnline)
			if err := h.userRepo.Update(ctx, userEntity); err != nil {
				logger.Error("❌ Failed to update user online status in database",
					zap.String("user_id", userID.String()),
					zap.Error(err),
				)
			} else {
				logger.Info("✅ Updated user online status in database",
					zap.String("user_id", userID.String()),
				)
			}
		}(client.UserID)
	}

	h.mu.Unlock()
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if clients, ok := h.clients[client.UserID]; ok {
		if _, exists := clients[client.ID]; exists {
			delete(clients, client.ID)
			close(client.send)

			remainingDevices := len(clients)
			isLastDevice := remainingDevices == 0

			logger.Info("🔴 Client disconnected",
				zap.String("user_id", client.UserID.String()),
				zap.String("client_id", client.ID),
				zap.Int("remaining_devices", remainingDevices),
				zap.Bool("is_last_device", isLastDevice),
			)

			// If user has no more connected devices, remove from map
			if isLastDevice {
				delete(h.clients, client.UserID)

				logger.Info("📢 Broadcasting user OFFLINE immediately",
					zap.String("user_id", client.UserID.String()),
				)
				// Broadcast user offline status IMMEDIATELY if this was their last device
				h.broadcastUserStatusSync(client.UserID, false)

				// Update user status to offline and set last seen in database (asynchronous)
				go func(userID uuid.UUID) {
					ctx := context.Background()
					userEntity, err := h.userRepo.FindByID(ctx, userID)
					if err != nil {
						logger.Error("❌ Failed to fetch user for status update",
							zap.String("user_id", userID.String()),
							zap.Error(err),
						)
						return
					}

					userEntity.UpdateStatus(user.StatusOffline)
					if err := h.userRepo.Update(ctx, userEntity); err != nil {
						logger.Error("❌ Failed to update user offline status in database",
							zap.String("user_id", userID.String()),
							zap.Error(err),
						)
					} else {
						logger.Info("✅ Updated user offline status and last seen in database",
							zap.String("user_id", userID.String()),
						)
					}
				}(client.UserID)
			}

			h.mu.Unlock()
			return
		}
	}
	h.mu.Unlock()
}

// BroadcastToConversation broadcasts an event to all participants in a conversation
// This is the CORRECT implementation with participant filtering
func (h *Hub) BroadcastToConversation(ctx context.Context, conversationID uuid.UUID, event *Event) error {
	logger.Info("🔔 BroadcastToConversation called",
		zap.String("conversation_id", conversationID.String()),
		zap.String("event_type", string(event.Type)),
	)

	// Get conversation participants from database
	logger.Info("🔍 Querying conversation participants from DB")
	participants, err := h.conversationRepo.FindParticipants(ctx, conversationID)
	if err != nil {
		logger.Error("❌ Failed to get conversation participants for broadcast",
			zap.String("conversation_id", conversationID.String()),
			zap.Error(err),
		)
		return err
	}

	logger.Info("✅ Got participants from DB",
		zap.Int("count", len(participants)),
	)

	// Marshal event once
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event", zap.Error(err))
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	logger.Info("🔍 Checking connected clients",
		zap.Int("total_online_users", len(h.clients)),
	)

	// Track successful sends for logging
	successCount := 0
	totalDevices := 0

	// Broadcast to each participant's connected devices
	for _, participant := range participants {
		logger.Info("🔍 Checking participant",
			zap.String("user_id", participant.UserID.String()),
		)

		if clientMap, ok := h.clients[participant.UserID]; ok {
			totalDevices += len(clientMap)
			logger.Info("✅ Participant is online",
				zap.String("user_id", participant.UserID.String()),
				zap.Int("devices", len(clientMap)),
			)

			for _, client := range clientMap {
				// Non-blocking send with synchronous delivery
				select {
				case client.send <- data:
					successCount++
					logger.Info("✅ Sent to client",
						zap.String("user_id", participant.UserID.String()),
						zap.String("client_id", client.ID),
					)
				default:
					logger.Warn("❌ Client send buffer full",
						zap.String("user_id", participant.UserID.String()),
						zap.String("client_id", client.ID),
					)
				}
			}
		} else {
			logger.Info("⚠️ Participant is offline",
				zap.String("user_id", participant.UserID.String()),
			)
		}
	}

	logger.Debug("Broadcast to conversation",
		zap.String("conversation_id", conversationID.String()),
		zap.String("event_type", string(event.Type)),
		zap.Int("participants", len(participants)),
		zap.Int("total_devices", totalDevices),
		zap.Int("delivered", successCount),
	)

	return nil
}

// BroadcastToUser broadcasts an event to a specific user (all their devices)
func (h *Hub) BroadcastToUser(userID uuid.UUID, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event", zap.Error(err))
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clientMap, ok := h.clients[userID]; ok {
		successCount := 0
		for _, client := range clientMap {
			select {
			case client.send <- data:
				successCount++
			default:
				logger.Warn("Client send buffer full",
					zap.String("user_id", userID.String()),
					zap.String("client_id", client.ID),
				)
			}
		}

		logger.Debug("Broadcast to user",
			zap.String("user_id", userID.String()),
			zap.String("event_type", string(event.Type)),
			zap.Int("devices", len(clientMap)),
			zap.Int("delivered", successCount),
		)
	} else {
		logger.Debug("User not connected",
			zap.String("user_id", userID.String()),
			zap.String("event_type", string(event.Type)),
		)
	}

	return nil
}

// BroadcastToUsers broadcasts an event to multiple users
func (h *Hub) BroadcastToUsers(userIDs []uuid.UUID, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event", zap.Error(err))
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	totalDevices := 0
	successCount := 0

	for _, userID := range userIDs {
		if clientMap, ok := h.clients[userID]; ok {
			totalDevices += len(clientMap)
			for _, client := range clientMap {
				select {
				case client.send <- data:
					successCount++
				default:
					logger.Warn("Client send buffer full",
						zap.String("user_id", userID.String()),
						zap.String("client_id", client.ID),
					)
				}
			}
		}
	}

	logger.Debug("Broadcast to users",
		zap.String("event_type", string(event.Type)),
		zap.Int("users", len(userIDs)),
		zap.Int("total_devices", totalDevices),
		zap.Int("delivered", successCount),
	)

	return nil
}

// IsUserOnline checks if a user has any connected devices
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clientMap, ok := h.clients[userID]
	return ok && len(clientMap) > 0
}

// GetOnlineUsers returns a list of all online user IDs
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}

	return users
}

// GetUserDeviceCount returns the number of connected devices for a user
func (h *Hub) GetUserDeviceCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clientMap, ok := h.clients[userID]; ok {
		return len(clientMap)
	}
	return 0
}

// Stats returns hub statistics
func (h *Hub) Stats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalDevices := 0
	for _, clientMap := range h.clients {
		totalDevices += len(clientMap)
	}

	return map[string]interface{}{
		"online_users":  len(h.clients),
		"total_devices": totalDevices,
	}
}

// broadcastUserStatusSync broadcasts user online/offline status synchronously
// This is a lightweight broadcast that doesn't query the database
// IMPORTANT: Called with h.mu already locked - does NOT lock again
func (h *Hub) broadcastUserStatusSync(userID uuid.UUID, isOnline bool) {
	eventType := EventUserOnline
	if !isOnline {
		eventType = EventUserOffline
	}

	event, err := NewEvent(eventType, map[string]interface{}{
		"user_id":   userID.String(),
		"is_online": isOnline,
	})
	if err != nil {
		logger.Error("❌ Failed to create user status event", zap.Error(err))
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("❌ Failed to marshal user status event", zap.Error(err))
		return
	}

	successCount := 0
	totalDevices := 0
	status := "ONLINE"
	if !isOnline {
		status = "OFFLINE"
	}

	// Broadcast to ALL connected users (mutex already held by caller)
	for recipientUserID, clientMap := range h.clients {
		totalDevices += len(clientMap)
		for _, client := range clientMap {
			select {
			case client.send <- data:
				successCount++
				logger.Info("✅ Sent presence update",
					zap.String("status", status),
					zap.String("changed_user", userID.String()),
					zap.String("sent_to", recipientUserID.String()),
					zap.String("client_id", client.ID),
				)
			default:
				logger.Warn("⚠️ Client send buffer full during presence broadcast",
					zap.String("user_id", client.UserID.String()),
					zap.String("client_id", client.ID),
				)
			}
		}
	}

	logger.Info("📢 Broadcast user status complete",
		zap.String("user_id", userID.String()),
		zap.String("status", status),
		zap.Int("total_devices", totalDevices),
		zap.Int("delivered", successCount),
	)
}
