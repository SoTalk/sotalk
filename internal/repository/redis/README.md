# Redis Caching Layer

This package provides Redis-based caching for the Solana messaging platform. All caches follow Clean Architecture principles and are designed for high performance and scalability.

## Available Caches

### 1. Session Cache (`session_cache.go`)
**Purpose:** Manages user authentication sessions and JWT token data.

**Features:**
- Store/retrieve session data with TTL
- Multi-device support (multiple sessions per user)
- Token blacklisting for logout
- Session activity tracking
- Bulk session management

**Usage Example:**
```go
// Initialize
sessionCache := redis.NewSessionCache(redisClient)

// Store a session (e.g., after successful login)
sessionData := &redis.SessionData{
    UserID:        userID,
    WalletAddress: "abc123...",
    Username:      "john_doe",
    DisplayName:   "John Doe",
    CreatedAt:     time.Now(),
    LastActivity:  time.Now(),
}
err := sessionCache.Set(ctx, sessionID, sessionData, 15*time.Minute)

// Retrieve a session (e.g., in JWT middleware)
session, err := sessionCache.Get(ctx, sessionID)
if session == nil {
    return errors.New("session not found")
}

// Extend session TTL (on user activity)
err = sessionCache.Extend(ctx, sessionID, 15*time.Minute)

// Logout - delete session
err = sessionCache.Delete(ctx, sessionID)

// Blacklist a token
err = sessionCache.InvalidateToken(ctx, tokenID, 24*time.Hour)

// Check if token is blacklisted
isBlacklisted, err := sessionCache.IsTokenBlacklisted(ctx, tokenID)
```

**Integration Points:**
- Auth service: Store sessions on login, validate on protected endpoints
- JWT middleware: Check session validity and blacklist
- Logout endpoint: Delete session and blacklist token
- Multi-device management: Track and manage all user sessions

---

### 2. Message Cache (`message_cache.go`)
**Purpose:** Caches recent messages per conversation for fast retrieval.

**Features:**
- Individual message caching
- Conversation message list (ordered by timestamp)
- Unread message counters
- Message status updates
- Automatic cache size limits (last 100 messages per conversation)

**Usage Example:**
```go
// Initialize
messageCache := redis.NewMessageCache(redisClient)

// Cache a new message (after sending)
cachedMsg := &redis.CachedMessage{
    ID:             messageID,
    ConversationID: conversationID,
    SenderID:       senderID,
    Content:        encryptedContent,
    ContentType:    "text",
    Status:         "sent",
    CreatedAt:      time.Now(),
}
err := messageCache.SetMessage(ctx, cachedMsg, 15*time.Minute)

// Add message to conversation's message list
err = messageCache.AddMessageToConversation(ctx, conversationID, cachedMsg)

// Retrieve recent messages (before hitting DB)
messages, err := messageCache.GetConversationMessages(ctx, conversationID, 50)
if messages == nil {
    // Cache miss - fetch from DB and populate cache
    messages = fetchFromDB()
    messageCache.SetConversationMessages(ctx, conversationID, messages, 15*time.Minute)
}

// Update message status (delivered, read)
err = messageCache.UpdateMessageStatus(ctx, messageID, "read")

// Manage unread counts
err = messageCache.IncrementUnreadCount(ctx, conversationID, userID)
unreadCount, err := messageCache.GetUnreadCount(ctx, conversationID, userID)
err = messageCache.ResetUnreadCount(ctx, conversationID, userID) // After marking as read

// Delete message from cache
err = messageCache.DeleteMessage(ctx, messageID)
```

**Integration Points:**
- Message service: Cache messages on send, retrieve for pagination
- WebSocket handler: Update cache when messages are delivered/read
- Conversation list: Use cached unread counts
- Message deletion: Invalidate cache

---

### 3. Presence Cache (`presence_cache.go`)
**Purpose:** Tracks real-time user online/offline status and typing indicators.

**Features:**
- Online/offline/away status tracking
- Last seen timestamps
- Typing indicators per conversation
- Bulk status queries
- Automatic TTL-based offline detection

**Usage Example:**
```go
// Initialize
presenceCache := redis.NewPresenceCache(redisClient)

// Set user online (when WebSocket connects)
err := presenceCache.SetOnline(ctx, userID, 5*time.Minute)

// Heartbeat - extend online status
err = presenceCache.ExtendOnline(ctx, userID, 5*time.Minute)

// Check if user is online
isOnline, err := presenceCache.IsOnline(ctx, userID)

// Get user status
status, err := presenceCache.GetStatus(ctx, userID) // returns: online, offline, away

// Set user offline (when WebSocket disconnects)
err = presenceCache.SetOffline(ctx, userID)

// Get last seen time
lastSeen, err := presenceCache.GetLastSeen(ctx, userID)

// Bulk status check (for contact list)
userIDs := []uuid.UUID{user1, user2, user3}
statusMap, err := presenceCache.GetBulkStatus(ctx, userIDs)
// Returns map[uuid.UUID]PresenceStatus

// Typing indicators
err = presenceCache.SetTyping(ctx, conversationID, userID, 10*time.Second)
isTyping, err := presenceCache.IsTyping(ctx, conversationID, userID)
typingUsers, err := presenceCache.GetTypingUsers(ctx, conversationID)
err = presenceCache.ClearTyping(ctx, conversationID, userID)

// Get all online users
onlineUsers, err := presenceCache.GetOnlineUsers(ctx)
onlineCount, err := presenceCache.GetOnlineCount(ctx)
```

**Integration Points:**
- WebSocket hub: Update presence on connect/disconnect (already integrated)
- Contact list: Show online/offline status
- Conversation view: Display last seen time
- Typing indicators: Real-time typing status
- Admin dashboard: Monitor online user count

---

### 4. Conversation Cache (`conversation_cache.go`)
**Purpose:** Caches user's conversation list with metadata.

**Features:**
- Sorted conversation list (by last message timestamp)
- Conversation metadata (name, avatar, last message)
- Unread count per conversation
- Pinned conversations
- Muted conversations
- Real-time updates on new messages

**Usage Example:**
```go
// Initialize
conversationCache := redis.NewConversationCache(redisClient)

// Cache user's conversation list
conversations := []*redis.CachedConversation{
    {
        ID:              convID1,
        Type:            "direct",
        Name:            "John Doe",
        LastMessageText: "Hello!",
        LastMessageAt:   time.Now(),
        UnreadCount:     3,
        IsPinned:        true,
    },
    // ...more conversations
}
err := conversationCache.SetUserConversations(ctx, userID, conversations, 10*time.Minute)

// Retrieve conversation list (fast!)
conversations, err := conversationCache.GetUserConversations(ctx, userID, 50)
if conversations == nil {
    // Cache miss - fetch from DB
    conversations = fetchFromDB()
    conversationCache.SetUserConversations(ctx, userID, conversations, 10*time.Minute)
}

// Update on new message
err = conversationCache.UpdateLastMessage(
    ctx,
    userID,
    conversationID,
    messageID,
    "New message text",
    time.Now(),
)

// Manage unread counts
err = conversationCache.IncrementUnread(ctx, userID, conversationID)
err = conversationCache.ResetUnread(ctx, userID, conversationID)
totalUnread, err := conversationCache.GetTotalUnreadCount(ctx, userID)

// Pin/unpin conversation
err = conversationCache.SetPinned(ctx, userID, conversationID, true)

// Mute/unmute conversation
err = conversationCache.SetMuted(ctx, userID, conversationID, true)

// Remove conversation from cache
err = conversationCache.RemoveConversation(ctx, userID, conversationID)

// Clear all conversations (e.g., on logout)
err = conversationCache.DeleteUserConversations(ctx, userID)
```

**Integration Points:**
- Conversation list endpoint: Fast retrieval of user's conversations
- New message handler: Update last message and unread count
- Message read handler: Reset unread count
- Pin/mute endpoints: Update conversation metadata

---

## Cache Keys Format

All cache keys follow a consistent naming pattern:

```
session:{sessionID}                          # User session data
user_session:{userID}:{sessionID}            # Session lookup by user
blacklist:token:{tokenID}                    # Blacklisted tokens

message:{messageID}                          # Individual message
conversation:messages:{conversationID}       # Conversation message list (sorted set)
unread:{conversationID}:{userID}             # Unread count

presence:{userID}                            # User online status
lastseen:{userID}                           # Last seen timestamp
typing:{conversationID}:{userID}            # Typing indicator

conversations:user:{userID}                  # User's conversation list (sorted set)
```

---

## Cache Invalidation Strategy

### Write-Through Pattern
When data is updated in the database, also update the cache:
```go
// Update database
err := repo.UpdateMessage(ctx, message)

// Update cache
err = messageCache.SetMessage(ctx, cachedMessage, 15*time.Minute)
```

### Cache-Aside Pattern
Read from cache first, if miss, read from DB and populate cache:
```go
// Try cache first
messages, err := messageCache.GetConversationMessages(ctx, conversationID, 50)
if messages == nil {
    // Cache miss - fetch from DB
    messages, err = messageRepo.GetMessages(ctx, conversationID, 50)
    if err != nil {
        return nil, err
    }
    // Populate cache
    messageCache.SetConversationMessages(ctx, conversationID, messages, 15*time.Minute)
}
return messages, nil
```

### Invalidation on Delete
```go
// Delete from database
err := repo.DeleteMessage(ctx, messageID)

// Invalidate cache
err = messageCache.DeleteMessage(ctx, messageID)
err = messageCache.DeleteConversationMessages(ctx, conversationID)
```

---

## Performance Considerations

### TTL Recommendations
- **Sessions**: 15 minutes (extend on activity)
- **Messages**: 15 minutes (high volatility)
- **Presence**: 5 minutes with heartbeat extension
- **Conversations**: 10 minutes (moderate volatility)

### Cache Size Limits
- Message cache: Last 100 messages per conversation (using ZREMRANGEBYRANK)
- Presence: TTL-based cleanup (no size limit)
- Sessions: Per-user limit via manual cleanup

### Pipeline Usage
Use Redis pipelines for bulk operations to reduce network round-trips:
```go
pipe := client.Pipeline()
for _, msg := range messages {
    pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: data})
}
pipe.Expire(ctx, key, ttl)
_, err := pipe.Exec(ctx)
```

---

## Integration Checklist

### ✅ Completed
- [x] Redis client initialization in `main.go`
- [x] Session cache repository
- [x] Message cache repository
- [x] Presence cache repository
- [x] Conversation cache repository
- [x] WebSocket hub presence integration

### 🚧 Recommended Integrations

#### Auth Service
```go
// In auth service constructor
type service struct {
    userRepo     user.Repository
    jwtManager   *middleware.JWTManager
    sessionCache *redis.SessionCache  // Add this
}

// After successful login
sessionData := &redis.SessionData{...}
sessionCache.Set(ctx, sessionID, sessionData, 15*time.Minute)

// In JWT middleware
session, err := sessionCache.Get(ctx, sessionID)
if session != nil {
    // Session valid - extend TTL
    sessionCache.Extend(ctx, sessionID, 15*time.Minute)
}
```

#### Message Service
```go
// When sending a message
cachedMsg := &redis.CachedMessage{...}
messageCache.AddMessageToConversation(ctx, conversationID, cachedMsg)
conversationCache.UpdateLastMessage(ctx, recipientID, conversationID, msgID, text, time.Now())
conversationCache.IncrementUnread(ctx, recipientID, conversationID)

// When fetching messages
messages, err := messageCache.GetConversationMessages(ctx, conversationID, limit)
if messages == nil {
    messages = fetchFromDB()
    messageCache.SetConversationMessages(ctx, conversationID, messages, 15*time.Minute)
}

// When marking as read
messageCache.ResetUnreadCount(ctx, conversationID, userID)
conversationCache.ResetUnread(ctx, userID, conversationID)
```

#### User Handler (Presence Endpoint)
```go
// GET /api/v1/users/:id/presence
func (h *UserHandler) GetPresence(c *gin.Context) {
    status, _ := presenceCache.GetStatus(ctx, userID)
    lastSeen, _ := presenceCache.GetLastSeen(ctx, userID)

    c.JSON(200, gin.H{
        "status": status,
        "last_seen": lastSeen,
        "is_online": status == "online",
    })
}

// GET /api/v1/users/presence/bulk
func (h *UserHandler) GetBulkPresence(c *gin.Context) {
    userIDs := []uuid.UUID{...} // from request
    statusMap, _ := presenceCache.GetBulkStatus(ctx, userIDs)
    c.JSON(200, statusMap)
}
```

---

## Monitoring & Debugging

### Check Redis Connection
```bash
redis-cli -h localhost -p 6379
PING  # Should return PONG
```

### View Cached Data
```bash
# List all keys
KEYS *

# View session data
GET session:{sessionID}

# View presence
GET presence:{userID}

# View sorted sets (conversations, messages)
ZRANGE conversations:user:{userID} 0 -1 WITHSCORES
```

### Monitor Performance
```bash
# Monitor all commands
redis-cli MONITOR

# Get stats
redis-cli INFO stats

# Check memory usage
redis-cli INFO memory
```

### Clear Cache (Development)
```bash
# Clear all data (DANGEROUS!)
redis-cli FLUSHALL

# Clear specific database
redis-cli -n 0 FLUSHDB
```

---

## Error Handling

All cache operations should fail gracefully:
```go
// Cache read - if fails, just fetch from DB
messages, err := messageCache.GetConversationMessages(ctx, conversationID, 50)
if err != nil || messages == nil {
    logger.Warn("Cache miss, fetching from DB")
    messages = fetchFromDB()
}

// Cache write - if fails, log but don't block
if err := messageCache.SetMessage(ctx, msg, ttl); err != nil {
    logger.Error("Failed to cache message", zap.Error(err))
    // Continue anyway - DB has the data
}
```

**Golden Rule:** Cache failures should NOT block main application flow. Always have a fallback to the database.

---

## Testing

```go
// Use Redis test container
func TestSessionCache(t *testing.T) {
    // Setup test Redis
    client := setupTestRedis(t)
    defer client.FlushDB(context.Background())

    sessionCache := redis.NewSessionCache(client)

    // Test cache operations
    sessionData := &redis.SessionData{...}
    err := sessionCache.Set(ctx, "test-session", sessionData, time.Minute)
    assert.NoError(t, err)

    retrieved, err := sessionCache.Get(ctx, "test-session")
    assert.NoError(t, err)
    assert.Equal(t, sessionData.UserID, retrieved.UserID)
}
```

---

## Next Steps

1. ✅ Redis infrastructure is ready
2. 🚧 Integrate caches into services (see Integration Checklist)
3. 🚧 Add cache metrics (Prometheus)
4. 🚧 Write integration tests
5. 🚧 Load test with caching enabled
6. 🚧 Monitor cache hit rates in production

---

**Redis Caching Layer Status: ✅ COMPLETE**

The infrastructure is production-ready. Services can now be updated to use these caches for significant performance improvements.
