package channel

import (
	"time"

	"github.com/google/uuid"
)

// Channel represents a broadcast channel entity
type Channel struct {
	ID              uuid.UUID
	ConversationID  uuid.UUID
	Name            string
	Username        string // Unique channel username (e.g., @soldefnews)
	Description     string
	Avatar          string
	OwnerID         uuid.UUID
	IsPublic        bool
	SubscriberCount int
	Settings        *Settings
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Settings represents channel settings
type Settings struct {
	AdminsCanPost bool   // If false, only owner can post
	LinkPreview   bool   // Show link previews
	ForwardingAllowed bool // Allow forwarding messages
}

// Subscriber represents a channel subscriber
type Subscriber struct {
	ChannelID           uuid.UUID
	UserID              uuid.UUID
	NotificationsEnabled bool
	SubscribedAt        time.Time
}

// Admin represents a channel admin
type Admin struct {
	ChannelID   uuid.UUID
	UserID      uuid.UUID
	Permissions *AdminPermissions
	AddedAt     time.Time
}

// AdminPermissions represents admin permissions
type AdminPermissions struct {
	CanPostMessages  bool
	CanEditMessages  bool
	CanDeleteMessages bool
	CanManageAdmins  bool
}

// NewChannel creates a new channel
func NewChannel(name, username, description string, ownerID uuid.UUID, conversationID uuid.UUID, isPublic bool) *Channel {
	return &Channel{
		ID:              uuid.New(),
		ConversationID:  conversationID,
		Name:            name,
		Username:        username,
		Description:     description,
		OwnerID:         ownerID,
		IsPublic:        isPublic,
		SubscriberCount: 0,
		Settings: &Settings{
			AdminsCanPost:     true,
			LinkPreview:       true,
			ForwardingAllowed: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewSubscriber creates a new channel subscriber
func NewSubscriber(channelID, userID uuid.UUID) *Subscriber {
	return &Subscriber{
		ChannelID:            channelID,
		UserID:               userID,
		NotificationsEnabled: true,
		SubscribedAt:         time.Now(),
	}
}

// NewAdmin creates a new channel admin
func NewAdmin(channelID, userID uuid.UUID) *Admin {
	return &Admin{
		ChannelID: channelID,
		UserID:    userID,
		Permissions: &AdminPermissions{
			CanPostMessages:   true,
			CanEditMessages:   true,
			CanDeleteMessages: true,
			CanManageAdmins:   false,
		},
		AddedAt: time.Now(),
	}
}

// UpdateInfo updates channel information
func (c *Channel) UpdateInfo(name, description, avatar string) {
	if name != "" {
		c.Name = name
	}
	if description != "" {
		c.Description = description
	}
	if avatar != "" {
		c.Avatar = avatar
	}
	c.UpdatedAt = time.Now()
}

// UpdateSettings updates channel settings
func (c *Channel) UpdateSettings(settings *Settings) {
	if settings != nil {
		c.Settings = settings
		c.UpdatedAt = time.Now()
	}
}

// IncrementSubscriberCount increments the subscriber count
func (c *Channel) IncrementSubscriberCount() {
	c.SubscriberCount++
	c.UpdatedAt = time.Now()
}

// DecrementSubscriberCount decrements the subscriber count
func (c *Channel) DecrementSubscriberCount() {
	if c.SubscriberCount > 0 {
		c.SubscriberCount--
	}
	c.UpdatedAt = time.Now()
}

// ToggleNotifications toggles notifications for a subscriber
func (s *Subscriber) ToggleNotifications() {
	s.NotificationsEnabled = !s.NotificationsEnabled
}

// GrantFullPermissions grants full admin permissions
func (a *Admin) GrantFullPermissions() {
	a.Permissions.CanPostMessages = true
	a.Permissions.CanEditMessages = true
	a.Permissions.CanDeleteMessages = true
	a.Permissions.CanManageAdmins = true
}

// RevokeManageAdmins revokes the ability to manage admins
func (a *Admin) RevokeManageAdmins() {
	a.Permissions.CanManageAdmins = false
}

// CanPost checks if admin can post messages
func (a *Admin) CanPost() bool {
	return a.Permissions.CanPostMessages
}
