# 14-Day Enterprise Plan: Solana Secure Messaging Platform

## Project Overview
A secure, privacy-focused messaging application using Solana wallet addresses as user identities, featuring end-to-end encryption, real-time messaging, and integrated wallet functionality.

**Tech Stack:**
- Backend: Go (Gin framework)
- Database: PostgreSQL with GORM (ORM)
- Real-time: WebSocket
- Blockchain: Solana
- Architecture: Clean Architecture (Uncle Bob's Pattern)
- Security: End-to-end encryption, message signing with Solana keys

---

## Architecture Overview

### Clean Architecture (Uncle Bob)
```
┌──────────────────────────────────────────────────────────────────┐
│                    Frameworks & Drivers                          │
│                    (External Layer)                              │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │   Gin    │  │  GORM    │  │  Redis   │  │  Solana  │      │
│  │   Web    │  │PostgreSQL│  │  Cache   │  │   RPC    │      │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘      │
└────────────────────────────┬─────────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────────┐
│              Interface Adapters Layer                            │
│              (Controllers, Presenters, Gateways)                 │
│                                                                  │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │   Controllers    │  │   Repositories   │                    │
│  │  (HTTP/WS)       │  │   (GORM impl)    │                    │
│  └──────────────────┘  └──────────────────┘                    │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │   Presenters     │  │    Gateways      │                    │
│  │   (Response DTO) │  │ (External APIs)  │                    │
│  └──────────────────┘  └──────────────────┘                    │
└────────────────────────────┬─────────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────────┐
│                 Use Cases Layer                                  │
│         (Application Business Rules)                             │
│                                                                  │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐   │
│  │  Auth Service  │  │Message Service │  │ Wallet Service │   │
│  └────────────────┘  └────────────────┘  └────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         Repository Interfaces (Input Boundaries)        │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────┬─────────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────────┐
│                   Entities Layer                                 │
│          (Enterprise Business Rules)                             │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │   User   │  │ Message  │  │  Wallet  │  │  Media   │       │
│  │  Entity  │  │  Entity  │  │  Entity  │  │  Entity  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │        Domain Rules & Value Objects                     │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

**The Dependency Rule:**
> Source code dependencies must point only inward, toward higher-level policies.

**Layer Responsibilities:**

1. **Entities (Domain Layer)** - Core business objects
   - User, Message, Wallet, Conversation entities
   - Domain rules and business logic
   - Framework-independent

2. **Use Cases (Application Layer)** - Application-specific business rules
   - AuthService, MessageService, WalletService
   - Orchestrates data flow to/from entities
   - Defines repository interfaces

3. **Interface Adapters (Adapter Layer)** - Convert data formats
   - HTTP Controllers (Gin handlers)
   - WebSocket handlers
   - GORM repository implementations
   - DTO (Data Transfer Objects)

4. **Frameworks & Drivers (Infrastructure Layer)** - External tools
   - Web framework (Gin)
   - Database (PostgreSQL + GORM)
   - Cache (Redis)
   - Blockchain (Solana SDK)

### Core Modules
1. **Authentication** - Solana wallet-based auth
2. **Messaging** - 1-to-1, group chats, channels
3. **Wallet** - Solana transactions, balance management
4. **Media** - File uploads, media sharing
5. **Presence** - Online/offline status
6. **Notifications** - Push notifications
7. **Encryption** - E2E encryption layer

---

## Day 1-2: Foundation & Authentication

### Day 1: Project Setup & Core Infrastructure

**Tasks:**
1. Initialize Go project structure (Clean Architecture)
   ```
   soldef/
   ├── cmd/
   │   └── api/
   │       └── main.go                    # Application entry point
   ├── internal/
   │   ├── domain/                        # ENTITIES LAYER (Enterprise Business Rules)
   │   │   ├── user/
   │   │   │   ├── entity.go              # User entity
   │   │   │   ├── repository.go          # Repository interface
   │   │   │   └── errors.go              # Domain errors
   │   │   ├── message/
   │   │   │   ├── entity.go
   │   │   │   ├── repository.go
   │   │   │   └── value_objects.go
   │   │   ├── conversation/
   │   │   ├── wallet/
   │   │   ├── media/
   │   │   └── common/
   │   │       └── types.go
   │   ├── usecase/                       # USE CASES LAYER (Application Business Rules)
   │   │   ├── auth/
   │   │   │   ├── service.go             # Auth use cases
   │   │   │   ├── interface.go           # Service interface
   │   │   │   └── solana_verifier.go
   │   │   ├── message/
   │   │   │   ├── service.go
   │   │   │   └── interface.go
   │   │   ├── wallet/
   │   │   │   ├── service.go
   │   │   │   └── interface.go
   │   │   └── dto/                       # DTOs for use cases
   │   │       ├── auth_dto.go
   │   │       ├── message_dto.go
   │   │       └── wallet_dto.go
   │   ├── delivery/                      # INTERFACE ADAPTERS LAYER (Controllers & Presenters)
   │   │   ├── http/
   │   │   │   ├── handler/
   │   │   │   │   ├── auth_handler.go    # HTTP controllers
   │   │   │   │   ├── message_handler.go
   │   │   │   │   └── wallet_handler.go
   │   │   │   ├── middleware/
   │   │   │   │   ├── auth.go
   │   │   │   │   ├── cors.go
   │   │   │   │   └── rate_limit.go
   │   │   │   ├── request/               # Request DTOs
   │   │   │   │   └── auth_request.go
   │   │   │   ├── response/              # Response DTOs (Presenters)
   │   │   │   │   └── auth_response.go
   │   │   │   └── router.go
   │   │   └── websocket/
   │   │       ├── handler.go
   │   │       ├── hub.go
   │   │       └── client.go
   │   └── repository/                    # INTERFACE ADAPTERS (Repository Implementations)
   │       ├── postgres/
   │       │   ├── user_repository.go     # GORM implementation
   │       │   ├── message_repository.go
   │       │   ├── models.go              # GORM models
   │       │   └── gorm.go                # DB connection
   │       ├── redis/
   │       │   ├── cache.go
   │       │   └── presence.go
   │       ├── solana/
   │       │   ├── client.go
   │       │   └── transaction.go
   │       └── s3/
   │           └── storage.go
   ├── pkg/                               # FRAMEWORKS & DRIVERS (Shared)
   │   ├── config/                        # Configuration
   │   │   └── config.go
   │   ├── logger/                        # Logger
   │   │   └── logger.go
   │   ├── validator/                     # Validation
   │   │   └── validator.go
   │   ├── crypto/                        # Encryption utilities
   │   │   └── crypto.go
   │   └── middleware/                    # Reusable middleware
   │       └── jwt.go
   ├── migrations/                        # Database migrations
   ├── scripts/                           # Build & deploy scripts
   └── docker/                            # Docker configs
   ```

2. Setup dependencies
   ```go
   // go.mod essential packages
   - github.com/gin-gonic/gin              // HTTP framework
   - gorm.io/gorm                          // ORM
   - gorm.io/driver/postgres               // PostgreSQL driver
   - github.com/gorilla/websocket          // WebSocket
   - github.com/gagliardetto/solana-go     // Solana SDK
   - github.com/redis/go-redis/v9          // Redis client
   - github.com/golang-migrate/migrate/v4  // Database migrations
   - github.com/google/uuid                // UUID generation
   - github.com/golang-jwt/jwt/v5          // JWT tokens
   - github.com/joho/godotenv              // Environment variables
   - go.uber.org/zap                       // Structured logging
   ```

3. Configuration management
   - Environment-based config
   - Config validation
   - GORM connection pooling
   - Solana RPC endpoints (mainnet, devnet, testnet)

4. Database setup with GORM
   - PostgreSQL connection with GORM
   - Auto-migration support
   - Custom GORM configuration
   - Migration files for production

**Deliverables:**
- ✅ Project structure initialized
- ✅ Docker Compose setup (Postgres, Redis)
- ✅ Configuration management
- ✅ Logger implementation
- ✅ Health check endpoint

---

### Day 2: Solana Authentication System

**Tasks:**
1. **Domain Layer** - User entity (ENTITIES)
   ```go
   // internal/domain/user/entity.go
   package user

   import (
       "time"
       "github.com/google/uuid"
   )

   // Domain Entity (Pure business logic, no dependencies)
   type User struct {
       ID            uuid.UUID
       WalletAddress string
       Username      *string
       DisplayName   string
       Avatar        *string
       PublicKey     string
       Status        Status
       LastSeen      time.Time
       CreatedAt     time.Time
       UpdatedAt     time.Time
   }

   type Status string

   const (
       StatusOnline  Status = "online"
       StatusOffline Status = "offline"
       StatusAway    Status = "away"
   )

   // internal/domain/user/repository.go
   package user

   import (
       "context"
       "github.com/google/uuid"
   )

   // Repository interface (defined in domain)
   type Repository interface {
       Create(ctx context.Context, user *User) error
       FindByID(ctx context.Context, id uuid.UUID) (*User, error)
       FindByWalletAddress(ctx context.Context, address string) (*User, error)
       Update(ctx context.Context, user *User) error
   }
   ```

2. **Use Case Layer** - Auth service (APPLICATION BUSINESS RULES)
   ```go
   // internal/usecase/auth/interface.go
   package auth

   type Service interface {
       GenerateChallenge(walletAddress string) (string, error)
       VerifyAndLogin(walletAddress, signature, message string) (string, error)
       RefreshToken(refreshToken string) (string, error)
   }

   // internal/usecase/auth/service.go
   package auth

   import (
       "your-module/internal/domain/user"
   )

   type service struct {
       userRepo user.Repository
       jwtSecret string
   }

   func NewService(userRepo user.Repository, jwtSecret string) Service {
       return &service{
           userRepo: userRepo,
           jwtSecret: jwtSecret,
       }
   }

   func (s *service) GenerateChallenge(walletAddress string) (string, error) {
       // Business logic here
       return "sign-this-message", nil
   }
   ```

3. **Repository Layer** - GORM implementation (INTERFACE ADAPTER)
   ```go
   // internal/repository/postgres/models.go
   package postgres

   import (
       "time"
       "github.com/google/uuid"
       "gorm.io/gorm"
   )

   type User struct {
       ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
       WalletAddress string    `gorm:"type:varchar(44);uniqueIndex;not null"`
       Username      *string   `gorm:"type:varchar(50);uniqueIndex"`
       DisplayName   string    `gorm:"type:varchar(100)"`
       Avatar        *string   `gorm:"type:text"`
       PublicKey     string    `gorm:"type:text;not null"`
       Status        string    `gorm:"type:varchar(20);default:'offline'"`
       LastSeen      time.Time
       CreatedAt     time.Time
       UpdatedAt     time.Time
       DeletedAt     gorm.DeletedAt `gorm:"index"`
   }

   // internal/repository/postgres/user_repository.go
   package postgres

   import (
       "context"
       "gorm.io/gorm"
       domainUser "your-module/internal/domain/user"
   )

   type userRepository struct {
       db *gorm.DB
   }

   func NewUserRepository(db *gorm.DB) domainUser.Repository {
       return &userRepository{db: db}
   }

   func (r *userRepository) Create(ctx context.Context, user *domainUser.User) error {
       dbUser := &User{
           ID:            user.ID,
           WalletAddress: user.WalletAddress,
           Username:      user.Username,
           DisplayName:   user.DisplayName,
           PublicKey:     user.PublicKey,
           Status:        string(user.Status),
       }
       return r.db.WithContext(ctx).Create(dbUser).Error
   }

   func (r *userRepository) FindByWalletAddress(ctx context.Context, address string) (*domainUser.User, error) {
       var dbUser User
       if err := r.db.WithContext(ctx).Where("wallet_address = ?", address).First(&dbUser).Error; err != nil {
           return nil, err
       }
       return &domainUser.User{
           ID:            dbUser.ID,
           WalletAddress: dbUser.WalletAddress,
           Username:      dbUser.Username,
           DisplayName:   dbUser.DisplayName,
           Status:        domainUser.Status(dbUser.Status),
       }, nil
   }
   ```

4. **HTTP Handler** - Controllers (INTERFACE ADAPTER)
   ```go
   // internal/delivery/http/handler/auth_handler.go
   package handler

   import (
       "github.com/gin-gonic/gin"
       "your-module/internal/usecase/auth"
   )

   type AuthHandler struct {
       authService auth.Service
   }

   func NewAuthHandler(authService auth.Service) *AuthHandler {
       return &AuthHandler{authService: authService}
   }

   func (h *AuthHandler) GenerateChallenge(c *gin.Context) {
       // Handle HTTP request
   }
   ```

5. Authentication flow
   - Challenge generation
   - Signature verification with Solana wallet
   - JWT token generation
   - Refresh token mechanism

6. API endpoints
   - `POST /api/v1/auth/challenge` - Get auth challenge
   - `POST /api/v1/auth/verify` - Verify signature & login
   - `POST /api/v1/auth/refresh` - Refresh tokens
   - `GET /api/v1/auth/me` - Get current user

**Deliverables:**
- ✅ Solana wallet authentication
- ✅ JWT middleware
- ✅ User repository with GORM
- ✅ Clean Architecture implemented
- ✅ Auth endpoints tested

---

## Day 3-4: Messaging Core

### Day 3: Direct Messaging (1-to-1)

**Tasks:**
1. Domain entities
   ```go
   type Message struct {
       ID              uuid.UUID
       ConversationID  uuid.UUID
       SenderID        uuid.UUID
       Content         []byte    // Encrypted content
       ContentType     MessageType
       Signature       string    // Solana signature
       ReplyToID       *uuid.UUID
       Status          MessageStatus
       CreatedAt       time.Time
       UpdatedAt       time.Time
   }

   type Conversation struct {
       ID              uuid.UUID
       Type            ConversationType // direct, group, channel
       Participants    []uuid.UUID
       LastMessageID   *uuid.UUID
       LastMessageAt   time.Time
       CreatedAt       time.Time
   }
   ```

2. Database schema
   ```sql
   CREATE TABLE conversations (
       id UUID PRIMARY KEY,
       type VARCHAR(20) NOT NULL,
       last_message_id UUID,
       last_message_at TIMESTAMP,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE conversation_participants (
       conversation_id UUID REFERENCES conversations(id),
       user_id UUID REFERENCES users(id),
       role VARCHAR(20),
       joined_at TIMESTAMP DEFAULT NOW(),
       last_read_at TIMESTAMP,
       PRIMARY KEY (conversation_id, user_id)
   );

   CREATE TABLE messages (
       id UUID PRIMARY KEY,
       conversation_id UUID REFERENCES conversations(id),
       sender_id UUID REFERENCES users(id),
       content BYTEA NOT NULL,
       content_type VARCHAR(20),
       signature TEXT,
       reply_to_id UUID,
       status VARCHAR(20),
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW()
   );

   CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
   CREATE INDEX idx_messages_sender ON messages(sender_id);
   ```

3. Messaging use cases
   - Send message
   - Get conversation messages (pagination)
   - Mark as read
   - Delete message
   - Edit message

4. API endpoints
   - `POST /api/v1/conversations` - Create conversation
   - `GET /api/v1/conversations` - List user conversations
   - `GET /api/v1/conversations/:id/messages` - Get messages
   - `POST /api/v1/messages` - Send message
   - `DELETE /api/v1/messages/:id` - Delete message

**Deliverables:**
- ✅ Message domain models
- ✅ Repository layer with GORM
- ✅ REST API for messaging
- ✅ Clean Architecture for messaging
- ✅ Message signing with Solana keys

---

### Day 4: WebSocket & Real-time Messaging

**Tasks:**
1. WebSocket infrastructure
   ```go
   type Hub struct {
       clients    map[uuid.UUID]*Client
       broadcast  chan *Message
       register   chan *Client
       unregister chan *Client
   }

   type Client struct {
       id     uuid.UUID
       hub    *Hub
       conn   *websocket.Conn
       send   chan []byte
   }
   ```

2. WebSocket events
   ```json
   {
     "type": "message.new",
     "payload": {...}
   }
   {
     "type": "message.delivered",
     "payload": {...}
   }
   {
     "type": "message.read",
     "payload": {...}
   }
   {
     "type": "typing.start",
     "payload": {...}
   }
   ```

3. Connection management
   - Authentication on connect
   - Heartbeat/ping-pong
   - Reconnection handling
   - Multiple device support

4. Message delivery
   - Online delivery via WebSocket
   - Offline message queuing
   - Delivery receipts
   - Read receipts

**Deliverables:**
- ✅ WebSocket server implementation
- ✅ Real-time message delivery
- ✅ Connection state management
- ✅ Typing indicators

---

## Day 5-6: Group Features & Channels

### Day 5: Group Messaging

**Tasks:**
1. Group domain models
   ```go
   type Group struct {
       ID          uuid.UUID
       Name        string
       Description string
       Avatar      string
       CreatorID   uuid.UUID
       Settings    GroupSettings
       CreatedAt   time.Time
   }

   type GroupMember struct {
       GroupID   uuid.UUID
       UserID    uuid.UUID
       Role      MemberRole // admin, moderator, member
       JoinedAt  time.Time
   }
   ```

2. Database schema
   ```sql
   CREATE TABLE groups (
       id UUID PRIMARY KEY,
       conversation_id UUID REFERENCES conversations(id),
       name VARCHAR(100) NOT NULL,
       description TEXT,
       avatar TEXT,
       creator_id UUID REFERENCES users(id),
       max_members INT DEFAULT 256,
       settings JSONB,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE group_members (
       group_id UUID REFERENCES groups(id),
       user_id UUID REFERENCES users(id),
       role VARCHAR(20) DEFAULT 'member',
       permissions JSONB,
       joined_at TIMESTAMP DEFAULT NOW(),
       PRIMARY KEY (group_id, user_id)
   );
   ```

3. Group operations
   - Create group
   - Add/remove members
   - Update group info
   - Member permissions
   - Group settings (who can message, add members, etc.)

4. API endpoints
   - `POST /api/v1/groups` - Create group
   - `GET /api/v1/groups/:id` - Get group info
   - `POST /api/v1/groups/:id/members` - Add member
   - `DELETE /api/v1/groups/:id/members/:userId` - Remove member
   - `PUT /api/v1/groups/:id` - Update group

**Deliverables:**
- ✅ Group messaging functionality
- ✅ Member management
- ✅ Role-based permissions
- ✅ Group settings

---

### Day 6: Channels & Broadcast Lists

**Tasks:**
1. Channel implementation
   - One-to-many messaging
   - Admin-only posting
   - Unlimited subscribers
   - Public/private channels

2. Database schema
   ```sql
   CREATE TABLE channels (
       id UUID PRIMARY KEY,
       conversation_id UUID REFERENCES conversations(id),
       name VARCHAR(100) NOT NULL,
       username VARCHAR(50) UNIQUE,
       description TEXT,
       avatar TEXT,
       owner_id UUID REFERENCES users(id),
       is_public BOOLEAN DEFAULT true,
       subscriber_count INT DEFAULT 0,
       settings JSONB,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE channel_subscribers (
       channel_id UUID REFERENCES channels(id),
       user_id UUID REFERENCES users(id),
       notifications_enabled BOOLEAN DEFAULT true,
       subscribed_at TIMESTAMP DEFAULT NOW(),
       PRIMARY KEY (channel_id, user_id)
   );

   CREATE TABLE channel_admins (
       channel_id UUID REFERENCES channels(id),
       user_id UUID REFERENCES users(id),
       permissions JSONB,
       PRIMARY KEY (channel_id, user_id)
   );
   ```

3. Channel features
   - Create/manage channels
   - Subscribe/unsubscribe
   - Post to channel
   - Channel analytics
   - Forward messages

4. API endpoints
   - `POST /api/v1/channels` - Create channel
   - `GET /api/v1/channels/:username` - Get channel by username
   - `POST /api/v1/channels/:id/subscribe` - Subscribe
   - `POST /api/v1/channels/:id/messages` - Post message

**Deliverables:**
- ✅ Channel functionality
- ✅ Broadcast messaging
- ✅ Subscription management
- ✅ Channel discovery

---

## Day 7-8: Media & File Handling

### Day 7: Media Upload & Storage

**Tasks:**
1. Media domain models
   ```go
   type Media struct {
       ID           uuid.UUID
       UserID       uuid.UUID
       MessageID    *uuid.UUID
       Type         MediaType // image, video, audio, document
       FileName     string
       FileSize     int64
       MimeType     string
       URL          string
       ThumbnailURL string
       Metadata     MediaMetadata
       EncryptionKey []byte
       CreatedAt    time.Time
   }
   ```

2. Storage strategy
   - Local storage for development
   - S3-compatible storage (AWS S3, MinIO, Cloudflare R2)
   - CDN integration
   - Encryption at rest

3. Database schema
   ```sql
   CREATE TABLE media (
       id UUID PRIMARY KEY,
       user_id UUID REFERENCES users(id),
       message_id UUID REFERENCES messages(id),
       type VARCHAR(20) NOT NULL,
       file_name VARCHAR(255),
       file_size BIGINT,
       mime_type VARCHAR(100),
       url TEXT NOT NULL,
       thumbnail_url TEXT,
       metadata JSONB,
       encryption_key BYTEA,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE INDEX idx_media_message ON media(message_id);
   CREATE INDEX idx_media_user ON media(user_id);
   ```

4. Upload handling
   - Multipart upload
   - Chunked upload for large files
   - Image compression
   - Thumbnail generation
   - Virus scanning

5. API endpoints
   - `POST /api/v1/media/upload` - Upload media
   - `GET /api/v1/media/:id` - Get media
   - `DELETE /api/v1/media/:id` - Delete media
   - `POST /api/v1/media/presigned-url` - Get presigned URL

**Deliverables:**
- ✅ Media upload system
- ✅ Storage integration
- ✅ Image/video processing
- ✅ Encrypted media storage

---

### Day 8: Voice Messages & File Sharing

**Tasks:**
1. Voice message handling
   - Audio recording support
   - Audio encoding (Opus codec)
   - Waveform generation
   - Audio duration tracking

2. File sharing
   - Document upload (PDF, DOC, etc.)
   - File preview generation
   - Download tracking
   - Quota management

3. Media features
   - Media gallery
   - Shared media view
   - Media search
   - Auto-download settings

4. API endpoints
   - `POST /api/v1/voice/upload` - Upload voice message
   - `POST /api/v1/files/upload` - Upload document
   - `GET /api/v1/conversations/:id/media` - Get shared media

**Deliverables:**
- ✅ Voice message support
- ✅ File sharing functionality
- ✅ Media gallery
- ✅ Storage optimization

---

## Day 9-10: Solana Wallet Integration

### Day 9: Wallet Core Features

**Tasks:**
1. Wallet domain models
   ```go
   type Wallet struct {
       ID            uuid.UUID
       UserID        uuid.UUID
       Address       string
       Label         string
       Balance       uint64  // lamports
       TokenBalances map[string]uint64
       IsDefault     bool
       CreatedAt     time.Time
   }

   type Transaction struct {
       ID            uuid.UUID
       UserID        uuid.UUID
       Signature     string
       FromAddress   string
       ToAddress     string
       Amount        uint64
       TokenMint     *string
       Type          TransactionType
       Status        TransactionStatus
       Metadata      TransactionMetadata
       CreatedAt     time.Time
   }
   ```

2. Database schema
   ```sql
   CREATE TABLE wallets (
       id UUID PRIMARY KEY,
       user_id UUID REFERENCES users(id),
       address VARCHAR(44) UNIQUE NOT NULL,
       label VARCHAR(100),
       balance BIGINT DEFAULT 0,
       token_balances JSONB,
       is_default BOOLEAN DEFAULT false,
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE transactions (
       id UUID PRIMARY KEY,
       user_id UUID REFERENCES users(id),
       signature VARCHAR(88) UNIQUE NOT NULL,
       from_address VARCHAR(44),
       to_address VARCHAR(44),
       amount BIGINT NOT NULL,
       token_mint VARCHAR(44),
       type VARCHAR(20) NOT NULL,
       status VARCHAR(20) NOT NULL,
       metadata JSONB,
       block_time TIMESTAMP,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE INDEX idx_transactions_user ON transactions(user_id, created_at DESC);
   CREATE INDEX idx_transactions_signature ON transactions(signature);
   ```

3. Solana RPC integration
   ```go
   - Balance checking
   - Transaction history
   - Token account detection
   - Transaction monitoring
   - Confirmation tracking
   ```

4. API endpoints
   - `GET /api/v1/wallet/balance` - Get wallet balance
   - `GET /api/v1/wallet/tokens` - Get token balances
   - `GET /api/v1/wallet/transactions` - Transaction history
   - `POST /api/v1/wallet/prepare-transfer` - Prepare transfer
   - `POST /api/v1/wallet/confirm-transfer` - Confirm transfer

**Deliverables:**
- ✅ Wallet balance tracking
- ✅ Transaction history
- ✅ Solana RPC integration
- ✅ Multi-wallet support

---

### Day 10: In-App Payments & Transfers

**Tasks:**
1. P2P payments in chat
   ```go
   type PaymentRequest struct {
       ID              uuid.UUID
       ConversationID  uuid.UUID
       FromUserID      uuid.UUID
       ToUserID        uuid.UUID
       Amount          uint64
       TokenMint       *string
       Message         string
       Status          PaymentStatus
       TransactionSig  *string
       ExpiresAt       time.Time
       CreatedAt       time.Time
   }
   ```

2. Payment features
   - Send SOL/tokens in chat
   - Payment requests
   - Payment confirmation
   - Payment history
   - QR code generation

3. Database schema
   ```sql
   CREATE TABLE payment_requests (
       id UUID PRIMARY KEY,
       conversation_id UUID REFERENCES conversations(id),
       message_id UUID REFERENCES messages(id),
       from_user_id UUID REFERENCES users(id),
       to_user_id UUID REFERENCES users(id),
       amount BIGINT NOT NULL,
       token_mint VARCHAR(44),
       message TEXT,
       status VARCHAR(20) NOT NULL,
       transaction_sig VARCHAR(88),
       expires_at TIMESTAMP,
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW()
   );

   CREATE INDEX idx_payment_requests_user ON payment_requests(to_user_id, status);
   ```

4. WebSocket events for payments
   ```json
   {
     "type": "payment.received",
     "payload": {
       "transaction_sig": "...",
       "amount": 1000000,
       "from": "..."
     }
   }
   ```

5. API endpoints
   - `POST /api/v1/payments/send` - Send payment
   - `POST /api/v1/payments/request` - Request payment
   - `POST /api/v1/payments/:id/accept` - Accept payment request
   - `GET /api/v1/payments/history` - Payment history

**Deliverables:**
- ✅ In-app payment system
- ✅ Payment requests
- ✅ Transaction notifications
- ✅ Payment UI integration

---

## Day 11-12: Security & Encryption

### Day 11: End-to-End Encryption

**Tasks:**
1. Encryption architecture
   - Signal Protocol implementation
   - X25519 key exchange
   - Double Ratchet algorithm
   - Message keys rotation

2. Key management
   ```go
   type IdentityKey struct {
       UserID      uuid.UUID
       PublicKey   []byte
       PrivateKey  []byte  // Stored client-side only
       SignedBy    string  // Solana signature
       CreatedAt   time.Time
   }

   type PreKey struct {
       ID          uint32
       UserID      uuid.UUID
       PublicKey   []byte
       PrivateKey  []byte
       Used        bool
       CreatedAt   time.Time
   }
   ```

3. Database schema
   ```sql
   CREATE TABLE identity_keys (
       user_id UUID PRIMARY KEY REFERENCES users(id),
       public_key BYTEA NOT NULL,
       signed_by TEXT NOT NULL,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE pre_keys (
       id SERIAL,
       user_id UUID REFERENCES users(id),
       key_id INT NOT NULL,
       public_key BYTEA NOT NULL,
       used BOOLEAN DEFAULT false,
       created_at TIMESTAMP DEFAULT NOW(),
       PRIMARY KEY (user_id, key_id)
   );

   CREATE TABLE sessions (
       id UUID PRIMARY KEY,
       user_id UUID REFERENCES users(id),
       peer_id UUID REFERENCES users(id),
       session_state BYTEA NOT NULL,
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW(),
       UNIQUE(user_id, peer_id)
   );
   ```

4. Encryption flows
   - Session establishment
   - Message encryption/decryption
   - Key exchange
   - Forward secrecy
   - Device verification

5. API endpoints
   - `POST /api/v1/keys/identity` - Upload identity key
   - `GET /api/v1/keys/:userId/bundle` - Get key bundle
   - `POST /api/v1/keys/prekeys` - Upload pre-keys
   - `GET /api/v1/sessions/:peerId` - Get session

**Deliverables:**
- ✅ E2E encryption implementation
- ✅ Key exchange system
- ✅ Session management
- ✅ Forward secrecy

---

### Day 12: Security Features & Privacy

**Tasks:**
1. Security features
   - Message disappearing (self-destruct)
   - Screenshot detection
   - Screen lock with PIN/biometric
   - Hidden chats
   - Two-step verification

2. Privacy settings
   ```go
   type PrivacySettings struct {
       UserID                  uuid.UUID
       ProfilePhotoVisibility  Visibility
       LastSeenVisibility      Visibility
       StatusVisibility        Visibility
       ReadReceiptsEnabled     bool
       TypingIndicatorEnabled  bool
       BlockedUsers            []uuid.UUID
   }
   ```

3. Database schema
   ```sql
   CREATE TABLE privacy_settings (
       user_id UUID PRIMARY KEY REFERENCES users(id),
       profile_photo_visibility VARCHAR(20) DEFAULT 'everyone',
       last_seen_visibility VARCHAR(20) DEFAULT 'everyone',
       status_visibility VARCHAR(20) DEFAULT 'everyone',
       read_receipts_enabled BOOLEAN DEFAULT true,
       typing_indicator_enabled BOOLEAN DEFAULT true,
       updated_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE blocked_users (
       user_id UUID REFERENCES users(id),
       blocked_user_id UUID REFERENCES users(id),
       blocked_at TIMESTAMP DEFAULT NOW(),
       PRIMARY KEY (user_id, blocked_user_id)
   );

   CREATE TABLE disappearing_messages_config (
       conversation_id UUID PRIMARY KEY REFERENCES conversations(id),
       duration_seconds INT NOT NULL,
       enabled_at TIMESTAMP DEFAULT NOW()
   );
   ```

4. Rate limiting & abuse prevention
   - Request rate limiting (Redis)
   - Message spam detection
   - Account abuse prevention
   - IP blocking

5. API endpoints
   - `PUT /api/v1/settings/privacy` - Update privacy settings
   - `POST /api/v1/users/block` - Block user
   - `PUT /api/v1/conversations/:id/disappearing` - Enable disappearing messages
   - `POST /api/v1/security/2fa` - Setup 2FA

**Deliverables:**
- ✅ Disappearing messages
- ✅ Privacy controls
- ✅ User blocking
- ✅ Rate limiting
- ✅ Security headers

---

## Day 13: Advanced Features

### Day 13: Advanced Messaging Features

**Tasks:**
1. Message features
   - Message reactions (emojis)
   - Message forwarding
   - Message pinning
   - Message search (full-text)
   - Mentions (@username)
   - Hashtags

2. Database schema
   ```sql
   CREATE TABLE message_reactions (
       id UUID PRIMARY KEY,
       message_id UUID REFERENCES messages(id),
       user_id UUID REFERENCES users(id),
       emoji VARCHAR(10) NOT NULL,
       created_at TIMESTAMP DEFAULT NOW(),
       UNIQUE(message_id, user_id, emoji)
   );

   CREATE TABLE pinned_messages (
       conversation_id UUID REFERENCES conversations(id),
       message_id UUID REFERENCES messages(id),
       pinned_by UUID REFERENCES users(id),
       pinned_at TIMESTAMP DEFAULT NOW(),
       PRIMARY KEY (conversation_id, message_id)
   );

   CREATE EXTENSION IF NOT EXISTS pg_trgm;
   CREATE INDEX idx_messages_content_search ON messages
       USING gin(to_tsvector('english', content));
   ```

3. Status & Stories
   ```go
   type Status struct {
       ID        uuid.UUID
       UserID    uuid.UUID
       MediaID   uuid.UUID
       Caption   string
       ExpiresAt time.Time
       CreatedAt time.Time
   }

   type StatusView struct {
       StatusID  uuid.UUID
       ViewerID  uuid.UUID
       ViewedAt  time.Time
   }
   ```

4. Database schema
   ```sql
   CREATE TABLE statuses (
       id UUID PRIMARY KEY,
       user_id UUID REFERENCES users(id),
       media_id UUID REFERENCES media(id),
       caption TEXT,
       privacy VARCHAR(20) DEFAULT 'contacts',
       view_count INT DEFAULT 0,
       expires_at TIMESTAMP NOT NULL,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE status_views (
       status_id UUID REFERENCES statuses(id),
       viewer_id UUID REFERENCES users(id),
       viewed_at TIMESTAMP DEFAULT NOW(),
       PRIMARY KEY (status_id, viewer_id)
   );

   CREATE INDEX idx_statuses_expires ON statuses(expires_at)
       WHERE expires_at > NOW();
   ```

5. Contact management
   - Contact list
   - Contact sync
   - Invite system
   - User search

6. API endpoints
   - `POST /api/v1/messages/:id/react` - React to message
   - `POST /api/v1/messages/search` - Search messages
   - `POST /api/v1/statuses` - Create status
   - `GET /api/v1/statuses/feed` - Get status feed
   - `POST /api/v1/contacts/sync` - Sync contacts

**Deliverables:**
- ✅ Message reactions
- ✅ Message search
- ✅ Status/Stories feature
- ✅ Contact management
- ✅ Advanced messaging UI

---

## Day 14: Testing, Optimization & Deployment

### Day 14: Final Integration & Production Readiness

**Tasks:**

1. **Testing Suite**
   ```bash
   # Unit tests
   - Domain layer tests
   - Use case tests
   - Repository tests
   - Handler tests

   # Integration tests
   - API endpoint tests
   - WebSocket tests
   - Database tests
   - Solana integration tests

   # Load testing
   - Concurrent WebSocket connections
   - Message throughput
   - Database query optimization
   ```

2. **Performance Optimization**
   - Database indexing review
   - Query optimization with EXPLAIN ANALYZE
   - Redis caching strategy
     - User sessions
     - Message cache (recent messages)
     - Presence cache
     - Rate limiting
   - Connection pooling tuning
   - WebSocket scaling

3. **Monitoring & Observability**
   ```go
   - Prometheus metrics
   - Grafana dashboards
   - Log aggregation (ELK/Loki)
   - Error tracking (Sentry)
   - Performance monitoring
   - Alert rules
   ```

4. **Database Optimization**
   ```sql
   -- Additional indexes for performance
   CREATE INDEX CONCURRENTLY idx_messages_unread
       ON messages(conversation_id, created_at)
       WHERE status != 'read';

   CREATE INDEX CONCURRENTLY idx_conversations_last_message
       ON conversations(last_message_at DESC);

   -- Partitioning for large tables
   CREATE TABLE messages_partitioned (LIKE messages)
       PARTITION BY RANGE (created_at);
   ```

5. **Security Hardening**
   - Security audit checklist
   - OWASP Top 10 validation
   - SQL injection prevention
   - XSS protection
   - CSRF tokens
   - Rate limiting fine-tuning
   - DDoS protection
   - SSL/TLS configuration

6. **Documentation**
   ```markdown
   - API documentation (Swagger/OpenAPI)
   - Architecture documentation
   - Database schema documentation
   - Deployment guide
   - Environment setup guide
   - WebSocket protocol documentation
   - Security best practices
   ```

7. **Docker & Deployment**
   ```dockerfile
   # Multi-stage Docker build
   # Docker Compose for development
   # Kubernetes manifests
   # CI/CD pipeline (GitHub Actions)
   # Health checks
   # Graceful shutdown
   ```

8. **Production Checklist**
   - [ ] Environment variables configured
   - [ ] Database migrations applied
   - [ ] SSL certificates installed
   - [ ] Backup strategy implemented
   - [ ] Monitoring alerts configured
   - [ ] Load balancer configured
   - [ ] CDN configured for media
   - [ ] Redis cluster setup
   - [ ] Database replication
   - [ ] Disaster recovery plan

**Deliverables:**
- ✅ Complete test coverage (>80%)
- ✅ Performance benchmarks
- ✅ Monitoring dashboards
- ✅ Production deployment
- ✅ API documentation
- ✅ Security audit completed

---

## Project Structure (Final)

```
soldef/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── errors.go
│   │   ├── message/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── types.go
│   │   ├── conversation/
│   │   ├── wallet/
│   │   ├── media/
│   │   ├── group/
│   │   ├── channel/
│   │   └── payment/
│   ├── usecase/
│   │   ├── auth/
│   │   │   ├── service.go
│   │   │   ├── interface.go
│   │   │   └── solana_verifier.go
│   │   ├── messaging/
│   │   │   ├── service.go
│   │   │   ├── encryption.go
│   │   │   └── delivery.go
│   │   ├── wallet/
│   │   │   ├── service.go
│   │   │   └── transaction_monitor.go
│   │   └── payment/
│   ├── delivery/
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── auth.go
│   │   │   │   ├── message.go
│   │   │   │   ├── wallet.go
│   │   │   │   └── media.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── ratelimit.go
│   │   │   │   └── cors.go
│   │   │   └── router.go
│   │   └── websocket/
│   │       ├── hub.go
│   │       ├── client.go
│   │       ├── handler.go
│   │       └── events.go
│   └── repository/
│       ├── postgres/
│       │   ├── gorm.go              # GORM DB connection
│       │   ├── models.go            # GORM models
│       │   ├── user_repository.go   # User repository implementation
│       │   └── message_repository.go # Message repository implementation
│       ├── redis/
│       │   ├── cache.go
│       │   └── presence.go
│       ├── solana/
│       │   ├── client.go
│       │   ├── verifier.go
│       │   └── transaction.go
│       ├── s3/
│       │   └── storage.go
│       └── encryption/
│           ├── signal.go
│           ├── keys.go
│           └── session.go
├── pkg/
│   ├── logger/
│   │   └── logger.go
│   ├── config/
│   │   └── config.go
│   ├── middleware/
│   │   └── jwt.go
│   └── utils/
│       ├── crypto.go
│       └── validator.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_conversations.up.sql
│   └── ...
├── scripts/
│   ├── generate_keys.sh
│   └── migrate.sh
├── docker/
│   ├── Dockerfile
│   ├── Dockerfile.prod
│   └── docker-compose.yml
├── docs/
│   ├── api/
│   │   └── openapi.yaml
│   ├── architecture/
│   │   └── system-design.md
│   └── deployment/
│       └── kubernetes.md
├── tests/
│   ├── integration/
│   ├── e2e/
│   └── load/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── deploy.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── .env.example
```

---

## Key Technologies & Libraries

### Core Backend
```go
// HTTP Framework
github.com/gin-gonic/gin v1.10.0

// Database (ORM)
gorm.io/gorm v1.25.5
gorm.io/driver/postgres v1.5.4
github.com/golang-migrate/migrate/v4 v4.17.0  // For production migrations

// WebSocket
github.com/gorilla/websocket v1.5.1

// Solana
github.com/gagliardetto/solana-go v1.10.0
github.com/portto/solana-go-sdk v1.26.0

// Cache & Queue
github.com/redis/go-redis/v9 v9.3.0

// JWT & Auth
github.com/golang-jwt/jwt/v5 v5.2.0

// Encryption
golang.org/x/crypto v0.17.0
github.com/FiloSottile/age v1.1.1

// Storage
github.com/aws/aws-sdk-go-v2 v1.24.0
github.com/minio/minio-go/v7 v7.0.66

// Config
github.com/spf13/viper v1.18.0
github.com/joho/godotenv v1.5.1

// Logging
go.uber.org/zap v1.26.0
github.com/rs/zerolog v1.31.0

// Monitoring
github.com/prometheus/client_golang v1.18.0

// Testing
github.com/stretchr/testify v1.8.4
github.com/golang/mock v1.6.0
```

---

## Database Design Summary

### Core Tables
1. **users** - User accounts (wallet-based)
2. **conversations** - Chat containers (direct, group, channel)
3. **messages** - All messages
4. **conversation_participants** - Participants in conversations
5. **groups** - Group metadata
6. **group_members** - Group membership
7. **channels** - Channel metadata
8. **channel_subscribers** - Channel subscriptions
9. **media** - File/media storage references
10. **wallets** - Solana wallet tracking
11. **transactions** - Blockchain transaction history
12. **payment_requests** - In-app payment requests
13. **identity_keys** - E2E encryption identity keys
14. **pre_keys** - E2E encryption pre-keys
15. **sessions** - E2E encryption sessions
16. **privacy_settings** - User privacy preferences
17. **blocked_users** - Block list
18. **message_reactions** - Message reactions
19. **statuses** - Stories/status updates
20. **status_views** - Status view tracking

### Performance Optimizations
- Indexes on all foreign keys
- Composite indexes for common queries
- Partial indexes for filtered queries
- Table partitioning for messages (by date)
- Connection pooling (GORM)
- Read replicas for scaling
- Redis caching layer
- GORM optimizations (PrepareStmt, eager loading)

---

## Security Considerations

### Authentication
- Solana wallet signature verification
- JWT with short expiry + refresh tokens
- Rate limiting on auth endpoints
- Account lockout after failed attempts

### Encryption
- End-to-end encryption using Signal Protocol
- Messages encrypted client-side
- Server never has access to plaintext
- Forward secrecy with rotating keys
- Media files encrypted before upload

### Privacy
- No phone numbers required
- Wallet addresses as identity
- Optional usernames
- Privacy settings per user
- Message disappearing feature
- No message content indexing

### Network Security
- HTTPS only (TLS 1.3)
- WebSocket over TLS (WSS)
- CORS configuration
- Security headers (CSP, HSTS, etc.)
- DDoS protection
- Rate limiting per IP/user

### Data Protection
- Encrypted database backups
- Regular security audits
- No sensitive data logging
- GDPR compliance
- Right to deletion
- Data export capability

---

## Scaling Strategy

### Horizontal Scaling
- Stateless API servers
- Load balancer (Nginx/HAProxy)
- WebSocket sticky sessions
- Redis for shared state
- Message queue for async tasks

### Database Scaling
- Read replicas
- Connection pooling
- Query optimization
- Table partitioning
- Archived message storage

### WebSocket Scaling
- Redis Pub/Sub for message broadcasting
- Multiple WebSocket servers
- Consistent hashing for routing
- Graceful connection migration

### Media Scaling
- CDN for media delivery
- Multiple storage regions
- Image resizing service
- Lazy loading
- Compression

---

## Monitoring & Alerting

### Metrics to Track
- API response times
- WebSocket connection count
- Message delivery latency
- Database query performance
- Error rates
- Solana RPC latency
- Active users
- Message throughput

### Alerts
- High error rate
- Database connection pool exhaustion
- High CPU/memory usage
- Slow queries (>1s)
- Failed Solana RPC calls
- Disk space low
- SSL certificate expiry

---

## Deployment Architecture

```
┌─────────────────────────────────────────────┐
│           Load Balancer (Nginx)             │
└──────────────┬──────────────────────────────┘
               │
      ┌────────┴────────┐
      │                 │
┌─────▼────┐      ┌─────▼────┐
│ API      │      │ API      │
│ Server 1 │      │ Server 2 │
└─────┬────┘      └─────┬────┘
      │                 │
      └────────┬────────┘
               │
      ┌────────┴────────┐
      │                 │
┌─────▼────┐      ┌─────▼────────┐
│PostgreSQL│      │   Redis      │
│(Primary) │      │   Cluster    │
└─────┬────┘      └──────────────┘
      │
┌─────▼────────┐
│PostgreSQL    │
│(Replica)     │
└──────────────┘

┌──────────────┐      ┌──────────────┐
│   S3/CDN     │      │Solana RPC    │
│  (Media)     │      │   Nodes      │
└──────────────┘      └──────────────┘
```

---

## Cost Estimation (Monthly)

### Development Phase
- Development servers: $50
- Database (Postgres): $50
- Redis: $30
- S3 storage (100GB): $3
- Solana devnet: Free
**Total: ~$133/month**

### Production (Small Scale - 1K users)
- 2x API servers: $100
- Database (managed): $200
- Redis (managed): $100
- CDN + S3 (500GB): $50
- Monitoring: $50
- Solana RPC: $50
**Total: ~$550/month**

### Production (Medium Scale - 10K users)
- 4x API servers: $400
- Database (HA): $500
- Redis cluster: $300
- CDN + S3 (5TB): $200
- Monitoring: $100
- Solana RPC: $200
**Total: ~$1,700/month**

---

## Success Metrics

### Technical KPIs
- Message delivery latency: <500ms
- API response time p95: <300ms
- WebSocket uptime: >99.9%
- Database query p95: <100ms
- Test coverage: >80%

### Business KPIs
- Daily active users
- Messages sent per day
- Transaction volume (SOL)
- User retention rate
- Average session duration

---

## Next Steps After Day 14

### Phase 2 Features (Weeks 3-4)
- Mobile push notifications
- Video/voice calls (WebRTC)
- Message sync across devices
- Bot API for developers
- NFT integration (profile pictures, stickers)
- Token-gated channels
- DAO integration

### Infrastructure
- Kubernetes deployment
- Multi-region setup
- Advanced analytics
- Machine learning spam detection
- Content moderation AI

### Mobile Development
- React Native app
- Native iOS app (Swift)
- Native Android app (Kotlin)
- Desktop app (Electron/Tauri)

---

## Risk Mitigation

### Technical Risks
- **Solana RPC downtime**: Multiple RPC providers (Alchemy, QuickNode)
- **Database bottleneck**: Sharding + read replicas
- **WebSocket scaling**: Redis Pub/Sub + horizontal scaling
- **Storage costs**: Compression + archival strategy

### Security Risks
- **Key compromise**: Key rotation strategy
- **DoS attacks**: Rate limiting + DDoS protection
- **Data breach**: Encryption at rest + audit logs
- **Spam**: Rate limits + reputation system

### Business Risks
- **User adoption**: Marketing + referral system
- **Regulatory**: Legal compliance review
- **Competition**: Unique Solana features + UX focus

---

## Conclusion

This 14-day plan provides a complete foundation for an enterprise-grade Solana-based secure messaging platform. The architecture is designed for scalability, security, and maintainability.

**Key Differentiators:**
1. Wallet-based identity (no phone numbers)
2. Built-in Solana payments
3. End-to-end encryption
4. Clean architecture
5. Production-ready from day one

**Technologies Used:**
- Go + Gin (high performance HTTP framework)
- PostgreSQL + GORM (powerful ORM with type safety)
- WebSocket (real-time bidirectional communication)
- Solana blockchain (payments + wallet-based identity)
- Signal Protocol (E2E encryption)
- Clean Architecture (maintainable & testable)

Ready to revolutionize messaging with Web3! 🚀
