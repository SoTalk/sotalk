# SoTalk Backend-Managed Encryption Protocol

## Overview
This document outlines a **backend-managed encryption system** where all encryption/decryption logic runs on the server. The client sends plaintext, and the backend handles encryption before storage and decryption before delivery.

---

## Architecture Decision

### Why Backend-Managed Encryption?

**Traditional E2E Encryption (Signal Protocol)**:
- ✅ Maximum security (server can't read messages)
- ❌ Complex client implementation
- ❌ Requires native crypto libraries on client
- ❌ Key management burden on client
- ❌ Difficult to implement search, moderation
- ❌ No message sync across devices easily

**Backend-Managed Encryption (Our Approach)**:
- ✅ Simpler client implementation (just HTTP/WebSocket)
- ✅ All crypto logic centralized on backend
- ✅ Easier to maintain and update
- ✅ Server-side features (search, moderation) possible
- ✅ Better user experience (auto-sync, recovery)
- ⚠️ Server can decrypt messages (acceptable for our use case)
- ✅ Still encrypted at rest and in transit
- ✅ Per-conversation encryption keys

---

## Encryption Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│              Backend-Managed Encryption Flow                     │
└──────────────────────────────────────────────────────────────────┘

Client A (Sender)                Backend                Client B (Receiver)
     │                              │                          │
     │  1. Send Plaintext Message   │                          │
     ├─────────────────────────────>│                          │
     │  POST /messages              │                          │
     │  {content: "Hello"}          │                          │
     │                              │                          │
     │                              ├─ Derive Conversation Key │
     │                              ├─ Generate Message Nonce  │
     │                              ├─ Encrypt with AES-256-GCM│
     │                              ├─ Store Encrypted Content │
     │                              │   in Database ([]byte)    │
     │                              │                          │
     │  2. Success Response         │                          │
     │  <─────────────────────────────                          │
     │  {id, status: "sent"}        │                          │
     │                              │                          │
     │                              │  3. WebSocket Notification│
     │                              ├─────────────────────────>│
     │                              │  {new_message event}      │
     │                              │                          │
     │                              │  4. Fetch Messages        │
     │                              │  <────────────────────────┤
     │                              │  GET /messages?conv_id    │
     │                              │                          │
     │                              ├─ Retrieve Encrypted      │
     │                              ├─ Derive Conversation Key │
     │                              ├─ Decrypt with AES-256-GCM│
     │                              │                          │
     │                              │  5. Return Plaintext      │
     │                              ├─────────────────────────>│
     │                              │  {content: "Hello"}       │
     │                              │                          │
```

---

## Key Management Strategy

### 1. Master Encryption Key (Server Secret)

Stored in environment variables, **never** in database:

```bash
# .env
ENCRYPTION_MASTER_KEY=base64_encoded_32_byte_key
```

This master key is used to derive all conversation keys.

### 2. Conversation Keys (Derived)

Each conversation gets a unique encryption key derived from:
- Master key
- Conversation ID
- Creation timestamp (for key rotation)

```go
// Derive conversation key
conversationKey = HKDF-SHA256(
    masterKey,
    conversationID + timestamp,
    "conversation-encryption-v1",
    32 bytes
)
```

### 3. Message Keys (Per-Message Nonce)

Each message encrypted with:
- Conversation Key (symmetric key)
- Random Nonce (12 bytes, unique per message)
- AES-256-GCM (authenticated encryption)

```go
// Encrypt message
ciphertext, tag = AES-256-GCM.Encrypt(
    key: conversationKey,
    nonce: randomNonce (12 bytes),
    plaintext: messageContent,
    additionalData: messageID + senderID
)
```

---

## Database Schema Updates

### Messages Table (Already Exists)

```sql
-- No changes needed!
-- Content field already stores encrypted data as BYTEA
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content BYTEA NOT NULL,              -- Encrypted content
    content_type VARCHAR(20),
    signature TEXT,
    reply_to_id UUID,
    status VARCHAR(20),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Encryption Metadata Table (New)

Store encryption metadata per message:

```sql
CREATE TABLE message_encryption_metadata (
    message_id UUID PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    nonce BYTEA NOT NULL,                -- 12 bytes for AES-GCM
    tag BYTEA NOT NULL,                  -- 16 bytes authentication tag
    algorithm VARCHAR(20) DEFAULT 'AES-256-GCM',
    key_version INTEGER DEFAULT 1,       -- For key rotation
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_encryption_metadata_message ON message_encryption_metadata(message_id);
```

### Conversation Encryption Keys Cache (Optional - Redis)

```
Key: "conv_key:{conversation_id}"
Value: base64(derived_key)
TTL: 1 hour
```

---

## Package Structure

```
sotalk/
├── pkg/
│   └── crypto/
│       ├── encryption.go        # Core encryption functions
│       ├── keys.go              # Key derivation (HKDF)
│       ├── aes.go               # AES-256-GCM implementation
│       └── types.go             # EncryptedData struct
│
├── internal/
│   ├── infrastructure/
│   │   └── encryption/
│   │       ├── service.go       # Encryption service implementation
│   │       ├── conversation_key_manager.go
│   │       └── cache.go         # Redis caching for keys
│   │
│   ├── domain/
│   │   └── encryption/
│   │       ├── entity.go        # EncryptionMetadata entity
│   │       └── repository.go    # Repository interface
│   │
│   ├── repository/
│   │   └── postgres/
│   │       ├── encryption_metadata_repository.go
│   │       └── models.go        # Add EncryptionMetadata model
│   │
│   └── usecase/
│       └── message/
│           └── service.go       # Update to use encryption
│
└── migrations/
    ├── 000016_create_message_encryption_metadata.up.sql
    └── 000016_create_message_encryption_metadata.down.sql
```

---

## Implementation Details

### 1. Encryption Service Interface

```go
// pkg/crypto/encryption.go
package crypto

type EncryptionService interface {
    // Encrypt plaintext message for a conversation
    EncryptMessage(conversationID uuid.UUID, plaintext []byte) (*EncryptedData, error)

    // Decrypt ciphertext for a message
    DecryptMessage(conversationID uuid.UUID, encrypted *EncryptedData) ([]byte, error)

    // Derive conversation key
    DeriveConversationKey(conversationID uuid.UUID, timestamp time.Time) ([]byte, error)

    // Rotate conversation key (for enhanced security)
    RotateConversationKey(conversationID uuid.UUID) error
}

type EncryptedData struct {
    Ciphertext []byte  // Encrypted content
    Nonce      []byte  // 12-byte nonce for AES-GCM
    Tag        []byte  // 16-byte authentication tag
    Algorithm  string  // "AES-256-GCM"
    KeyVersion int     // For key rotation support
}
```

### 2. Core Encryption Functions

```go
// pkg/crypto/aes.go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "fmt"
)

// EncryptAESGCM encrypts data using AES-256-GCM
func EncryptAESGCM(key, plaintext, additionalData []byte) (ciphertext, nonce, tag []byte, err error) {
    // Validate key size (must be 32 bytes for AES-256)
    if len(key) != 32 {
        return nil, nil, nil, fmt.Errorf("invalid key size: expected 32 bytes, got %d", len(key))
    }

    // Create AES cipher
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
    }

    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
    }

    // Generate random nonce (12 bytes for GCM)
    nonce = make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
    }

    // Encrypt and authenticate
    // GCM Seal returns: ciphertext || tag
    sealed := gcm.Seal(nil, nonce, plaintext, additionalData)

    // Split ciphertext and tag
    // Tag is last 16 bytes
    tagSize := gcm.Overhead() // 16 bytes
    ciphertext = sealed[:len(sealed)-tagSize]
    tag = sealed[len(sealed)-tagSize:]

    return ciphertext, nonce, tag, nil
}

// DecryptAESGCM decrypts data using AES-256-GCM
func DecryptAESGCM(key, ciphertext, nonce, tag, additionalData []byte) (plaintext []byte, err error) {
    // Validate key size
    if len(key) != 32 {
        return nil, fmt.Errorf("invalid key size: expected 32 bytes, got %d", len(key))
    }

    // Create AES cipher
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }

    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }

    // Validate nonce size
    if len(nonce) != gcm.NonceSize() {
        return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", gcm.NonceSize(), len(nonce))
    }

    // Reconstruct sealed data (ciphertext || tag)
    sealed := append(ciphertext, tag...)

    // Decrypt and verify
    plaintext, err = gcm.Open(nil, nonce, sealed, additionalData)
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }

    return plaintext, nil
}
```

### 3. Key Derivation

```go
// pkg/crypto/keys.go
package crypto

import (
    "crypto/sha256"
    "fmt"
    "golang.org/x/crypto/hkdf"
    "io"
)

// DeriveKey derives a key using HKDF-SHA256
func DeriveKey(masterKey, salt, info []byte, keySize int) ([]byte, error) {
    h := hkdf.New(sha256.New, masterKey, salt, info)

    derivedKey := make([]byte, keySize)
    if _, err := io.ReadFull(h, derivedKey); err != nil {
        return nil, fmt.Errorf("failed to derive key: %w", err)
    }

    return derivedKey, nil
}

// DeriveConversationKey derives a unique key for a conversation
func DeriveConversationKey(masterKey []byte, conversationID string, timestamp int64, version int) ([]byte, error) {
    // Salt combines conversation ID and timestamp
    salt := []byte(fmt.Sprintf("%s:%d:%d", conversationID, timestamp, version))

    // Info string for key separation
    info := []byte("sotalk-conversation-encryption-v1")

    // Derive 32-byte key for AES-256
    return DeriveKey(masterKey, salt, info, 32)
}
```

### 4. Encryption Service Implementation

```go
// internal/infrastructure/encryption/service.go
package encryption

import (
    "context"
    "encoding/base64"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/yourusername/sotalk/pkg/crypto"
    "github.com/yourusername/sotalk/pkg/redis"
)

type service struct {
    masterKey    []byte
    redisClient  *redis.Client
}

func NewService(masterKeyB64 string, redisClient *redis.Client) (*service, error) {
    masterKey, err := base64.StdEncoding.DecodeString(masterKeyB64)
    if err != nil {
        return nil, fmt.Errorf("invalid master key: %w", err)
    }

    if len(masterKey) != 32 {
        return nil, fmt.Errorf("master key must be 32 bytes")
    }

    return &service{
        masterKey:   masterKey,
        redisClient: redisClient,
    }, nil
}

func (s *service) EncryptMessage(ctx context.Context, conversationID uuid.UUID, plaintext []byte) (*crypto.EncryptedData, error) {
    // Derive conversation key
    convKey, err := s.getOrDeriveConversationKey(ctx, conversationID)
    if err != nil {
        return nil, err
    }

    // Additional authenticated data (prevents ciphertext manipulation)
    aad := []byte(conversationID.String())

    // Encrypt
    ciphertext, nonce, tag, err := crypto.EncryptAESGCM(convKey, plaintext, aad)
    if err != nil {
        return nil, fmt.Errorf("encryption failed: %w", err)
    }

    return &crypto.EncryptedData{
        Ciphertext: ciphertext,
        Nonce:      nonce,
        Tag:        tag,
        Algorithm:  "AES-256-GCM",
        KeyVersion: 1,
    }, nil
}

func (s *service) DecryptMessage(ctx context.Context, conversationID uuid.UUID, encrypted *crypto.EncryptedData) ([]byte, error) {
    // Derive conversation key
    convKey, err := s.getOrDeriveConversationKey(ctx, conversationID)
    if err != nil {
        return nil, err
    }

    // Additional authenticated data
    aad := []byte(conversationID.String())

    // Decrypt
    plaintext, err := crypto.DecryptAESGCM(convKey, encrypted.Ciphertext, encrypted.Nonce, encrypted.Tag, aad)
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }

    return plaintext, nil
}

func (s *service) getOrDeriveConversationKey(ctx context.Context, conversationID uuid.UUID) ([]byte, error) {
    // Try to get from cache
    cacheKey := fmt.Sprintf("conv_key:%s", conversationID.String())
    cachedKey, err := s.redisClient.Get(ctx, cacheKey)

    if err == nil && cachedKey != "" {
        // Decode from base64
        return base64.StdEncoding.DecodeString(cachedKey)
    }

    // Derive new key
    // Use conversation creation time as salt component (for deterministic derivation)
    timestamp := time.Now().Unix() / 86400 // Change key daily (optional)

    convKey, err := crypto.DeriveConversationKey(
        s.masterKey,
        conversationID.String(),
        timestamp,
        1, // key version
    )
    if err != nil {
        return nil, err
    }

    // Cache for 1 hour
    keyB64 := base64.StdEncoding.EncodeToString(convKey)
    s.redisClient.Set(ctx, cacheKey, keyB64, time.Hour)

    return convKey, nil
}
```

---

## Integration with Message Service

### Update Message Service

```go
// internal/usecase/message/service.go

type service struct {
    messageRepo      message.Repository
    conversationRepo conversation.Repository
    userRepo         user.Repository
    encryptionSvc    encryption.Service  // ADD THIS
    metadataRepo     encryption.MetadataRepository  // ADD THIS
}

func (s *service) SendMessage(ctx context.Context, senderID uuid.UUID, req *dto.SendMessageRequest) (*dto.SendMessageResponse, error) {
    // ... existing validation ...

    // Get or create conversation
    conversationID, err := s.GetOrCreateDirectConversation(ctx, senderID, req.RecipientID)
    if err != nil {
        return nil, fmt.Errorf("failed to get/create conversation: %w", err)
    }

    // ✨ NEW: Encrypt message content
    encrypted, err := s.encryptionSvc.EncryptMessage(ctx, conversationID, req.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt message: %w", err)
    }

    // Create message with encrypted content
    msg := message.NewMessage(
        conversationID,
        senderID,
        encrypted.Ciphertext,  // Store encrypted content
        message.ContentType(req.ContentType),
    )

    // Save message
    msg.MarkAsSent()
    if err := s.messageRepo.Create(ctx, msg); err != nil {
        return nil, fmt.Errorf("failed to create message: %w", err)
    }

    // ✨ NEW: Save encryption metadata
    metadata := &encryption.Metadata{
        MessageID:  msg.ID,
        Nonce:      encrypted.Nonce,
        Tag:        encrypted.Tag,
        Algorithm:  encrypted.Algorithm,
        KeyVersion: encrypted.KeyVersion,
    }
    if err := s.metadataRepo.Create(ctx, metadata); err != nil {
        return nil, fmt.Errorf("failed to save encryption metadata: %w", err)
    }

    // ... rest of existing code ...

    // ✨ IMPORTANT: Return PLAINTEXT to sender for confirmation
    messageDTO := toMessageDTO(msg, sender, recipient)
    messageDTO.Content = req.Content  // Return original plaintext

    return &dto.SendMessageResponse{
        Message: messageDTO,
    }, nil
}

func (s *service) GetMessages(ctx context.Context, userID uuid.UUID, req *dto.GetMessagesRequest) (*dto.GetMessagesResponse, error) {
    // ... existing code to fetch messages ...

    // ✨ NEW: Decrypt each message before returning
    messageDTOs := make([]dto.MessageDTO, len(messages))
    for i, msg := range messages {
        // Fetch encryption metadata
        metadata, err := s.metadataRepo.FindByMessageID(ctx, msg.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to get encryption metadata: %w", err)
        }

        // Decrypt message
        encrypted := &crypto.EncryptedData{
            Ciphertext: msg.Content,
            Nonce:      metadata.Nonce,
            Tag:        metadata.Tag,
            Algorithm:  metadata.Algorithm,
            KeyVersion: metadata.KeyVersion,
        }

        plaintext, err := s.encryptionSvc.DecryptMessage(ctx, msg.ConversationID, encrypted)
        if err != nil {
            return nil, fmt.Errorf("failed to decrypt message: %w", err)
        }

        // Get sender info
        sender, _ := s.userRepo.FindByID(ctx, msg.SenderID)

        // Create DTO with decrypted content
        msgDTO := toMessageDTO(msg, sender, nil)
        msgDTO.Content = plaintext  // Return plaintext
        messageDTOs[i] = msgDTO
    }

    // ... return response ...
}
```

---

## API Flow

### 1. Send Message (Client → Backend)

```http
POST /api/v1/messages
Authorization: Bearer {token}
Content-Type: application/json

{
    "recipient_id": "uuid",
    "content": "Hello, this is plaintext!",  // ✨ Client sends PLAINTEXT
    "content_type": "text"
}
```

**Backend Processing**:
1. Validate sender & recipient
2. Get/create conversation
3. **Encrypt content** → `EncryptMessage()`
4. Store encrypted `content` as `BYTEA` in database
5. Store encryption metadata (nonce, tag)
6. Return success (with plaintext for confirmation)

### 2. Fetch Messages (Client ← Backend)

```http
GET /api/v1/conversations/{id}/messages
Authorization: Bearer {token}
```

**Backend Processing**:
1. Fetch encrypted messages from database
2. Fetch encryption metadata
3. **Decrypt each message** → `DecryptMessage()`
4. Return plaintext messages to client

```json
{
    "messages": [
        {
            "id": "uuid",
            "content": "Hello, this is plaintext!",  // ✨ Backend sends PLAINTEXT
            "content_type": "text",
            "sender_id": "uuid",
            "created_at": "2024-01-01T00:00:00Z"
        }
    ]
}
```

---

## Security Considerations

### ✅ Protections

1. **Encrypted at Rest**: Messages stored encrypted in database
2. **Encrypted in Transit**: HTTPS/TLS for all API calls
3. **Per-Conversation Keys**: Each conversation has unique encryption key
4. **Authenticated Encryption**: AES-GCM prevents tampering
5. **Key Derivation**: Conversation keys derived from master key (not stored)
6. **Master Key Security**: Stored in environment, never in database/code

### ⚠️ Limitations

1. **Backend Can Decrypt**: Server has master key, can decrypt all messages
2. **Not True E2E**: Not suitable for extremely sensitive communications
3. **Trust Requirement**: Users must trust the server operator
4. **Compliance**: May not meet regulatory requirements for E2E encryption

### 🔒 Mitigation

1. **Access Controls**: Strict access to master key (only production servers)
2. **Audit Logging**: Log all decryption operations
3. **Key Rotation**: Periodic master key rotation
4. **HSM (Future)**: Store master key in Hardware Security Module
5. **Transparency**: Document that encryption is server-side, not E2E

---

## Implementation Steps

### Phase 1: Core Crypto Package (2-3 hours)

- [ ] Create `pkg/crypto/aes.go` - AES-256-GCM functions
- [ ] Create `pkg/crypto/keys.go` - HKDF key derivation
- [ ] Create `pkg/crypto/types.go` - EncryptedData struct
- [ ] Write unit tests for encryption/decryption

### Phase 2: Infrastructure Layer (2-3 hours)

- [ ] Create `internal/infrastructure/encryption/service.go`
- [ ] Implement conversation key derivation
- [ ] Implement Redis caching for keys
- [ ] Generate master key and add to `.env`

### Phase 3: Domain & Repository (1-2 hours)

- [ ] Create `internal/domain/encryption/entity.go`
- [ ] Create encryption metadata repository interface
- [ ] Implement Postgres repository
- [ ] Create database migration for metadata table

### Phase 4: Message Service Integration (2-3 hours)

- [ ] Update `internal/usecase/message/service.go`
- [ ] Add encryption on message send
- [ ] Add decryption on message fetch
- [ ] Handle encryption errors gracefully

### Phase 5: Testing & Validation (2-3 hours)

- [ ] Integration tests for encrypt/decrypt flow
- [ ] Test message send/receive with encryption
- [ ] Test key caching
- [ ] Test error handling

### Phase 6: Documentation (1 hour)

- [ ] Update API documentation
- [ ] Document encryption approach
- [ ] Create security guidelines

---

## Master Key Generation

```bash
# Generate a secure 32-byte (256-bit) master key
openssl rand -base64 32

# Example output:
# 7x9k3mP2vQ8wR5tY6uI4oP1aS2dF3gH4jK5lZ6xC7vB8n==

# Add to .env:
ENCRYPTION_MASTER_KEY=7x9k3mP2vQ8wR5tY6uI4oP1aS2dF3gH4jK5lZ6xC7vB8n==
```

---

## Testing

```go
// Test encryption/decryption round-trip
func TestEncryptDecryptRoundTrip(t *testing.T) {
    masterKey, _ := base64.StdEncoding.DecodeString("test_key_32_bytes_base64_encoded==")
    svc := encryption.NewService(base64.StdEncoding.EncodeToString(masterKey), redisClient)

    conversationID := uuid.New()
    plaintext := []byte("Hello, World!")

    // Encrypt
    encrypted, err := svc.EncryptMessage(ctx, conversationID, plaintext)
    require.NoError(t, err)
    require.NotEmpty(t, encrypted.Ciphertext)
    require.NotEmpty(t, encrypted.Nonce)
    require.NotEmpty(t, encrypted.Tag)

    // Decrypt
    decrypted, err := svc.DecryptMessage(ctx, conversationID, encrypted)
    require.NoError(t, err)
    require.Equal(t, plaintext, decrypted)
}
```

---

## Performance Considerations

1. **Key Caching**: Cache conversation keys in Redis (1-hour TTL)
2. **Batch Decryption**: Decrypt multiple messages in parallel
3. **Lazy Decryption**: Only decrypt messages when fetched, not on storage
4. **Index Optimization**: No index on encrypted content (can't search anyway)

---

## Future Enhancements

1. **Key Rotation**: Periodic master key rotation with versioning
2. **HSM Integration**: Store master key in Hardware Security Module
3. **Per-User Keys**: Derive keys from user's Solana wallet (hybrid approach)
4. **Searchable Encryption**: Research homomorphic encryption for search
5. **Client-Side Option**: Offer E2E mode for ultra-sensitive conversations

---

## Conclusion

This backend-managed encryption approach provides:
- ✅ **Strong encryption at rest and in transit**
- ✅ **Simple client implementation** (no crypto libraries needed)
- ✅ **Centralized security management**
- ✅ **Easy to maintain and update**
- ✅ **Supports server-side features** (moderation, search, etc.)

Perfect balance between security and usability for a messaging platform like SoTalk.

**Estimated Implementation Time**: 10-14 hours total

Ready to implement Phase 1?
