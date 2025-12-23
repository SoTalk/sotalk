package postgres

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the GORM model for users table
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WalletAddress string         `gorm:"type:varchar(44);uniqueIndex;not null"`
	Username      string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	Avatar        *string        `gorm:"type:text"`
	Bio           *string        `gorm:"type:varchar(500)"`
	PublicKey     string         `gorm:"type:text;not null"`
	ReferralCode  string         `gorm:"type:varchar(8);index"`
	Status        string         `gorm:"type:varchar(20);default:'offline'"`
	LastSeen      time.Time      `gorm:"type:timestamp;not null"`
	CreatedAt     time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook is called before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.ReferralCode == "" {
		u.ReferralCode = generateReferralCode()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = time.Now()
	}
	if u.LastSeen.IsZero() {
		u.LastSeen = time.Now()
	}
	return nil
}

// generateReferralCode generates a unique 8-character referral code
func generateReferralCode() string {
	b := make([]byte, 5)
	rand.Read(b)
	code := base32.StdEncoding.EncodeToString(b)[:8]
	return strings.ToUpper(strings.TrimRight(code, "="))
}

// BeforeUpdate hook is called before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// Conversation is the GORM model for conversations table
type Conversation struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type          string     `gorm:"type:varchar(20);not null"`
	LastMessageID *uuid.UUID `gorm:"type:uuid"`
	LastMessageAt *time.Time `gorm:"type:timestamp"`
	CreatedAt     time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Conversation model
func (Conversation) TableName() string {
	return "conversations"
}

// BeforeCreate hook for Conversation
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Conversation
func (c *Conversation) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// ConversationParticipant is the GORM model for conversation_participants table
type ConversationParticipant struct {
	ConversationID uuid.UUID  `gorm:"type:uuid;primaryKey;not null"`
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey;not null"`
	Role           string     `gorm:"type:varchar(20);default:'member'"`
	JoinedAt       time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	LastReadAt     *time.Time `gorm:"type:timestamp"`
	ArchivedAt     *time.Time `gorm:"type:timestamp"` // For archiving conversations per user
}

// TableName specifies the table name for ConversationParticipant model
func (ConversationParticipant) TableName() string {
	return "conversation_participants"
}

// Message is the GORM model for messages table
type Message struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID  `gorm:"type:uuid;not null;index:idx_messages_conversation"`
	SenderID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Content        string     `gorm:"type:text;not null"`
	ContentType    string     `gorm:"type:varchar(20);not null"`
	Signature      string     `gorm:"type:text"`
	ReplyToID      *uuid.UUID `gorm:"type:uuid"`
	Status         string     `gorm:"type:varchar(20);default:'sending'"`
	CreatedAt      time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;index:idx_messages_conversation"`
	UpdatedAt      time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Message model
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate hook for Message
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Message
func (m *Message) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// Group is the GORM model for groups table
type Group struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Name           string    `gorm:"type:varchar(100);not null"`
	Description    string    `gorm:"type:text"`
	Avatar         string    `gorm:"type:text"`
	CreatorID      uuid.UUID `gorm:"type:uuid;not null;index"`
	MaxMembers     int       `gorm:"type:int;default:256"`
	Settings       string    `gorm:"type:jsonb"` // JSON for settings
	CreatedAt      time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Group model
func (Group) TableName() string {
	return "groups"
}

// BeforeCreate hook for Group
func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	if g.CreatedAt.IsZero() {
		g.CreatedAt = time.Now()
	}
	if g.UpdatedAt.IsZero() {
		g.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Group
func (g *Group) BeforeUpdate(tx *gorm.DB) error {
	g.UpdatedAt = time.Now()
	return nil
}

// GroupMember is the GORM model for group_members table
type GroupMember struct {
	GroupID     uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	Role        string    `gorm:"type:varchar(20);default:'member'"`
	Permissions string    `gorm:"type:jsonb"` // JSON for permissions
	JoinedAt    time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GroupMember model
func (GroupMember) TableName() string {
	return "group_members"
}

// Channel is the GORM model for channels table
type Channel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Name            string    `gorm:"type:varchar(100);not null"`
	Username        string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description     string    `gorm:"type:text"`
	Avatar          string    `gorm:"type:text"`
	OwnerID         uuid.UUID `gorm:"type:uuid;not null;index"`
	IsPublic        bool      `gorm:"type:boolean;default:true"`
	SubscriberCount int       `gorm:"type:int;default:0"`
	Settings        string    `gorm:"type:jsonb"` // JSON for settings
	CreatedAt       time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Channel model
func (Channel) TableName() string {
	return "channels"
}

// BeforeCreate hook for Channel
func (c *Channel) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Channel
func (c *Channel) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// ChannelSubscriber is the GORM model for channel_subscribers table
type ChannelSubscriber struct {
	ChannelID            uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	UserID               uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	NotificationsEnabled bool      `gorm:"type:boolean;default:true"`
	SubscribedAt         time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for ChannelSubscriber model
func (ChannelSubscriber) TableName() string {
	return "channel_subscribers"
}

// ChannelAdmin is the GORM model for channel_admins table
type ChannelAdmin struct {
	ChannelID   uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey;not null"`
	Permissions string    `gorm:"type:jsonb"` // JSON for permissions
	AddedAt     time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for ChannelAdmin model
func (ChannelAdmin) TableName() string {
	return "channel_admins"
}

// Media is the GORM model for media table
type Media struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	MessageID     *uuid.UUID `gorm:"type:uuid;index"`
	Type          string     `gorm:"type:varchar(20);not null"`
	FileName      string     `gorm:"type:varchar(255);not null"`
	FileSize      int64      `gorm:"type:bigint;not null"`
	MimeType      string     `gorm:"type:varchar(100);not null"`
	URL           string     `gorm:"type:text;not null"`
	ThumbnailURL  string     `gorm:"type:text"`
	Metadata      string     `gorm:"type:jsonb"` // JSON for metadata
	CreatedAt     time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Media model
func (Media) TableName() string {
	return "media"
}

// BeforeCreate hook for Media
func (m *Media) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

// Wallet is the GORM model for wallets table
type Wallet struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	Address       string    `gorm:"type:varchar(44);uniqueIndex;not null"`
	Label         string    `gorm:"type:varchar(100)"`
	Balance       uint64    `gorm:"type:bigint;default:0"`
	TokenBalances string    `gorm:"type:jsonb"` // JSON map of token balances
	IsDefault     bool      `gorm:"type:boolean;default:false"`
	CreatedAt     time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Wallet model
func (Wallet) TableName() string {
	return "wallets"
}

// BeforeCreate hook for Wallet
func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Wallet
func (w *Wallet) BeforeUpdate(tx *gorm.DB) error {
	w.UpdatedAt = time.Now()
	return nil
}

// Transaction is the GORM model for transactions table
type Transaction struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	Signature   string     `gorm:"type:varchar(88);uniqueIndex;not null"`
	FromAddress string     `gorm:"type:varchar(44);not null;index"`
	ToAddress   string     `gorm:"type:varchar(44);not null;index"`
	Amount      uint64     `gorm:"type:bigint;not null"`
	TokenMint   *string    `gorm:"type:varchar(44)"`
	Type        string     `gorm:"type:varchar(20);not null"`
	Status      string     `gorm:"type:varchar(20);not null"`
	Fee         uint64     `gorm:"type:bigint;default:0"`
	BlockTime   *time.Time `gorm:"type:timestamp"`
	Metadata    string     `gorm:"type:jsonb"` // JSON for metadata
	CreatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate hook for Transaction
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Transaction
func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// PaymentRequest is the GORM model for payment_requests table
type PaymentRequest struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID  `gorm:"type:uuid;not null;index"`
	MessageID      *uuid.UUID `gorm:"type:uuid;index"`
	FromUserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	ToUserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Amount         uint64     `gorm:"type:bigint;not null"`
	TokenMint      *string    `gorm:"type:varchar(44)"`
	Message        string     `gorm:"type:text"`
	Status         string     `gorm:"type:varchar(20);not null"`
	TransactionSig *string    `gorm:"type:varchar(88)"`
	ExpiresAt      time.Time  `gorm:"type:timestamp;not null"`
	CreatedAt      time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for PaymentRequest model
func (PaymentRequest) TableName() string {
	return "payment_requests"
}

// BeforeCreate hook for PaymentRequest
func (p *PaymentRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for PaymentRequest
func (p *PaymentRequest) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// PrivacySettings is the GORM model for privacy_settings table (Day 12)
type PrivacySettings struct {
	UserID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProfilePhotoVisibility string    `gorm:"type:varchar(20);not null;default:'everyone'"`
	LastSeenVisibility     string    `gorm:"type:varchar(20);not null;default:'everyone'"`
	StatusVisibility       string    `gorm:"type:varchar(20);not null;default:'everyone'"`
	ReadReceiptsEnabled    bool      `gorm:"type:boolean;not null;default:true"`
	TypingIndicatorEnabled bool      `gorm:"type:boolean;not null;default:true"`
	CreatedAt              time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt              time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for PrivacySettings model
func (PrivacySettings) TableName() string {
	return "privacy_settings"
}

// BeforeCreate hook for PrivacySettings
func (p *PrivacySettings) BeforeCreate(tx *gorm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for PrivacySettings
func (p *PrivacySettings) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// BlockedUser is the GORM model for blocked_users table (Day 12)
type BlockedUser struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_blocked_users_user"`
	BlockedUserID uuid.UUID `gorm:"type:uuid;not null;index:idx_blocked_users_blocked"`
	Reason        string    `gorm:"type:text"`
	BlockedAt     time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for BlockedUser model
func (BlockedUser) TableName() string {
	return "blocked_users"
}

// BeforeCreate hook for BlockedUser
func (b *BlockedUser) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.BlockedAt.IsZero() {
		b.BlockedAt = time.Now()
	}
	return nil
}

// DisappearingMessagesConfig is the GORM model for disappearing_messages_config table (Day 12)
type DisappearingMessagesConfig struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	DurationSeconds int       `gorm:"type:int;not null;default:0"`
	EnabledBy       uuid.UUID `gorm:"type:uuid;not null"`
	EnabledAt       time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for DisappearingMessagesConfig model
func (DisappearingMessagesConfig) TableName() string {
	return "disappearing_messages_config"
}

// BeforeCreate hook for DisappearingMessagesConfig
func (d *DisappearingMessagesConfig) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.EnabledAt.IsZero() {
		d.EnabledAt = time.Now()
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for DisappearingMessagesConfig
func (d *DisappearingMessagesConfig) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now()
	return nil
}

// TwoFactorAuth is the GORM model for two_factor_auth table (Day 12)
type TwoFactorAuth struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	Enabled     bool      `gorm:"type:boolean;not null;default:false"`
	Secret      string    `gorm:"type:text;not null"` // Encrypted TOTP secret
	BackupCodes string    `gorm:"type:text"`          // Encrypted JSON array of backup codes
	EnabledAt   time.Time `gorm:"type:timestamp"`
	LastUsedAt  time.Time `gorm:"type:timestamp"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for TwoFactorAuth model
func (TwoFactorAuth) TableName() string {
	return "two_factor_auth"
}

// BeforeCreate hook for TwoFactorAuth
func (t *TwoFactorAuth) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for TwoFactorAuth
func (t *TwoFactorAuth) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// MessageReaction is the GORM model for message_reactions table (Day 13)
type MessageReaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index:idx_message_reactions"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_reactions"`
	Emoji     string    `gorm:"type:varchar(10);not null"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for MessageReaction model
func (MessageReaction) TableName() string {
	return "message_reactions"
}

// BeforeCreate hook for MessageReaction
func (m *MessageReaction) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

// PinnedMessage is the GORM model for pinned_messages table (Day 13)
// Each user can pin messages independently (composite unique index on message_id + pinned_by)
type PinnedMessage struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index:idx_pinned_conversation"`
	MessageID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_pinned_message_user"`
	PinnedBy       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_pinned_message_user"`
	PinnedAt       time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for PinnedMessage model
func (PinnedMessage) TableName() string {
	return "pinned_messages"
}

// BeforeCreate hook for PinnedMessage
func (p *PinnedMessage) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.PinnedAt.IsZero() {
		p.PinnedAt = time.Now()
	}
	return nil
}

// ForwardedMessage is the GORM model for forwarded_messages table (Day 13)
type ForwardedMessage struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OriginalMessageID    uuid.UUID `gorm:"type:uuid;not null;index:idx_forwarded_original"`
	NewMessageID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	ForwardedBy          uuid.UUID `gorm:"type:uuid;not null;index:idx_forwarded_by"`
	TargetConversationID uuid.UUID `gorm:"type:uuid;not null"`
	ForwardedAt          time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for ForwardedMessage model
func (ForwardedMessage) TableName() string {
	return "forwarded_messages"
}

// BeforeCreate hook for ForwardedMessage
func (f *ForwardedMessage) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.ForwardedAt.IsZero() {
		f.ForwardedAt = time.Now()
	}
	return nil
}

// MessageMention is the GORM model for message_mentions table (Day 13)
type MessageMention struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index:idx_mention_message"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_mention_user"`
	Position  int       `gorm:"type:int;not null"`
	Length    int       `gorm:"type:int;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for MessageMention model
func (MessageMention) TableName() string {
	return "message_mentions"
}

// BeforeCreate hook for MessageMention
func (m *MessageMention) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

// Status is the GORM model for statuses table (Day 13)
type Status struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_status_user"`
	MediaID   uuid.UUID `gorm:"type:uuid;not null"`
	Caption   string    `gorm:"type:text"`
	Privacy   string    `gorm:"type:varchar(20);not null;default:'contacts'"`
	ViewCount int       `gorm:"type:int;not null;default:0"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null;index:idx_status_expires"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Status model
func (Status) TableName() string {
	return "statuses"
}

// BeforeCreate hook for Status
func (s *Status) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	// Default expiry: 24 hours
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	return nil
}

// StatusView is the GORM model for status_views table (Day 13)
type StatusView struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StatusID uuid.UUID `gorm:"type:uuid;not null;index:idx_view_status"`
	ViewerID uuid.UUID `gorm:"type:uuid;not null;index:idx_view_viewer"`
	ViewedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for StatusView model
func (StatusView) TableName() string {
	return "status_views"
}

// BeforeCreate hook for StatusView
func (s *StatusView) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.ViewedAt.IsZero() {
		s.ViewedAt = time.Now()
	}
	return nil
}

// Contact is the GORM model for contacts table (Day 13)
type Contact struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:idx_contact_user"`
	ContactID   uuid.UUID `gorm:"type:uuid;not null;index:idx_contact_contact"`
	DisplayName string    `gorm:"type:varchar(100)"`
	IsFavorite  bool      `gorm:"type:boolean;not null;default:false"`
	IsBlocked   bool      `gorm:"type:boolean;not null;default:false"`
	AddedAt     time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Contact model
func (Contact) TableName() string {
	return "contacts"
}

// BeforeCreate hook for Contact
func (c *Contact) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.AddedAt.IsZero() {
		c.AddedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Contact
func (c *Contact) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// ContactInvite is the GORM model for contact_invites table (Day 13)
type ContactInvite struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SenderID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_invite_sender"`
	RecipientID uuid.UUID  `gorm:"type:uuid;not null;index:idx_invite_recipient"`
	Message     string     `gorm:"type:text"`
	Status      string     `gorm:"type:varchar(20);not null;default:'pending'"`
	ExpiresAt   time.Time  `gorm:"type:timestamp;not null"`
	CreatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	RespondedAt *time.Time `gorm:"type:timestamp"`
}

// TableName specifies the table name for ContactInvite model
func (ContactInvite) TableName() string {
	return "contact_invites"
}

// BeforeCreate hook for ContactInvite
func (c *ContactInvite) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	// Default expiry: 7 days
	if c.ExpiresAt.IsZero() {
		c.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	return nil
}

// Notification is the GORM model for notifications table
type Notification struct {
	ID        uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID              `gorm:"type:uuid;not null;index:idx_notification_user"`
	Type      string                 `gorm:"type:varchar(20);not null"`
	Title     string                 `gorm:"type:varchar(255);not null"`
	Body      string                 `gorm:"type:text;not null"`
	Data      map[string]interface{} `gorm:"type:jsonb"`
	Read      bool                   `gorm:"type:boolean;not null;default:false"`
	CreatedAt time.Time              `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time              `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for Notification model
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate hook for Notification
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Notification
func (n *Notification) BeforeUpdate(tx *gorm.DB) error {
	n.UpdatedAt = time.Now()
	return nil
}

// NotificationSettings is the GORM model for notification_settings table
type NotificationSettings struct {
	UserID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	MessagesEnabled  bool      `gorm:"type:boolean;not null;default:true"`
	GroupsEnabled    bool      `gorm:"type:boolean;not null;default:true"`
	ChannelsEnabled  bool      `gorm:"type:boolean;not null;default:true"`
	PaymentsEnabled  bool      `gorm:"type:boolean;not null;default:true"`
	MentionsEnabled  bool      `gorm:"type:boolean;not null;default:true"`
	ReactionsEnabled bool      `gorm:"type:boolean;not null;default:true"`
	SoundEnabled     bool      `gorm:"type:boolean;not null;default:true"`
	VibrationEnabled bool      `gorm:"type:boolean;not null;default:true"`
	CreatedAt        time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for NotificationSettings model
func (NotificationSettings) TableName() string {
	return "notification_settings"
}

// BeforeCreate hook for NotificationSettings
func (ns *NotificationSettings) BeforeCreate(tx *gorm.DB) error {
	if ns.CreatedAt.IsZero() {
		ns.CreatedAt = time.Now()
	}
	if ns.UpdatedAt.IsZero() {
		ns.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for NotificationSettings
func (ns *NotificationSettings) BeforeUpdate(tx *gorm.DB) error {
	ns.UpdatedAt = time.Now()
	return nil
}

// Referral is the GORM model for referrals table
type Referral struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReferrerID uuid.UUID      `gorm:"type:uuid;not null;index"`
	RefereeID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"`
	Code       string         `gorm:"type:varchar(8);not null;index"`
	Status     string         `gorm:"type:varchar(20);not null;default:'pending'"`
	RewardType *string        `gorm:"type:varchar(20)"`
	RewardAmount *uint64      `gorm:"type:bigint"`
	RewardTxSig  *string      `gorm:"type:varchar(88)"`
	CreatedAt  time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Referral model
func (Referral) TableName() string {
	return "referrals"
}

// BeforeCreate hook for Referral
func (r *Referral) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Referral
func (r *Referral) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// UserPreferences is the GORM model for user_preferences table
type UserPreferences struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Language             string    `gorm:"type:varchar(10);not null;default:'en'"`
	Theme                string    `gorm:"type:varchar(20);not null;default:'dark'"`
	NotificationsEnabled bool      `gorm:"type:boolean;not null;default:true"`
	SoundEnabled         bool      `gorm:"type:boolean;not null;default:true"`
	EmailNotifications   bool      `gorm:"type:boolean;not null;default:true"`
	CreatedAt            time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt            time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for UserPreferences model
func (UserPreferences) TableName() string {
	return "user_preferences"
}

// BeforeCreate hook for UserPreferences
func (up *UserPreferences) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	if up.CreatedAt.IsZero() {
		up.CreatedAt = time.Now()
	}
	if up.UpdatedAt.IsZero() {
		up.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for UserPreferences
func (up *UserPreferences) BeforeUpdate(tx *gorm.DB) error {
	up.UpdatedAt = time.Now()
	return nil
}

// PasskeyCredential is the GORM model for passkey_credentials table
type PasskeyCredential struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_passkey_user_id"`
	CredentialID    string     `gorm:"type:text;not null;uniqueIndex"`
	PublicKey       []byte     `gorm:"type:bytea;not null"`
	AttestationType string     `gorm:"type:text;default:'none'"`
	AAGUID          []byte     `gorm:"type:bytea"`
	SignCount       uint32     `gorm:"type:bigint;default:0"`
	Transports      string     `gorm:"type:text[]"` // PostgreSQL array of strings
	BackupEligible  bool       `gorm:"type:boolean;default:false"`
	BackupState     bool       `gorm:"type:boolean;default:false"`
	CreatedAt       time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	LastUsedAt      *time.Time `gorm:"type:timestamp"`
}

// TableName specifies the table name for PasskeyCredential model
func (PasskeyCredential) TableName() string {
	return "passkey_credentials"
}

// BeforeCreate hook for PasskeyCredential
func (p *PasskeyCredential) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	return nil
}

