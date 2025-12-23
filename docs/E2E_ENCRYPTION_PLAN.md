# SoTalk E2E Encryption Protocol Plan

## Overview
This document outlines the implementation plan for End-to-End Encryption (E2E) in the SoTalk messaging platform. The encryption protocol is inspired by the Signal Protocol, using X25519 for key exchange, Double Ratchet algorithm for forward secrecy, and AES-256-GCM for message encryption.

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Cryptographic Primitives](#cryptographic-primitives)
3. [Key Management](#key-management)
4. [Session Establishment](#session-establishment)
5. [Message Encryption/Decryption](#message-encryptiondecryption)
6. [Double Ratchet Algorithm](#double-ratchet-algorithm)
7. [Database Schema](#database-schema)
8. [API Endpoints](#api-endpoints)
9. [Package Structure](#package-structure)
10. [Implementation Steps](#implementation-steps)
11. [Security Considerations](#security-considerations)
12. [Testing Strategy](#testing-strategy)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      E2E Encryption Flow                        │
└─────────────────────────────────────────────────────────────────┘

User A (Sender)                                    User B (Receiver)
     │                                                    │
     ├─ Generate Identity Key (long-term)                ├─ Generate Identity Key
     ├─ Generate Signed Pre-Key                          ├─ Generate Signed Pre-Key
     ├─ Generate One-Time Pre-Keys (bundle)              ├─ Generate One-Time Pre-Keys
     │                                                    │
     │  Upload Public Keys to Server                     │  Upload Public Keys to Server
     ├────────────────────────────────────────────────────>
     │                                                    │
     │  Request Key Bundle for User B                    │
     ├────────────────────────────────────────────────────>
     │  <────────────────────────────────────────────────┤
     │  Receive Key Bundle (Identity, Signed PreKey,     │
     │  OneTime PreKey)                                  │
     │                                                    │
     ├─ X3DH Key Agreement (establish shared secret)     │
     ├─ Initialize Double Ratchet                        │
     ├─ Derive Chain Keys & Message Keys                 │
     │                                                    │
     ├─ Encrypt Message with Message Key                 │
     │  (AES-256-GCM)                                    │
     │                                                    │
     │  Send Encrypted Message + Header                  │
     ├────────────────────────────────────────────────────>
     │                                                    ├─ Receive Encrypted Message
     │                                                    ├─ Initialize/Update Ratchet
     │                                                    ├─ Derive Chain Keys & Message Keys
     │                                                    ├─ Decrypt Message
     │                                                    │
     │  <────────────────────────────────────────────────┤
     │  Receive Reply (Ratchet Step)                     │
     │                                                    │
     ├─ Continue Double Ratchet (forward secrecy)        ├─ Continue Double Ratchet
     │                                                    │
```

---

## Cryptographic Primitives

### 1. Key Exchange
- **Algorithm**: X25519 (Elliptic Curve Diffie-Hellman)
- **Purpose**: Establish shared secrets between parties
- **Library**: `golang.org/x/crypto/curve25519`

### 2. Key Derivation
- **Algorithm**: HKDF-SHA256
- **Purpose**: Derive encryption keys from shared secrets
- **Library**: `golang.org/x/crypto/hkdf`

### 3. Symmetric Encryption
- **Algorithm**: AES-256-GCM (Authenticated Encryption)
- **Purpose**: Encrypt message content
- **Key Size**: 256 bits (32 bytes)
- **Nonce Size**: 96 bits (12 bytes)
- **Library**: `crypto/aes` + `crypto/cipher`

### 4. Message Authentication
- **Algorithm**: HMAC-SHA256
- **Purpose**: Authenticate message integrity
- **Library**: `crypto/hmac`

### 5. Random Generation
- **Source**: `crypto/rand`
- **Purpose**: Generate cryptographically secure random keys

---

## Key Management

### Key Types

#### 1. Identity Key Pair (IK)
- **Lifetime**: Long-term (permanent)
- **Storage**: Client-side (private), Server-side (public)
- **Purpose**: User's root identity for E2E encryption
- **Generation**: Once per user, signed by Solana wallet
- **Format**: X25519 key pair (32 bytes each)

```go
type IdentityKey struct {
    UserID      uuid.UUID
    PublicKey   []byte  // 32 bytes, stored on server
    PrivateKey  []byte  // 32 bytes, NEVER sent to server
    SignedBy    string  // Solana signature for verification
    CreatedAt   time.Time
}
```

#### 2. Signed Pre-Key (SPK)
- **Lifetime**: Medium-term (rotated weekly/monthly)
- **Storage**: Client-side (private), Server-side (public)
- **Purpose**: Provide forward secrecy in asynchronous messaging
- **Signature**: Signed by Identity Key
- **Format**: X25519 key pair + signature

```go
type SignedPreKey struct {
    ID          uint32
    UserID      uuid.UUID
    PublicKey   []byte    // 32 bytes
    PrivateKey  []byte    // 32 bytes, client-side only
    Signature   []byte    // Signature by Identity Key
    CreatedAt   time.Time
    ExpiresAt   time.Time
}
```

#### 3. One-Time Pre-Keys (OPK)
- **Lifetime**: Single use
- **Storage**: Client-side (private), Server-side (public)
- **Purpose**: Additional forward secrecy, consumed on first use
- **Quantity**: Generate 100 at a time, replenish when low
- **Format**: X25519 key pair

```go
type OneTimePreKey struct {
    ID          uint32
    UserID      uuid.UUID
    PublicKey   []byte  // 32 bytes
    PrivateKey  []byte  // 32 bytes, client-side only
    Used        bool
    UsedAt      *time.Time
    CreatedAt   time.Time
}
```

#### 4. Ratchet Keys
- **DH Ratchet Key**: Rotated with each message exchange
- **Chain Keys**: Derive message keys (KDF ratchet)
- **Message Keys**: One-time keys for each message

```go
type RatchetState struct {
    SessionID       uuid.UUID
    DHPrivateKey    []byte  // Current DH ratchet private key
    DHPublicKey     []byte  // Current DH ratchet public key
    DHRemotePublic  []byte  // Peer's DH ratchet public key
    RootKey         []byte  // Root key (32 bytes)
    SendChainKey    []byte  // Sending chain key
    ReceiveChainKey []byte  // Receiving chain key
    SendCounter     uint32  // Message counter for sending
    ReceiveCounter  uint32  // Message counter for receiving
    PrevCounter     uint32  // Previous sending chain length
}
```

---

## Session Establishment

### X3DH (Extended Triple Diffie-Hellman) Protocol

When User A wants to send a message to User B for the first time:

#### Step 1: User B Uploads Key Bundle to Server
```
Key Bundle = {
    Identity Public Key (IK_B),
    Signed Pre-Key Public Key (SPK_B),
    Signed Pre-Key Signature,
    One-Time Pre-Key Public Key (OPK_B) [optional]
}
```

#### Step 2: User A Fetches User B's Key Bundle
```http
GET /api/v1/keys/{userB}/bundle
Response: Key Bundle
```

#### Step 3: User A Performs X3DH Key Agreement

Generate ephemeral key pair:
```
EK_A = GenerateEphemeralKeyPair()
```

Perform 4 Diffie-Hellman operations:
```
DH1 = ECDH(IK_A, SPK_B)
DH2 = ECDH(EK_A, IK_B)
DH3 = ECDH(EK_A, SPK_B)
DH4 = ECDH(EK_A, OPK_B)  // If OPK available
```

Derive shared secret:
```
If OPK available:
    SK = KDF(DH1 || DH2 || DH3 || DH4)
Else:
    SK = KDF(DH1 || DH2 || DH3)
```

#### Step 4: Initialize Double Ratchet
```
Root Key (RK) = HKDF(SK, "RootKey", 32)
Chain Key (CK) = HKDF(SK, "ChainKey", 32)
```

#### Step 5: Send Initial Message
```json
{
    "header": {
        "identityKey": "IK_A",
        "ephemeralKey": "EK_A",
        "preKeyID": 123,
        "oneTimePreKeyID": 456,
        "counter": 0
    },
    "ciphertext": "encrypted_message"
}
```

---

## Message Encryption/Decryption

### Encryption Process

```go
func EncryptMessage(plaintext []byte, session *Session) (*EncryptedMessage, error) {
    // 1. Advance the ratchet if needed
    session.AdvanceRatchet()

    // 2. Derive message key from chain key
    messageKey = DeriveMessageKey(session.SendChainKey)

    // 3. Advance chain key
    session.SendChainKey = DeriveNextChainKey(session.SendChainKey)

    // 4. Encrypt with AES-256-GCM
    ciphertext, nonce = AES_GCM_Encrypt(messageKey, plaintext)

    // 5. Increment counter
    session.SendCounter++

    // 6. Create message header
    header = {
        dhPublicKey: session.DHPublicKey,
        prevCounter: session.PrevCounter,
        counter: session.SendCounter,
    }

    return EncryptedMessage{
        Header: header,
        Ciphertext: ciphertext,
        Nonce: nonce,
    }
}
```

### Decryption Process

```go
func DecryptMessage(encrypted *EncryptedMessage, session *Session) ([]byte, error) {
    // 1. Check if DH ratchet needs update
    if encrypted.Header.DHPublicKey != session.DHRemotePublic {
        session.DHRatchetStep(encrypted.Header.DHPublicKey)
    }

    // 2. Derive message key
    messageKey = DeriveMessageKey(session.ReceiveChainKey, encrypted.Header.Counter)

    // 3. Decrypt with AES-256-GCM
    plaintext = AES_GCM_Decrypt(messageKey, encrypted.Ciphertext, encrypted.Nonce)

    // 4. Advance receive chain
    session.ReceiveChainKey = DeriveNextChainKey(session.ReceiveChainKey)
    session.ReceiveCounter++

    return plaintext, nil
}
```

### Encrypted Message Format

```go
type EncryptedMessage struct {
    Header struct {
        DHPublicKey  []byte  // Current DH ratchet public key
        PrevCounter  uint32  // Previous chain length
        Counter      uint32  // Message number in current chain
    }
    Ciphertext []byte  // AES-256-GCM encrypted message
    Nonce      []byte  // 12-byte nonce for AES-GCM
    Tag        []byte  // 16-byte authentication tag
}
```

---

## Double Ratchet Algorithm

The Double Ratchet algorithm provides forward secrecy and future secrecy through two ratcheting mechanisms:

### 1. Symmetric-Key Ratchet (KDF Ratchet)

Updates chain keys to derive message keys:

```
Message Key = HKDF(Chain Key, "MessageKey", 32)
Next Chain Key = HKDF(Chain Key, "ChainKey", 32)
```

**Properties**:
- Each message gets a unique encryption key
- Old message keys cannot be derived from new chain keys (forward secrecy)

### 2. Diffie-Hellman Ratchet (DH Ratchet)

Updates DH key pairs and root key with each message exchange:

```
# When receiving a new DH public key from peer:
Shared Secret = ECDH(My DH Private Key, Peer's DH Public Key)
Root Key, Chain Key = HKDF(Root Key, Shared Secret, "RatchetKeys", 64)

# Generate new DH key pair for next send:
New DH Key Pair = GenerateKeyPair()
```

**Properties**:
- Root key is ratcheted with each DH exchange
- Provides break-in recovery (future secrecy)

### Ratchet State Machine

```
Initial State:
    - Root Key (RK) from X3DH
    - DH Key Pair (Alice)
    - Peer DH Public Key (Bob)

Sending a Message:
    1. If no send chain, perform DH ratchet step
    2. Derive message key from send chain key
    3. Encrypt message with message key
    4. Advance send chain key
    5. Increment send counter

Receiving a Message:
    1. If DH public key changed, perform DH ratchet step
    2. Derive message key from receive chain key
    3. Decrypt message with message key
    4. Advance receive chain key
    5. Increment receive counter

DH Ratchet Step:
    1. Compute DH shared secret with peer's new public key
    2. Derive new root key and receive chain key
    3. Generate new DH key pair
    4. Compute DH shared secret with new key pair
    5. Derive new root key and send chain key
    6. Reset counters
```

---

## Database Schema

### 1. Identity Keys Table

```sql
CREATE TABLE identity_keys (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    public_key BYTEA NOT NULL,
    signed_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(public_key)
);

CREATE INDEX idx_identity_keys_user ON identity_keys(user_id);
```

### 2. Signed Pre-Keys Table

```sql
CREATE TABLE signed_pre_keys (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_id INTEGER NOT NULL,
    public_key BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(user_id, key_id)
);

CREATE INDEX idx_signed_pre_keys_user ON signed_pre_keys(user_id);
CREATE INDEX idx_signed_pre_keys_active ON signed_pre_keys(user_id, is_active);
```

### 3. One-Time Pre-Keys Table

```sql
CREATE TABLE one_time_pre_keys (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_id INTEGER NOT NULL,
    public_key BYTEA NOT NULL,
    used BOOLEAN DEFAULT false,
    used_at TIMESTAMP,
    used_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, key_id)
);

CREATE INDEX idx_one_time_pre_keys_user ON one_time_pre_keys(user_id);
CREATE INDEX idx_one_time_pre_keys_unused ON one_time_pre_keys(user_id, used) WHERE used = false;
```

### 4. Sessions Table

```sql
CREATE TABLE encryption_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    peer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    -- Session state (encrypted with user's key)
    session_state BYTEA NOT NULL,

    -- Ratchet counters (for tracking message order)
    send_counter INTEGER DEFAULT 0,
    receive_counter INTEGER DEFAULT 0,
    prev_counter INTEGER DEFAULT 0,

    -- Metadata
    initiated_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, peer_id, conversation_id)
);

CREATE INDEX idx_sessions_user ON encryption_sessions(user_id);
CREATE INDEX idx_sessions_peer ON encryption_sessions(peer_id);
CREATE INDEX idx_sessions_conversation ON encryption_sessions(conversation_id);
CREATE INDEX idx_sessions_updated ON encryption_sessions(updated_at);
```

### 5. Skipped Message Keys Table

For handling out-of-order messages:

```sql
CREATE TABLE skipped_message_keys (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES encryption_sessions(id) ON DELETE CASCADE,

    -- Message identification
    dh_public_key BYTEA NOT NULL,
    message_number INTEGER NOT NULL,

    -- Derived message key
    message_key BYTEA NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,

    UNIQUE(session_id, dh_public_key, message_number)
);

CREATE INDEX idx_skipped_keys_session ON skipped_message_keys(session_id);
CREATE INDEX idx_skipped_keys_expires ON skipped_message_keys(expires_at);

-- Cleanup old skipped keys (run periodically)
DELETE FROM skipped_message_keys WHERE expires_at < NOW();
```

---

## API Endpoints

### 1. Upload Identity Key

```http
POST /api/v1/keys/identity
Authorization: Bearer {jwt_token}

Request:
{
    "publicKey": "base64_encoded_public_key",
    "signedBy": "solana_signature"
}

Response:
{
    "success": true,
    "message": "Identity key uploaded successfully"
}
```

### 2. Upload Pre-Keys

```http
POST /api/v1/keys/prekeys
Authorization: Bearer {jwt_token}

Request:
{
    "signedPreKey": {
        "keyId": 1,
        "publicKey": "base64_encoded_public_key",
        "signature": "base64_encoded_signature"
    },
    "oneTimePreKeys": [
        {
            "keyId": 1,
            "publicKey": "base64_encoded_public_key"
        },
        // ... up to 100 keys
    ]
}

Response:
{
    "success": true,
    "message": "Pre-keys uploaded successfully",
    "signedPreKeyId": 1,
    "oneTimePreKeyCount": 100
}
```

### 3. Get Key Bundle

```http
GET /api/v1/keys/:userId/bundle
Authorization: Bearer {jwt_token}

Response:
{
    "userId": "uuid",
    "identityKey": "base64_encoded_public_key",
    "signedPreKey": {
        "keyId": 1,
        "publicKey": "base64_encoded_public_key",
        "signature": "base64_encoded_signature"
    },
    "oneTimePreKey": {
        "keyId": 42,
        "publicKey": "base64_encoded_public_key"
    }
}
```

### 4. Get Pre-Key Count

```http
GET /api/v1/keys/prekeys/count
Authorization: Bearer {jwt_token}

Response:
{
    "oneTimePreKeyCount": 23
}
```

### 5. Initialize Session

```http
POST /api/v1/sessions/init
Authorization: Bearer {jwt_token}

Request:
{
    "peerId": "uuid",
    "conversationId": "uuid",
    "initialMessage": {
        "header": {
            "identityKey": "base64",
            "ephemeralKey": "base64",
            "signedPreKeyId": 1,
            "oneTimePreKeyId": 42,
            "counter": 0
        },
        "ciphertext": "base64_encrypted_content"
    }
}

Response:
{
    "success": true,
    "sessionId": "uuid"
}
```

### 6. Get Session

```http
GET /api/v1/sessions/:sessionId
Authorization: Bearer {jwt_token}

Response:
{
    "sessionId": "uuid",
    "peerId": "uuid",
    "conversationId": "uuid",
    "sendCounter": 5,
    "receiveCounter": 3,
    "createdAt": "2024-01-01T00:00:00Z",
    "lastUsedAt": "2024-01-01T12:00:00Z"
}
```

---

## Package Structure

```
sotalk/
├── pkg/
│   └── crypto/
│       ├── keys.go              # Key generation and management
│       ├── kdf.go               # Key derivation functions (HKDF)
│       ├── x3dh.go              # X3DH key agreement protocol
│       ├── ratchet.go           # Double Ratchet implementation
│       ├── encrypt.go           # Message encryption/decryption
│       ├── session.go           # Session management
│       ├── bundle.go            # Key bundle encoding/decoding
│       ├── signature.go         # Signature verification
│       └── types.go             # Common types and interfaces
│
├── internal/
│   ├── domain/
│   │   └── encryption/
│   │       ├── entity.go        # Domain entities (IdentityKey, Session, etc.)
│   │       ├── repository.go    # Repository interfaces
│   │       └── errors.go        # Domain-specific errors
│   │
│   ├── usecase/
│   │   └── encryption/
│   │       ├── service.go       # Encryption use cases
│   │       └── interface.go     # Service interface
│   │
│   ├── repository/
│   │   └── postgres/
│   │       ├── encryption_repository.go  # Implementation
│   │       └── models.go         # Add encryption models
│   │
│   └── delivery/
│       └── http/
│           └── handler/
│               └── encryption_handler.go  # HTTP handlers
│
└── migrations/
    ├── 000011_create_identity_keys.up.sql
    ├── 000011_create_identity_keys.down.sql
    ├── 000012_create_signed_pre_keys.up.sql
    ├── 000012_create_signed_pre_keys.down.sql
    ├── 000013_create_one_time_pre_keys.up.sql
    ├── 000013_create_one_time_pre_keys.down.sql
    ├── 000014_create_encryption_sessions.up.sql
    ├── 000014_create_encryption_sessions.down.sql
    ├── 000015_create_skipped_message_keys.up.sql
    └── 000015_create_skipped_message_keys.down.sql
```

---

## Implementation Steps

### Phase 1: Core Cryptography (pkg/crypto)

#### Step 1.1: Key Generation and Management
- [ ] Implement X25519 key pair generation
- [ ] Create identity key pair functions
- [ ] Create signed pre-key generation with signing
- [ ] Create one-time pre-key generation (batch)
- [ ] Implement key encoding/decoding (base64)

#### Step 1.2: Key Derivation
- [ ] Implement HKDF-SHA256
- [ ] Create chain key derivation
- [ ] Create message key derivation
- [ ] Create root key derivation

#### Step 1.3: X3DH Protocol
- [ ] Implement X3DH key agreement (4-way DH)
- [ ] Handle cases with/without one-time pre-keys
- [ ] Derive initial shared secret
- [ ] Initialize session from shared secret

#### Step 1.4: Double Ratchet
- [ ] Implement DH ratchet step
- [ ] Implement symmetric key ratchet
- [ ] Handle ratchet state transitions
- [ ] Manage skipped message keys (out-of-order)

#### Step 1.5: Encryption/Decryption
- [ ] Implement AES-256-GCM encryption
- [ ] Implement AES-256-GCM decryption
- [ ] Create message header serialization
- [ ] Handle authentication tags

### Phase 2: Domain Layer

#### Step 2.1: Domain Entities
- [ ] Create IdentityKey entity
- [ ] Create SignedPreKey entity
- [ ] Create OneTimePreKey entity
- [ ] Create Session entity
- [ ] Create SkippedMessageKey entity

#### Step 2.2: Repository Interfaces
- [ ] Define IdentityKeyRepository interface
- [ ] Define PreKeyRepository interface
- [ ] Define SessionRepository interface
- [ ] Define SkippedMessageKeyRepository interface

### Phase 3: Infrastructure Layer

#### Step 3.1: Database Models
- [ ] Create GORM models for identity keys
- [ ] Create GORM models for pre-keys
- [ ] Create GORM models for sessions
- [ ] Create GORM models for skipped keys

#### Step 3.2: Database Migrations
- [ ] Create migration for identity_keys table
- [ ] Create migration for signed_pre_keys table
- [ ] Create migration for one_time_pre_keys table
- [ ] Create migration for encryption_sessions table
- [ ] Create migration for skipped_message_keys table

#### Step 3.3: Repository Implementation
- [ ] Implement IdentityKeyRepository (GORM)
- [ ] Implement PreKeyRepository (GORM)
- [ ] Implement SessionRepository (GORM)
- [ ] Implement SkippedMessageKeyRepository (GORM)

### Phase 4: Use Case Layer

#### Step 4.1: Encryption Service
- [ ] Create GenerateAndUploadKeys use case
- [ ] Create GetKeyBundle use case
- [ ] Create InitializeSession use case
- [ ] Create EncryptMessage use case
- [ ] Create DecryptMessage use case

### Phase 5: Delivery Layer

#### Step 5.1: HTTP Handlers
- [ ] Implement POST /keys/identity
- [ ] Implement POST /keys/prekeys
- [ ] Implement GET /keys/:userId/bundle
- [ ] Implement GET /keys/prekeys/count
- [ ] Implement POST /sessions/init
- [ ] Implement GET /sessions/:sessionId

### Phase 6: Integration

#### Step 6.1: Message Service Integration
- [ ] Integrate encryption into message sending
- [ ] Integrate decryption into message receiving
- [ ] Update Message entity to support encrypted content
- [ ] Update WebSocket to handle encrypted messages

#### Step 6.2: Client-Side Integration (Future)
- [ ] Document client-side key generation
- [ ] Document client-side encryption flow
- [ ] Create JavaScript/TypeScript SDK
- [ ] Create React hooks for encryption

---

## Security Considerations

### 1. Key Storage

**Server-Side**:
- Store ONLY public keys
- NEVER store private keys
- Use database encryption at rest
- Implement access controls

**Client-Side**:
- Store private keys in secure storage (Keychain/Keystore)
- Encrypt keys with device security
- Use biometric authentication for key access
- Implement key backup mechanism

### 2. Key Rotation

- **Identity Key**: Permanent (unless compromised)
- **Signed Pre-Key**: Rotate every 30 days
- **One-Time Pre-Keys**: Replenish when count < 20
- **Session Keys**: Automatically rotated via Double Ratchet

### 3. Forward Secrecy

- Achieved through Double Ratchet algorithm
- Each message uses a unique key
- Old keys cannot be derived from new keys
- Old keys should be securely deleted after use

### 4. Future Secrecy (Break-in Recovery)

- DH ratchet provides break-in recovery
- Compromise recovery after DH key exchange
- Limited by message authentication

### 5. Authentication

- Identity keys signed by Solana wallet
- Prevents man-in-the-middle attacks
- Verify signatures on key bundles
- Implement key fingerprint comparison

### 6. Out-of-Order Messages

- Store skipped message keys temporarily
- Set expiration time (7 days)
- Limit number of skipped keys (1000 max)
- Prevent memory exhaustion attacks

### 7. Replay Protection

- Use message counters
- Reject messages with duplicate counters
- Store message hashes temporarily

### 8. Side-Channel Attacks

- Use constant-time comparison for MACs
- Avoid timing leaks in decryption
- Use secure memory wiping for keys

---

## Testing Strategy

### Unit Tests

#### Cryptographic Primitives
- [ ] Test X25519 key generation
- [ ] Test ECDH shared secret derivation
- [ ] Test HKDF key derivation
- [ ] Test AES-256-GCM encryption/decryption
- [ ] Test key encoding/decoding

#### X3DH Protocol
- [ ] Test key agreement with all 4 DH operations
- [ ] Test key agreement with 3 DH operations (no OPK)
- [ ] Test shared secret derivation
- [ ] Test initial session setup

#### Double Ratchet
- [ ] Test DH ratchet step
- [ ] Test symmetric key ratchet
- [ ] Test message key derivation
- [ ] Test out-of-order message handling
- [ ] Test skipped message keys

#### Session Management
- [ ] Test session creation
- [ ] Test session state persistence
- [ ] Test session update
- [ ] Test session retrieval

### Integration Tests

- [ ] Test end-to-end encryption flow
- [ ] Test Alice -> Bob message encryption
- [ ] Test Bob -> Alice message encryption (ratchet)
- [ ] Test multiple message exchanges
- [ ] Test session recovery
- [ ] Test key bundle retrieval
- [ ] Test pre-key consumption
- [ ] Test pre-key replenishment

### Performance Tests

- [ ] Benchmark key generation (1000 keys)
- [ ] Benchmark encryption (various message sizes)
- [ ] Benchmark decryption (various message sizes)
- [ ] Benchmark ratchet steps
- [ ] Test concurrent sessions (1000+ users)

### Security Tests

- [ ] Test against known attack vectors
- [ ] Test key compromise scenarios
- [ ] Test replay attack prevention
- [ ] Test man-in-the-middle detection
- [ ] Fuzz testing for protocol edge cases

---

## Performance Optimizations

### 1. Key Caching
- Cache identity keys in Redis (public keys only)
- Cache active signed pre-keys
- TTL: 1 hour

### 2. Session Caching
- Cache active sessions in Redis
- Serialize session state efficiently
- TTL: 24 hours, refresh on use

### 3. Batch Operations
- Generate one-time pre-keys in batches (100)
- Upload pre-keys in single transaction
- Prefetch key bundles

### 4. Async Operations
- Make key generation async
- Queue pre-key replenishment
- Background cleanup of expired keys

### 5. Database Indexes
- Index on user_id for all key tables
- Index on (user_id, used) for one-time pre-keys
- Index on conversation_id for sessions
- Partial index for unused pre-keys

---

## Future Enhancements

### Phase 2 Features

1. **Group Messaging Encryption**
   - Sender Keys protocol
   - Efficient group key distribution
   - Member key changes handling

2. **Device Synchronization**
   - Multi-device session management
   - Session sharing between devices
   - Key backup and recovery

3. **Safety Numbers**
   - Generate key fingerprints
   - QR code verification
   - Key change notifications

4. **Sealed Sender**
   - Hide sender metadata
   - Anonymous message sending
   - Privacy-preserving delivery

5. **Post-Quantum Cryptography**
   - Integrate CRYSTALS-Dilithium (signatures)
   - Integrate CRYSTALS-Kyber (KEM)
   - Hybrid classical + PQ approach

---

## References

1. **Signal Protocol Specifications**
   - https://signal.org/docs/
   - X3DH: https://signal.org/docs/specifications/x3dh/
   - Double Ratchet: https://signal.org/docs/specifications/doubleratchet/

2. **Cryptographic Libraries**
   - Go crypto: https://pkg.go.dev/golang.org/x/crypto
   - Curve25519: https://pkg.go.dev/golang.org/x/crypto/curve25519

3. **Security Best Practices**
   - OWASP Cryptographic Storage Cheat Sheet
   - NIST Special Publication 800-57 (Key Management)

---

## Conclusion

This plan provides a comprehensive roadmap for implementing E2E encryption in SoTalk. The implementation follows industry best practices from the Signal Protocol while being adapted for the Go backend and Solana-based authentication system.

**Estimated Implementation Time**: 5-7 days for Phase 1-5

**Next Steps**:
1. Review and approve this plan
2. Start with Phase 1: Core Cryptography
3. Implement tests alongside each component
4. Document client-side integration requirements
