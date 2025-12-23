# SoTalk Protocol Implementation Plan

## Overview
Custom end-to-end encryption protocol leveraging Solana's Ed25519 keypairs with:
- **Full Double Ratchet** for forward/backward secrecy
- **Multi-device support** with device key management
- **Sender-keys protocol** for efficient group encryption
- **Client-side only** key storage with encrypted backups
- **Solana signature verification** for message authenticity

## Architecture Summary

```
┌─────────────────────────────────────────────────────────────┐
│                    SoTalk Protocol Stack                     │
├─────────────────────────────────────────────────────────────┤
│  Layer 4: Application (Message Encryption/Decryption)       │
├─────────────────────────────────────────────────────────────┤
│  Layer 3: Sender-Keys (Group Chat Encryption)               │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: Double Ratchet (Session Management)               │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: X3DH (Key Exchange using Solana keys)             │
├─────────────────────────────────────────────────────────────┤
│  Layer 0: Cryptographic Primitives (Ed25519, X25519, AES)   │
└─────────────────────────────────────────────────────────────┘
```

## Current State Analysis

**Existing Code:**
- Clean Architecture implemented: `internal/domain/`, `internal/usecase/`, `internal/delivery/`, `internal/repository/`
- Message entity: `internal/domain/message/entity.go` (Content as plain string)
- Solana integration: `internal/infrastructure/solana/client.go`, `pkg/solana/wallet.go`
- Ed25519 signing/verification already working
- golang.org/x/crypto dependency available
- Empty crypto package: `pkg/crypto/` (ready for implementation)

**Changes Needed:**
- Message content must change from `string` to `[]byte` for encrypted data
- Add encryption metadata to Message entity
- Create SoTalk Protocol Go package
- Integrate protocol into message service

---

## Phase 1: Create Standalone Go Package (pkg/sotalk)

### Directory Structure
```
pkg/sotalk/
├── protocol/
│   ├── primitives.go       # Crypto primitives (X25519, AES-GCM, HKDF)
│   ├── x3dh.go             # Extended Triple Diffie-Hellman key exchange
│   ├── double_ratchet.go   # Double Ratchet algorithm
│   ├── sender_keys.go      # Sender-keys for group encryption
│   └── session.go          # Session state management
├── keys/
│   ├── identity.go         # Identity key management (Ed25519 -> X25519)
│   ├── prekeys.go          # Pre-key bundle generation
│   ├── device.go           # Device key management
│   └── storage.go          # Key storage interfaces
├── message/
│   ├── encrypt.go          # Message encryption
│   ├── decrypt.go          # Message decryption
│   └── format.go           # SoTalk message format
├── group/
│   ├── sender_keys.go      # Group sender-keys implementation
│   └── session.go          # Group session management
├── backup/
│   ├── export.go           # Key backup export
│   └── import.go           # Key backup import
├── types.go                # Core types and interfaces
├── errors.go               # Custom errors
└── protocol.go             # Main protocol interface
```

### Key Types
```go
// Core protocol types
type IdentityKeyPair struct {
    PublicKey  [32]byte  // X25519 public key (derived from Ed25519)
    PrivateKey [32]byte  // X25519 private key
}

type Device struct {
    ID              string
    IdentityKey     IdentityKeyPair
    SignedPreKey    PreKey
    OneTimePreKeys  []PreKey
    RegistrationID  uint32
    Timestamp       time.Time
}

type PreKey struct {
    ID         uint32
    KeyPair    IdentityKeyPair
    Signature  []byte  // Signed with Solana Ed25519 key
}

type Session struct {
    ID              string
    RootKey         [32]byte
    SendingChain    ChainKey
    ReceivingChain  ChainKey
    MessageKeys     map[uint32][32]byte  // Skipped message keys
    SendCounter     uint32
    ReceiveCounter  uint32
}

type ChainKey struct {
    Key     [32]byte
    Index   uint32
}

type EncryptedMessage struct {
    Version         uint8
    SenderDeviceID  string
    ReceiverDevice  string
    EphemeralKey    [32]byte  // For X3DH
    Counter         uint32
    PreviousCounter uint32
    Ciphertext      []byte
    AuthTag         []byte
    RatchetKey      [32]byte  // Current ratchet public key
    Signature       []byte    // Solana signature for authenticity
}

// Group encryption
type SenderKeyState struct {
    ChainID      string
    ChainKey     ChainKey
    SigningKey   IdentityKeyPair
    MessageKeys  map[uint32][32]byte
}
```

---

## Phase 2: Core Cryptographic Primitives (pkg/sotalk/protocol/primitives.go)

### Implementation Steps

1. **Ed25519 to X25519 Conversion**
   - Convert Solana Ed25519 keys to X25519 for ECDH
   - Use: `golang.org/x/crypto/curve25519`
   ```go
   func Ed25519ToX25519Public(ed25519Key ed25519.PublicKey) ([32]byte, error)
   func Ed25519ToX25519Private(ed25519Key ed25519.PrivateKey) ([32]byte, error)
   ```

2. **Key Derivation (HKDF)**
   - HKDF-SHA256 for key derivation
   - Use: `golang.org/x/crypto/hkdf`
   ```go
   func DeriveKeys(inputKeyMaterial []byte, salt []byte, info string, outputLen int) ([]byte, error)
   ```

3. **Symmetric Encryption (AES-256-GCM)**
   - AES-256-GCM for message encryption
   - Use: `crypto/aes`, `crypto/cipher`
   ```go
   func Encrypt(key [32]byte, plaintext []byte, associatedData []byte) (ciphertext, tag []byte, err error)
   func Decrypt(key [32]byte, ciphertext, tag, associatedData []byte) ([]byte, error)
   ```

4. **Diffie-Hellman (X25519)**
   - ECDH for shared secret computation
   ```go
   func DH(privateKey, publicKey [32]byte) ([32]byte, error)
   ```

**Files to create:**
- `pkg/sotalk/protocol/primitives.go`

---

## Phase 3: X3DH Key Exchange (pkg/sotalk/protocol/x3dh.go)

### Extended Triple Diffie-Hellman
Modified X3DH that uses Solana keys as identity keys.

**Key Agreement:**
```
DH1 = DH(IK_A, SPK_B)      # Identity to Signed Pre-Key
DH2 = DH(EK_A, IK_B)       # Ephemeral to Identity
DH3 = DH(EK_A, SPK_B)      # Ephemeral to Signed Pre-Key
DH4 = DH(EK_A, OPK_B)      # Ephemeral to One-Time Pre-Key (if available)

SK = KDF(DH1 || DH2 || DH3 || DH4)
```

**Implementation:**
```go
// Initiator side (Alice wants to message Bob)
func InitiateX3DH(
    identityKeyAlice IdentityKeyPair,
    preKeyBundleBob PreKeyBundle,
    solanaSignature []byte,
) (sharedSecret [32]byte, ephemeralKey [32]byte, err error)

// Responder side (Bob receives message from Alice)
func RespondX3DH(
    identityKeyBob IdentityKeyPair,
    signedPreKeyBob PreKey,
    oneTimePreKeyBob *PreKey,
    ephemeralKeyAlice [32]byte,
    identityKeyAlice [32]byte,
) (sharedSecret [32]byte, err error)
```

**Pre-Key Bundle:**
```go
type PreKeyBundle struct {
    UserID          string
    DeviceID        string
    IdentityKey     [32]byte  // X25519 public key (derived from Solana key)
    SignedPreKey    PreKey
    PreKeySignature []byte    // Signed with Solana Ed25519 key
    OneTimePreKey   *PreKey   // Optional
}
```

**Files to create:**
- `pkg/sotalk/protocol/x3dh.go`
- `pkg/sotalk/keys/prekeys.go`

---

## Phase 4: Double Ratchet Algorithm (pkg/sotalk/protocol/double_ratchet.go)

### Double Ratchet Implementation
Provides forward secrecy and break-in recovery.

**Components:**
1. **Symmetric-key ratchet**: HKDF-based key derivation
2. **DH ratchet**: X25519 key exchange

**State Machine:**
```go
type RatchetSession struct {
    RootKey          [32]byte
    SendingChain     ChainState
    ReceivingChain   ChainState
    DHSelf           IdentityKeyPair  // Current DH key pair
    DHRemote         [32]byte         // Remote's current DH public key
    SendMessageNumber    uint32
    ReceiveMessageNumber uint32
    PreviousSendingChain uint32
    SkippedMessageKeys   map[MessageKey][32]byte  // For out-of-order messages
}

type ChainState struct {
    ChainKey      [32]byte
    MessageNumber uint32
}
```

**Core Functions:**
```go
// Initialize ratchet after X3DH
func InitializeRatchet(sharedSecret [32]byte, remoteDHPublicKey [32]byte) (*RatchetSession, error)

// Encrypt message (sender side)
func RatchetEncrypt(session *RatchetSession, plaintext []byte, associatedData []byte) (*EncryptedMessage, error)

// Decrypt message (receiver side)
func RatchetDecrypt(session *RatchetSession, message *EncryptedMessage, associatedData []byte) ([]byte, error)

// Perform DH ratchet step
func DHRatchetStep(session *RatchetSession) error
```

**Key Derivation Chain:**
```
RootKey, ChainKey = KDF(RootKey || DH_output)
MessageKey = KDF(ChainKey)
ChainKey = KDF(ChainKey)
```

**Files to create:**
- `pkg/sotalk/protocol/double_ratchet.go`
- `pkg/sotalk/protocol/session.go`

---

## Phase 5: Sender-Keys for Groups (pkg/sotalk/group/sender_keys.go)

### Sender-Keys Protocol
Efficient encryption for group messages (N members = 1 encryption instead of N).

**How it works:**
1. Each group member has a **Sender Chain**
2. When Alice sends to group: encrypt once with her sender chain key
3. All members can decrypt using Alice's sender chain state

**Types:**
```go
type GroupSession struct {
    GroupID        string
    SenderStates   map[string]*SenderKeyState  // userID -> state
}

type SenderKeyState struct {
    ChainID        string
    Iteration      uint32
    ChainKey       [32]byte
    SigningKey     IdentityKeyPair
    MessageKeys    map[uint32][32]byte  // Cached for out-of-order
}

type GroupEncryptedMessage struct {
    GroupID        string
    SenderID       string
    ChainID        string
    Iteration      uint32
    Ciphertext     []byte
    Signature      []byte  // Signed with sender's Solana key
}
```

**Functions:**
```go
// Create new sender chain for group
func CreateSenderKey(groupID string, senderID string) (*SenderKeyState, error)

// Encrypt message for group
func EncryptGroupMessage(state *SenderKeyState, plaintext []byte) (*GroupEncryptedMessage, error)

// Decrypt group message
func DecryptGroupMessage(state *SenderKeyState, message *GroupEncryptedMessage) ([]byte, error)

// Distribute sender key to group members
func CreateSenderKeyDistributionMessage(state *SenderKeyState, recipientPreKeyBundle PreKeyBundle) (*EncryptedMessage, error)
```

**Files to create:**
- `pkg/sotalk/group/sender_keys.go`
- `pkg/sotalk/group/session.go`

---

## Phase 6: Multi-Device Support (pkg/sotalk/keys/device.go)

### Device Management
Each user can have multiple devices, each with unique keys.

**Device Registration:**
```go
type DeviceManager struct {
    devices map[string]*Device  // deviceID -> Device
}

// Register new device
func (dm *DeviceManager) RegisterDevice(
    deviceID string,
    solanaPublicKey string,
    signedPreKey PreKey,
    oneTimePreKeys []PreKey,
) error

// Get device for encryption
func (dm *DeviceManager) GetDevice(deviceID string) (*Device, error)

// Sync session across devices (via encrypted backup)
func (dm *DeviceManager) SyncSession(session *RatchetSession, deviceID string) error
```

**Device Key Derivation:**
```
DeviceIdentityKey = HKDF(SolanaPrivateKey, "device:" + deviceID)
```

**Files to create:**
- `pkg/sotalk/keys/device.go`

---

## Phase 7: Message Format (pkg/sotalk/message/format.go)

### SoTalk Message Wire Format

**Encrypted Message Structure:**
```
+------------------+
| Version (1 byte) |
+------------------+
| Header           |  - Sender Device ID
|                  |  - Receiver Device ID
|                  |  - Message Counter
|                  |  - Previous Counter
|                  |  - Ratchet Public Key (32 bytes)
+------------------+
| Ephemeral Key    |  32 bytes (X25519 public key)
+------------------+
| Ciphertext       |  Variable length (AES-GCM encrypted)
+------------------+
| Auth Tag         |  16 bytes (GCM tag)
+------------------+
| Solana Signature |  64 bytes (Ed25519 signature)
+------------------+
```

**Serialization:**
```go
func SerializeEncryptedMessage(msg *EncryptedMessage) ([]byte, error)
func DeserializeEncryptedMessage(data []byte) (*EncryptedMessage, error)
```

**Files to create:**
- `pkg/sotalk/message/format.go`
- `pkg/sotalk/message/encrypt.go`
- `pkg/sotalk/message/decrypt.go`

---

## Phase 8: Key Backup (pkg/sotalk/backup/)

### Encrypted Key Backup
Client-side encrypted backup for multi-device sync.

**Backup Format:**
```go
type KeyBackup struct {
    Version       uint8
    UserID        string
    Devices       []DeviceBackup
    Sessions      []SessionBackup
    SenderKeys    []SenderKeyBackup
    Timestamp     time.Time
    EncryptedBlob []byte  // Encrypted with recovery key
}

type DeviceBackup struct {
    DeviceID      string
    IdentityKey   []byte  // Encrypted
    PreKeys       []byte  // Encrypted
}
```

**Recovery Key Derivation:**
```
RecoveryKey = Argon2id(UserPassword, Salt, Iterations)
BackupKey = HKDF(RecoveryKey || SolanaPublicKey)
```

**Functions:**
```go
func ExportKeys(dm *DeviceManager, password string) (*KeyBackup, error)
func ImportKeys(backup *KeyBackup, password string) (*DeviceManager, error)
```

**Files to create:**
- `pkg/sotalk/backup/export.go`
- `pkg/sotalk/backup/import.go`

---

## Phase 9: Main Protocol Interface (pkg/sotalk/protocol.go)

### High-Level Protocol API

```go
package sotalk

// Protocol is the main interface for SoTalk encryption
type Protocol interface {
    // Initialize protocol with Solana keypair
    Initialize(solanaPrivateKey []byte, deviceID string) error

    // Register device and generate pre-keys
    RegisterDevice() (*PreKeyBundle, error)

    // Encrypt 1-to-1 message
    EncryptMessage(recipientID string, plaintext []byte) (*EncryptedMessage, error)

    // Decrypt 1-to-1 message
    DecryptMessage(senderID string, encrypted *EncryptedMessage) ([]byte, error)

    // Create group session
    CreateGroupSession(groupID string, memberIDs []string) error

    // Encrypt group message
    EncryptGroupMessage(groupID string, plaintext []byte) (*GroupEncryptedMessage, error)

    // Decrypt group message
    DecryptGroupMessage(groupID string, encrypted *GroupEncryptedMessage) ([]byte, error)

    // Backup keys
    ExportBackup(password string) ([]byte, error)
    ImportBackup(backupData []byte, password string) error
}

// NewProtocol creates a new SoTalk protocol instance
func NewProtocol(config Config) (Protocol, error)
```

**Files to create:**
- `pkg/sotalk/protocol.go`
- `pkg/sotalk/types.go`
- `pkg/sotalk/errors.go`

---

## Phase 10: Backend Integration

### Changes to Existing Code

#### 1. Domain Layer Changes (`internal/domain/message/entity.go`)

**Before:**
```go
type Message struct {
    Content     string  // Plain text
    Signature   string
}
```

**After:**
```go
type Message struct {
    Content           []byte  // Encrypted content (SoTalk format)
    EncryptionVersion uint8   // SoTalk protocol version
    SenderDeviceID    string  // Device that sent the message
    IsEncrypted       bool    // Flag for backward compatibility
    Signature         string  // Solana signature (part of SoTalk message)
}
```

#### 2. New Domain: Encryption Keys (`internal/domain/encryption/`)

**Files to create:**
```
internal/domain/encryption/
├── entity.go          # Key entities (IdentityKey, PreKey, Device)
├── repository.go      # Repository interface
└── errors.go
```

**Key Entities:**
```go
// IdentityKey stores user's X25519 identity key (derived from Solana key)
type IdentityKey struct {
    UserID     uuid.UUID
    DeviceID   string
    PublicKey  []byte  // X25519 public key
    SignedBy   string  // Signed with Solana key for verification
    CreatedAt  time.Time
}

// PreKey stores pre-keys for X3DH
type PreKey struct {
    ID         uint32
    UserID     uuid.UUID
    DeviceID   string
    KeyID      uint32
    PublicKey  []byte
    Signature  []byte
    Used       bool
    CreatedAt  time.Time
}

// DeviceInfo stores device information
type Device struct {
    ID              string
    UserID          uuid.UUID
    Name            string
    IdentityKeyID   uuid.UUID
    LastActive      time.Time
    CreatedAt       time.Time
}

// Session stores Double Ratchet session state
type Session struct {
    ID              uuid.UUID
    UserID          uuid.UUID
    DeviceID        string
    PeerUserID      uuid.UUID
    PeerDeviceID    string
    SessionState    []byte  // Serialized RatchetSession
    LastUsed        time.Time
    CreatedAt       time.Time
}

// SenderKeyState stores group sender keys
type GroupSenderKey struct {
    ID           uuid.UUID
    GroupID      uuid.UUID
    SenderID     uuid.UUID
    ChainID      string
    KeyState     []byte  // Serialized SenderKeyState
    CreatedAt    time.Time
}
```

#### 3. Repository Implementation (`internal/repository/postgres/`)

**Files to create:**
```
internal/repository/postgres/
├── encryption_repository.go   # Key storage implementation
└── session_repository.go      # Session storage
```

**Database Schema (migrations/):**
```sql
-- migrations/000015_create_encryption_keys.up.sql

CREATE TABLE identity_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    public_key BYTEA NOT NULL,
    signed_by TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, device_id)
);

CREATE TABLE pre_keys (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    key_id INT NOT NULL,
    public_key BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, device_id, key_id)
);

CREATE INDEX idx_prekeys_unused ON pre_keys(user_id, device_id, used) WHERE used = FALSE;

CREATE TABLE devices (
    id VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255),
    identity_key_id UUID REFERENCES identity_keys(id),
    last_active TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE ratchet_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    peer_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    peer_device_id VARCHAR(255) NOT NULL,
    session_state BYTEA NOT NULL,
    last_used TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, device_id, peer_user_id, peer_device_id)
);

CREATE TABLE group_sender_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chain_id VARCHAR(255) NOT NULL,
    key_state BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(group_id, sender_id, chain_id)
);
```

#### 4. Use Case Layer (`internal/usecase/encryption/`)

**Files to create:**
```
internal/usecase/encryption/
├── service.go
├── interface.go
└── dto/
    └── encryption_dto.go
```

**Service Interface:**
```go
type Service interface {
    // Device management
    RegisterDevice(ctx context.Context, userID uuid.UUID, deviceID string, solanaKey []byte) (*dto.DeviceInfo, error)
    GetDevices(ctx context.Context, userID uuid.UUID) ([]*dto.DeviceInfo, error)

    // Pre-key management
    GeneratePreKeys(ctx context.Context, userID uuid.UUID, deviceID string, count int) error
    GetPreKeyBundle(ctx context.Context, userID uuid.UUID, deviceID string) (*dto.PreKeyBundle, error)

    // Session management
    InitiateSession(ctx context.Context, initiatorUserID uuid.UUID, recipientUserID uuid.UUID, deviceID string) error
    GetSession(ctx context.Context, userID uuid.UUID, peerUserID uuid.UUID, deviceID string) (*dto.Session, error)
    UpdateSession(ctx context.Context, session *dto.Session) error

    // Group encryption
    InitiateGroupSession(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
    DistributeSenderKey(ctx context.Context, groupID uuid.UUID, senderID uuid.UUID, recipientIDs []uuid.UUID) error
}
```

#### 5. Message Service Integration (`internal/usecase/message/service.go`)

**Modify existing message service:**
```go
type service struct {
    messageRepo       message.Repository
    conversationRepo  conversation.Repository
    userRepo          user.Repository
    encryptionService encryption.Service  // NEW
    sotalkProtocol    sotalk.Protocol     // NEW
    broadcaster       websocket.Broadcaster
}

// SendMessage (modified)
func (s *service) SendMessage(ctx context.Context, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
    // 1. Validate sender
    // 2. Check if recipient exists
    // 3. Initialize session if first message
    // 4. Encrypt message using SoTalk Protocol
    encryptedContent, err := s.sotalkProtocol.EncryptMessage(req.RecipientID, []byte(req.Content))
    if err != nil {
        return nil, err
    }

    // 5. Create message with encrypted content
    msg := message.NewMessage(
        req.ConversationID,
        req.SenderID,
        encryptedContent.Serialize(),  // []byte
        message.ContentTypeText,
    )
    msg.IsEncrypted = true
    msg.SenderDeviceID = req.DeviceID

    // 6. Save to database
    // 7. Broadcast via WebSocket
    // 8. Return response
}
```

#### 6. HTTP Handlers (`internal/delivery/http/handler/`)

**New handler: `encryption_handler.go`**
```go
type EncryptionHandler struct {
    encryptionService encryption.Service
}

// POST /api/v1/encryption/devices/register
func (h *EncryptionHandler) RegisterDevice(c *gin.Context)

// GET /api/v1/encryption/devices
func (h *EncryptionHandler) GetDevices(c *gin.Context)

// POST /api/v1/encryption/prekeys/generate
func (h *EncryptionHandler) GeneratePreKeys(c *gin.Context)

// GET /api/v1/encryption/keys/:userId/:deviceId
func (h *EncryptionHandler) GetPreKeyBundle(c *gin.Context)

// POST /api/v1/encryption/backup/export
func (h *EncryptionHandler) ExportBackup(c *gin.Context)

// POST /api/v1/encryption/backup/import
func (h *EncryptionHandler) ImportBackup(c *gin.Context)
```

#### 7. Main Application (`cmd/api/main.go`)

**Add initialization:**
```go
// Initialize SoTalk Protocol
sotalkProtocol, err := sotalk.NewProtocol(sotalk.Config{
    DeviceID: "server-device-1",  // Server device ID
})
if err != nil {
    logger.Fatal("Failed to initialize SoTalk Protocol", zap.Error(err))
}
logger.Info("✅ SoTalk Protocol initialized")

// Initialize encryption repository
encryptionRepo := postgres.NewEncryptionRepository(db)
sessionRepo := postgres.NewSessionRepository(db)

// Initialize encryption service
encryptionService := encryption.NewService(
    encryptionRepo,
    sessionRepo,
    userRepo,
    sotalkProtocol,
)
logger.Info("✅ Encryption service initialized")

// Update message service with encryption
messageService := message.NewService(
    messageRepo,
    conversationRepo,
    userRepo,
    encryptionService,  // NEW
    sotalkProtocol,     // NEW
    wsBroadcaster,
)
```

---

## Implementation Order

### Step 1: Create SoTalk Protocol Package (Standalone)
**Goal:** Build and test the protocol independently

1. Create directory structure: `pkg/sotalk/`
2. Implement cryptographic primitives: `protocol/primitives.go`
3. Implement X3DH: `protocol/x3dh.go`
4. Implement Double Ratchet: `protocol/double_ratchet.go`
5. Implement Sender-Keys: `group/sender_keys.go`
6. Implement device management: `keys/device.go`
7. Implement message format: `message/format.go`
8. Create main protocol interface: `protocol.go`
9. Write comprehensive tests

**Testing:**
- Unit tests for each component
- Integration tests for full message flow
- Benchmark tests for performance

### Step 2: Backend Integration (Domain Layer)
1. Create encryption domain entities: `internal/domain/encryption/entity.go`
2. Define repository interfaces: `internal/domain/encryption/repository.go`
3. Modify message entity: `internal/domain/message/entity.go`

### Step 3: Backend Integration (Repository Layer)
1. Create database migrations: `migrations/000015_create_encryption_keys.up.sql`
2. Implement encryption repository: `internal/repository/postgres/encryption_repository.go`
3. Implement session repository: `internal/repository/postgres/session_repository.go`

### Step 4: Backend Integration (Use Case Layer)
1. Create encryption service: `internal/usecase/encryption/service.go`
2. Modify message service: `internal/usecase/message/service.go`
3. Create DTOs: `internal/usecase/dto/encryption_dto.go`

### Step 5: Backend Integration (Delivery Layer)
1. Create encryption handler: `internal/delivery/http/handler/encryption_handler.go`
2. Add routes to router: `internal/delivery/http/router.go`
3. Update WebSocket handler for encrypted messages

### Step 6: Testing & Documentation
1. End-to-end tests for encrypted messaging
2. API documentation (OpenAPI/Swagger)
3. Protocol specification document
4. Client SDK guide

---

## Critical Files to Create

### New Files (70+ files)
```
pkg/sotalk/protocol.go
pkg/sotalk/types.go
pkg/sotalk/errors.go
pkg/sotalk/protocol/primitives.go
pkg/sotalk/protocol/x3dh.go
pkg/sotalk/protocol/double_ratchet.go
pkg/sotalk/protocol/sender_keys.go
pkg/sotalk/protocol/session.go
pkg/sotalk/keys/identity.go
pkg/sotalk/keys/prekeys.go
pkg/sotalk/keys/device.go
pkg/sotalk/keys/storage.go
pkg/sotalk/message/encrypt.go
pkg/sotalk/message/decrypt.go
pkg/sotalk/message/format.go
pkg/sotalk/group/sender_keys.go
pkg/sotalk/group/session.go
pkg/sotalk/backup/export.go
pkg/sotalk/backup/import.go
internal/domain/encryption/entity.go
internal/domain/encryption/repository.go
internal/repository/postgres/encryption_repository.go
internal/repository/postgres/session_repository.go
internal/usecase/encryption/service.go
internal/usecase/encryption/interface.go
internal/usecase/dto/encryption_dto.go
internal/delivery/http/handler/encryption_handler.go
migrations/000015_create_encryption_keys.up.sql
migrations/000015_create_encryption_keys.down.sql
```

### Files to Modify (8 files)
```
internal/domain/message/entity.go (add encryption fields)
internal/usecase/message/service.go (integrate encryption)
internal/delivery/http/router.go (add encryption routes)
internal/delivery/websocket/handler.go (handle encrypted messages)
cmd/api/main.go (initialize protocol)
internal/repository/postgres/models.go (update Message model)
go.mod (add crypto dependencies if needed)
```

---

## Security Considerations

### Protocol Security
1. **Forward Secrecy**: Double Ratchet ensures past messages remain secure
2. **Break-in Recovery**: New DH exchange after compromise
3. **Message Authentication**: Every message signed with Solana key
4. **Replay Protection**: Message counters prevent replay attacks
5. **Out-of-Order Handling**: Skipped message keys cached
6. **Group Security**: Sender-keys provide efficient group encryption

### Implementation Security
1. **Key Storage**: Keys never stored in plain text
2. **Memory Security**: Zeroing sensitive data after use
3. **Timing Attacks**: Constant-time comparisons
4. **Random Number Generation**: crypto/rand for all randomness
5. **Key Rotation**: Automatic rotation with Double Ratchet
6. **Device Verification**: Solana signatures verify device authenticity

### Client-Side Security
1. **Private Keys**: Never leave client devices
2. **Backup Encryption**: Strong password-based encryption (Argon2id)
3. **Session Storage**: Encrypted session storage in browser/app
4. **Key Derivation**: Proper KDF for all derived keys

---

## API Endpoints (New)

```
POST   /api/v1/encryption/devices/register          # Register device
GET    /api/v1/encryption/devices                   # Get user's devices
POST   /api/v1/encryption/prekeys/generate          # Generate pre-keys
GET    /api/v1/encryption/keys/:userId/:deviceId    # Get pre-key bundle
POST   /api/v1/encryption/backup/export             # Export encrypted backup
POST   /api/v1/encryption/backup/import             # Import backup
GET    /api/v1/encryption/sessions                  # Get active sessions
DELETE /api/v1/encryption/sessions/:sessionId       # Delete session
```

---

## Testing Strategy

### Unit Tests
- Test each cryptographic primitive
- Test X3DH key exchange
- Test Double Ratchet state transitions
- Test message serialization/deserialization

### Integration Tests
- Full message encryption/decryption flow
- Multi-device session management
- Group message encryption
- Key backup/restore

### Performance Tests
- Encryption/decryption speed
- Memory usage
- Session initialization time
- Group message scaling (100+ members)

### Security Tests
- Cryptographic correctness
- Replay attack prevention
- Forward secrecy verification
- Signature verification

---

## Timeline Estimate

**Phase 1 (SoTalk Package):** 5-7 days
- Primitives: 1 day
- X3DH: 1 day
- Double Ratchet: 2 days
- Sender-Keys: 1 day
- Testing: 2 days

**Phase 2 (Backend Integration):** 3-4 days
- Domain layer: 1 day
- Repository layer: 1 day
- Use case layer: 1 day
- Delivery layer: 1 day

**Total:** 8-11 days for full implementation

---

## Success Criteria

✅ SoTalk Protocol package is standalone and reusable
✅ Full Double Ratchet with forward secrecy working
✅ Multi-device support with device management
✅ Sender-keys protocol for groups (efficient encryption)
✅ Client-side only key storage (maximum security)
✅ Encrypted key backup/restore functionality
✅ All messages encrypted end-to-end
✅ Backward compatible (existing plain messages still work)
✅ Comprehensive test coverage (>80%)
✅ API documentation complete
✅ Protocol specification document written

---

## Next Steps After Approval

1. Create `pkg/sotalk/` directory structure
2. Implement cryptographic primitives first
3. Build X3DH key exchange
4. Implement Double Ratchet
5. Add Sender-Keys for groups
6. Integrate with backend
7. Test thoroughly
8. Document everything

Ready to build the most secure messaging protocol on Solana! 🔐🚀
