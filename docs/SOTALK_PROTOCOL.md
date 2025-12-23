# Sotalk Protocol - Enhanced E2E Encryption

**Version:** 2.0.0
**Status:** Design Phase
**Author:** Sotalk Team
**Date:** 2025-12-17

---

## Executive Summary

**Sotalk Protocol** is a pure Go package providing next-generation end-to-end encryption that enhances Signal Protocol with:
- ✅ **Post-Quantum Cryptography** (Hybrid approach)
- ✅ **Double Ratchet** (Signal Protocol standard)
- ✅ **Storage-Agnostic** (Interface-based design)
- ✅ **Scalable Group Encryption** (Tree-based)
- ✅ **Zero Backend Dependencies** (Pure Go implementation)

---

## Table of Contents

1. [Protocol Overview](#protocol-overview)
2. [Architecture](#architecture)
3. [Implementation Phases](#implementation-phases)
4. [API Specification](#api-specification)
5. [Cryptographic Primitives](#cryptographic-primitives)
6. [Message Format](#message-format)
7. [Testing Strategy](#testing-strategy)
8. [Integration Guide](#integration-guide)

---

## Protocol Overview

### Design Goals

1. **Quantum Resistance**: Secure against future quantum computers
2. **Zero Dependencies**: Pure Go, no infrastructure requirements
3. **Storage Agnostic**: Works with any storage backend
4. **Simple Integration**: Easy to integrate into any application
5. **Performance**: Minimal latency overhead (<100ms)
6. **Portability**: Works on any platform Go supports

### What This Package Provides

```
sotalk-protocol (Go Package)
├─ Hybrid key generation
├─ Key exchange (ECDH + KEM)
├─ Double ratchet encryption
├─ Message encrypt/decrypt
├─ Group key management
└─ Storage interfaces (implement yourself)
```

### What YOU Provide

```
Your Backend Implementation
├─ Key storage (database, file, memory, etc.)
├─ Session management
├─ Message delivery
├─ User management
└─ Rate limiting (optional)
```

---

## Architecture

### Package Structure

```
sotalk-protocol/
├── cmd/
│   └── example/
│       └── main.go                  # Example usage
├── pkg/
│   ├── crypto/
│   │   ├── hybrid.go                # Hybrid PQC crypto
│   │   ├── keys.go                  # Key generation
│   │   └── signatures.go            # Hybrid signatures
│   ├── ratchet/
│   │   ├── double_ratchet.go        # Double ratchet implementation
│   │   ├── symmetric.go             # Symmetric key ratchet
│   │   └── dh.go                    # DH ratchet
│   ├── session/
│   │   ├── session.go               # Session management
│   │   ├── store.go                 # Storage interfaces
│   │   └── manager.go               # Session manager
│   ├── message/
│   │   ├── encrypt.go               # Message encryption
│   │   ├── decrypt.go               # Message decryption
│   │   └── format.go                # Message format
│   ├── group/
│   │   ├── tree.go                  # Tree-based group keys
│   │   └── manager.go               # Group key management
│   └── types/
│       └── types.go                 # Shared types
├── examples/
│   ├── basic/                       # Basic 1-to-1 example
│   ├── group/                       # Group chat example
│   └── storage/                     # Storage implementations
│       ├── memory/                  # In-memory (testing)
│       ├── file/                    # File-based
│       └── postgres/                # PostgreSQL (optional)
├── go.mod
├── go.sum
└── README.md
```

---

## Implementation Phases

### Phase 1: Core Cryptography (Week 1)

**Goal:** Pure cryptographic primitives, zero dependencies

#### Tasks
- [ ] Implement hybrid key generation (X25519 + Kyber)
- [ ] Implement hybrid ECDH + KEM
- [ ] Implement hybrid signatures (Ed25519 + Dilithium)
- [ ] Implement HKDF key derivation
- [ ] Unit tests for all crypto functions
- [ ] Benchmarks

**Deliverables:**
```go
package crypto

// HybridKeyPair contains classical and post-quantum keys
type HybridKeyPair struct {
    // Classical keys
    X25519PrivateKey [32]byte
    X25519PublicKey  [32]byte
    Ed25519PrivateKey [64]byte
    Ed25519PublicKey  [32]byte

    // Post-quantum keys
    KyberPrivateKey []byte
    KyberPublicKey  []byte
    DilithiumPrivateKey []byte
    DilithiumPublicKey  []byte
}

// GenerateKeyPair generates a new hybrid key pair
func GenerateKeyPair() (*HybridKeyPair, error)

// ComputeSharedSecret performs hybrid key exchange
func ComputeSharedSecret(myPrivate, peerPublic *HybridKeyPair) ([]byte, error)

// Sign creates a hybrid signature
func Sign(message []byte, privateKey *HybridKeyPair) ([]byte, error)

// Verify checks a hybrid signature
func Verify(message, signature []byte, publicKey *HybridKeyPair) bool

// Fingerprint generates a key fingerprint
func (k *HybridKeyPair) Fingerprint() [32]byte
```

---

### Phase 2: Double Ratchet (Week 2)

**Goal:** Implement Signal's double ratchet algorithm

#### Tasks
- [ ] Implement symmetric key ratchet
- [ ] Implement DH ratchet
- [ ] Handle out-of-order messages
- [ ] State serialization/deserialization
- [ ] Unit tests
- [ ] Edge case testing

**Deliverables:**
```go
package ratchet

// DoubleRatchet manages encryption state
type DoubleRatchet struct {
    // Sending chain
    sendingChainKey []byte
    sendingCounter  uint32

    // Receiving chain
    receivingChainKey []byte
    receivingCounter  uint32

    // DH ratchet keys
    dhSendingKey  *crypto.HybridKeyPair
    dhReceivingKey *crypto.HybridKeyPair
    dhRatchetKey  *crypto.HybridKeyPair

    // Out-of-order message handling
    skippedKeys map[uint32][]byte
}

// NewDoubleRatchet creates a new ratchet
func NewDoubleRatchet(
    sharedSecret []byte,
    myDHKey *crypto.HybridKeyPair,
    peerDHPubKey *crypto.HybridKeyPair,
    isSender bool,
) (*DoubleRatchet, error)

// Encrypt encrypts a message
func (r *DoubleRatchet) Encrypt(plaintext []byte) (*EncryptedMessage, error)

// Decrypt decrypts a message
func (r *DoubleRatchet) Decrypt(msg *EncryptedMessage) ([]byte, error)

// GetState serializes the ratchet state
func (r *DoubleRatchet) GetState() ([]byte, error)

// LoadState deserializes the ratchet state
func LoadState(data []byte) (*DoubleRatchet, error)
```

---

### Phase 3: Session Management (Week 3)

**Goal:** High-level session API with storage interfaces

#### Tasks
- [ ] Define storage interfaces
- [ ] Implement session manager
- [ ] Session establishment flow
- [ ] Pre-key management
- [ ] In-memory storage (for testing)
- [ ] Integration tests

**Deliverables:**
```go
package session

// Storage interface - implement this in your backend
type Storage interface {
    // Identity keys
    SaveIdentityKey(userID string, key *crypto.HybridKeyPair) error
    GetIdentityKey(userID string) (*crypto.HybridKeyPair, error)

    // Pre-keys
    SavePreKey(userID string, keyID uint32, key *crypto.HybridKeyPair) error
    GetPreKey(userID string) (*crypto.HybridKeyPair, uint32, error)
    DeletePreKey(userID string, keyID uint32) error

    // Sessions
    SaveSession(sessionID string, state []byte) error
    GetSession(sessionID string) ([]byte, error)
    DeleteSession(sessionID string) error

    // Audit (optional)
    LogKeyRotation(userID string, reason string) error
}

// Manager handles session lifecycle
type Manager struct {
    storage Storage
    myUserID string
}

// NewManager creates a new session manager
func NewManager(storage Storage, myUserID string) *Manager

// Initialize generates initial keys
func (m *Manager) Initialize() error

// EstablishSession creates a new session with a peer
func (m *Manager) EstablishSession(peerUserID string) (*Session, error)

// GetSession retrieves an existing session
func (m *Manager) GetSession(peerUserID string) (*Session, error)

// RotateKeys generates new pre-keys
func (m *Manager) RotateKeys() error
```

---

### Phase 4: Message Encryption (Week 4)

**Goal:** Complete encrypt/decrypt API

#### Tasks
- [ ] Implement message encryption
- [ ] Implement message decryption
- [ ] Add signature verification
- [ ] Replay protection
- [ ] Associated data
- [ ] Performance optimization

**Deliverables:**
```go
package message

// Encryptor handles message encryption
type Encryptor struct {
    session *session.Session
}

// NewEncryptor creates a new encryptor
func NewEncryptor(sess *session.Session) *Encryptor

// Encrypt encrypts a message
func (e *Encryptor) Encrypt(plaintext []byte) (*EncryptedMessage, error)

// Decrypt decrypts a message
func (e *Encryptor) Decrypt(msg *EncryptedMessage) ([]byte, error)

// EncryptedMessage is the wire format
type EncryptedMessage struct {
    Version      uint8      // Protocol version
    SenderID     string     // Sender user ID
    RecipientID  string     // Recipient user ID
    Timestamp    int64      // Unix timestamp
    Counter      uint32     // Message counter

    // Crypto fields
    DHPublicKey  []byte     // Ephemeral DH public key
    Ciphertext   []byte     // Encrypted content

    // Authentication
    Signature    []byte     // Hybrid signature
}

// ToBytes serializes the message
func (m *EncryptedMessage) ToBytes() []byte

// FromBytes deserializes the message
func FromBytes(data []byte) (*EncryptedMessage, error)
```

---

### Phase 5: Group Encryption (Week 5)

**Goal:** Tree-based group key management

#### Tasks
- [ ] Implement tree structure
- [ ] Member add protocol
- [ ] Member remove protocol
- [ ] Key derivation
- [ ] Forward/backward secrecy
- [ ] Tests with large groups

**Deliverables:**
```go
package group

// Manager manages group keys
type Manager struct {
    storage session.Storage
}

// NewManager creates a new group manager
func NewManager(storage session.Storage) *Manager

// CreateGroup creates a new group
func (m *Manager) CreateGroup(creatorID string, memberIDs []string) (*Group, error)

// AddMember adds a member to the group
func (m *Manager) AddMember(groupID, memberID string) error

// RemoveMember removes a member from the group
func (m *Manager) RemoveMember(groupID, memberID string) error

// EncryptGroupMessage encrypts a message for all members
func (m *Manager) EncryptGroupMessage(groupID string, plaintext []byte) ([]*message.EncryptedMessage, error)

// DecryptGroupMessage decrypts a group message
func (m *Manager) DecryptGroupMessage(msg *message.EncryptedMessage) ([]byte, error)
```

---

### Phase 6: Testing & Documentation (Week 6)

**Goal:** Production-ready package

#### Tasks
- [ ] Comprehensive unit tests (>90% coverage)
- [ ] Integration tests
- [ ] Benchmarks
- [ ] Example implementations
- [ ] API documentation
- [ ] Security audit preparation

---

## API Specification

### Simple Usage Example

```go
package main

import (
    "github.com/sotalk/protocol/pkg/crypto"
    "github.com/sotalk/protocol/pkg/session"
    "github.com/sotalk/protocol/pkg/message"
)

func main() {
    // 1. Initialize storage (you implement this)
    storage := NewMyStorage()

    // 2. Create session manager
    alice := session.NewManager(storage, "alice")
    alice.Initialize()

    bob := session.NewManager(storage, "bob")
    bob.Initialize()

    // 3. Establish session
    aliceSession, _ := alice.EstablishSession("bob")

    // 4. Encrypt message
    encryptor := message.NewEncryptor(aliceSession)
    encrypted, _ := encryptor.Encrypt([]byte("Hello Bob!"))

    // 5. Decrypt message (on Bob's side)
    bobSession, _ := bob.GetSession("alice")
    decryptor := message.NewEncryptor(bobSession)
    plaintext, _ := decryptor.Decrypt(encrypted)

    println(string(plaintext)) // "Hello Bob!"
}
```

### Storage Implementation Example

```go
// Example: In-memory storage (for testing)
package storage

type MemoryStorage struct {
    identityKeys map[string]*crypto.HybridKeyPair
    preKeys      map[string]map[uint32]*crypto.HybridKeyPair
    sessions     map[string][]byte
    mu           sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        identityKeys: make(map[string]*crypto.HybridKeyPair),
        preKeys:      make(map[string]map[uint32]*crypto.HybridKeyPair),
        sessions:     make(map[string][]byte),
    }
}

func (s *MemoryStorage) SaveIdentityKey(userID string, key *crypto.HybridKeyPair) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.identityKeys[userID] = key
    return nil
}

func (s *MemoryStorage) GetIdentityKey(userID string) (*crypto.HybridKeyPair, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    key, ok := s.identityKeys[userID]
    if !ok {
        return nil, errors.New("key not found")
    }
    return key, nil
}

// ... implement other methods
```

---

## Cryptographic Primitives

### 1. Hybrid Key Exchange

```
Shared Secret = HKDF(
    classical_secret || pqc_secret,
    salt: "Sotalk-v2-kex",
    info: sender_id || recipient_id
)

where:
  classical_secret = X25519(my_x25519_priv, peer_x25519_pub)
  pqc_secret = Kyber1024.Decapsulate(my_kyber_priv, peer_kyber_encap)
```

### 2. Hybrid Signatures

```
Signature = classical_sig || pqc_sig

where:
  classical_sig = Ed25519.Sign(message, ed25519_priv)
  pqc_sig = Dilithium3.Sign(message, dilithium_priv)

Verification:
  valid = Ed25519.Verify(...) AND Dilithium3.Verify(...)
```

### 3. Message Encryption

```
AEAD Encryption:

message_key = HKDF(chain_key, "message-key", counter)
ciphertext = ChaCha20-Poly1305.Encrypt(
    key:   message_key,
    nonce: counter || timestamp,
    plaintext: message,
    ad: sender_id || recipient_id || counter
)
```

### 4. Key Derivation

```
Root Key → Chain Key → Message Key

root_key_new = HKDF(root_key, dh_output, "root-chain")
chain_key_new = HMAC(chain_key, 0x01)
message_key = HMAC(chain_key, 0x02 || counter)
```

---

## Message Format

### Wire Format (Binary)

```
EncryptedMessage (variable length):

┌─────────────────────────────────────────┐
│ Version (1 byte)                        │
├─────────────────────────────────────────┤
│ Sender ID (variable, length-prefixed)   │
├─────────────────────────────────────────┤
│ Recipient ID (variable, length-prefixed)│
├─────────────────────────────────────────┤
│ Timestamp (8 bytes, int64)              │
├─────────────────────────────────────────┤
│ Counter (4 bytes, uint32)               │
├─────────────────────────────────────────┤
│ DH Public Key (variable)                │
│   - Classical (32 bytes)                │
│   - PQC (1568 bytes)                    │
├─────────────────────────────────────────┤
│ Ciphertext (variable length)            │
├─────────────────────────────────────────┤
│ Signature (variable)                    │
│   - Classical (64 bytes)                │
│   - PQC (3309 bytes)                    │
└─────────────────────────────────────────┘

Total size: ~5KB + message length
```

### JSON Format (for debugging)

```json
{
  "version": 2,
  "sender_id": "alice",
  "recipient_id": "bob",
  "timestamp": 1734480000,
  "counter": 42,
  "dh_public_key": {
    "classical": "base64...",
    "pqc": "base64..."
  },
  "ciphertext": "base64...",
  "signature": {
    "classical": "base64...",
    "pqc": "base64..."
  }
}
```

---

## Performance Targets

### Benchmarks

```
Operation                    Target      Acceptable
──────────────────────────────────────────────────
Key generation (hybrid)      < 20ms      < 50ms
Key exchange                 < 50ms      < 100ms
Message encryption           < 2ms       < 5ms
Message decryption           < 2ms       < 5ms
Session establishment        < 100ms     < 200ms
Group member add             < 100ms     < 200ms

Memory usage per session:    < 10KB
Package size:                < 5MB
```

### Size Metrics

```
Data Type                   Size
────────────────────────────────────
Hybrid public key           1,600 bytes
Hybrid private key          3,200 bytes
Hybrid signature            3,373 bytes
Encrypted message overhead  ~5KB
Session state               ~3KB
```

---

## Testing Strategy

### Unit Tests

```go
// Test hybrid key exchange
func TestHybridKeyExchange(t *testing.T) {
    alice, _ := crypto.GenerateKeyPair()
    bob, _ := crypto.GenerateKeyPair()

    secretA, _ := crypto.ComputeSharedSecret(alice, bob)
    secretB, _ := crypto.ComputeSharedSecret(bob, alice)

    if !bytes.Equal(secretA, secretB) {
        t.Fatal("shared secrets don't match")
    }
}

// Test double ratchet
func TestDoubleRatchet(t *testing.T) {
    // ... test encryption/decryption
}

// Test out-of-order messages
func TestOutOfOrderMessages(t *testing.T) {
    // ... test message reordering
}
```

### Integration Tests

```go
// Full E2E flow
func TestEndToEnd(t *testing.T) {
    // 1. Setup
    storage := NewMemoryStorage()
    alice := session.NewManager(storage, "alice")
    bob := session.NewManager(storage, "bob")

    // 2. Initialize
    alice.Initialize()
    bob.Initialize()

    // 3. Establish session
    aliceSession, _ := alice.EstablishSession("bob")

    // 4. Send 100 messages
    for i := 0; i < 100; i++ {
        msg := fmt.Sprintf("Message %d", i)
        encrypted, _ := message.Encrypt(aliceSession, []byte(msg))

        bobSession, _ := bob.GetSession("alice")
        decrypted, _ := message.Decrypt(bobSession, encrypted)

        if string(decrypted) != msg {
            t.Fatalf("message mismatch: %s", string(decrypted))
        }
    }
}
```

### Benchmarks

```go
func BenchmarkEncryption(b *testing.B) {
    storage := NewMemoryStorage()
    alice := session.NewManager(storage, "alice")
    bob := session.NewManager(storage, "bob")
    alice.Initialize()
    bob.Initialize()

    sess, _ := alice.EstablishSession("bob")
    plaintext := []byte("Hello, world!")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        message.Encrypt(sess, plaintext)
    }
}
```

---

## Integration Guide

### For Backend Developers

**1. Install the package:**
```bash
go get github.com/sotalk/protocol
```

**2. Implement storage interface:**
```go
// Use PostgreSQL, Redis, files, or anything you want
type MyStorage struct {
    db *sql.DB
}

func (s *MyStorage) SaveIdentityKey(userID string, key *crypto.HybridKeyPair) error {
    // Store in your database
    return s.db.Exec("INSERT INTO identity_keys ...")
}
// ... implement other methods
```

**3. Integrate into your API:**
```go
// In your message handler
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
    var req SendMessageRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Get session
    sess, err := h.sessionManager.GetSession(req.RecipientID)
    if err != nil {
        // Establish new session
        sess, err = h.sessionManager.EstablishSession(req.RecipientID)
    }

    // Encrypt message
    encryptor := message.NewEncryptor(sess)
    encrypted, err := encryptor.Encrypt([]byte(req.Content))

    // Store/send encrypted message
    h.messageRepo.Save(encrypted)
    h.websocket.Send(req.RecipientID, encrypted)
}
```

### Example Implementations

The package includes reference implementations:

1. **Memory Storage** - For testing and development
2. **File Storage** - Simple file-based storage
3. **PostgreSQL Storage** - Production-ready SQL storage
4. **Redis Storage** - For session caching

You can use these as templates or implement your own!

---

## Dependencies

### Minimal Dependencies

```go
// go.mod
module github.com/sotalk/protocol

go 1.21

require (
    golang.org/x/crypto v0.17.0           // ChaCha20, HKDF
    github.com/cloudflare/circl v1.3.7    // Post-quantum crypto
)
```

**No database, no web framework, no infrastructure dependencies!**

---

## Security Considerations

### What This Package Protects

✅ Message confidentiality (end-to-end encryption)
✅ Message authenticity (hybrid signatures)
✅ Forward secrecy (double ratchet)
✅ Post-compromise security (key rotation)
✅ Quantum resistance (hybrid PQC)

### What YOU Must Protect

⚠️ Secure key storage (encrypt private keys at rest)
⚠️ Secure transport (use TLS for message delivery)
⚠️ Access control (authenticate users)
⚠️ Rate limiting (prevent spam/DoS)
⚠️ Audit logging (track key operations)

### Best Practices

1. **Never log plaintext or private keys**
2. **Encrypt private keys at rest**
3. **Use secure random for key generation**
4. **Verify key fingerprints out-of-band**
5. **Rotate pre-keys regularly (weekly)**
6. **Delete old session states**

---

## Roadmap

### Version 1.0 (Current)
- [x] Hybrid PQC cryptography
- [x] Double ratchet
- [x] Session management
- [x] 1-to-1 messaging

### Version 1.1 (Future)
- [ ] Group encryption
- [ ] Multi-device support
- [ ] Key verification protocol
- [ ] Session backup/restore

### Version 2.0 (Future)
- [ ] Sealed sender (metadata privacy)
- [ ] Deniable authentication
- [ ] Post-quantum signatures only
- [ ] Zero-knowledge proofs

---

## FAQ

**Q: Do I need Solana/blockchain?**
A: No! This is a pure Go package. Use any storage backend you want.

**Q: Do I need PostgreSQL?**
A: No! Implement the `Storage` interface with anything (files, SQLite, MongoDB, etc.)

**Q: Can I use this without Sotalk backend?**
A: Yes! This is a standalone encryption library. Build your own app with it.

**Q: Is this compatible with Signal?**
A: No. It uses hybrid PQC which Signal doesn't support yet. But the ratchet is the same.

**Q: How do I handle key verification?**
A: Compare key fingerprints out-of-band (QR codes, phone calls, etc.). We provide `Fingerprint()` method.

**Q: What if quantum computers break this?**
A: The hybrid approach protects you. Even if classical crypto breaks, PQC remains secure.

**Q: Can I use this in production?**
A: After security audit, yes. The crypto is based on NIST standards.

---

## Contributing

We welcome contributions! Areas we need help:

1. **Performance optimization** - SIMD, assembly
2. **Additional storage implementations** - MongoDB, DynamoDB, etc.
3. **Client SDKs** - JavaScript, Python, Swift
4. **Documentation** - More examples, tutorials
5. **Security review** - Cryptography experts

---

## License

MIT License - Use it however you want!

---

## Contact & Support

- **Documentation**: https://pkg.go.dev/github.com/sotalk/protocol
- **GitHub**: https://github.com/sotalk/protocol
- **Issues**: https://github.com/sotalk/protocol/issues
- **Email**: security@sotalk.app

---

**Document Version:** 2.0.0
**Last Updated:** 2025-12-17
**Next Review:** 2026-01-17

---

© 2025 Sotalk Team. All rights reserved.
