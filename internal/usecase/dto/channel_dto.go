package dto

import "time"

// CreateChannelRequest is the request DTO for creating a channel
type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required"`
	Username    string `json:"username" binding:"required"` // Unique channel username
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// CreateChannelResponse is the response DTO for creating a channel
type CreateChannelResponse struct {
	Channel ChannelDTO `json:"channel"`
}

// GetChannelResponse is the response DTO for getting a channel
type GetChannelResponse struct {
	Channel       ChannelDTO `json:"channel"`
	IsSubscribed  bool       `json:"is_subscribed"`
	IsOwner       bool       `json:"is_owner"`
	IsAdmin       bool       `json:"is_admin"`
}

// UpdateChannelRequest is the request DTO for updating a channel
type UpdateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// GetChannelsResponse is the response DTO for getting channels
type GetChannelsResponse struct {
	Channels []ChannelDTO `json:"channels"`
}

// GetSubscriptionsResponse is the response DTO for getting user's subscriptions
type GetSubscriptionsResponse struct {
	Channels []ChannelDTO `json:"channels"`
}

// AddAdminRequest is the request DTO for adding an admin
type AddAdminRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// UpdateAdminPermissionsRequest is the request DTO for updating admin permissions
type UpdateAdminPermissionsRequest struct {
	CanPostMessages   bool `json:"can_post_messages"`
	CanEditMessages   bool `json:"can_edit_messages"`
	CanDeleteMessages bool `json:"can_delete_messages"`
	CanManageAdmins   bool `json:"can_manage_admins"`
}

// ChannelDTO is the channel data transfer object
type ChannelDTO struct {
	ID              string           `json:"id"`
	ConversationID  string           `json:"conversation_id"`
	Name            string           `json:"name"`
	Username        string           `json:"username"`
	Description     string           `json:"description"`
	Avatar          string           `json:"avatar,omitempty"`
	OwnerID         string           `json:"owner_id"`
	IsPublic        bool             `json:"is_public"`
	SubscriberCount int              `json:"subscriber_count"`
	Settings        *ChannelSettings `json:"settings,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// ChannelSettings represents channel settings in DTO
type ChannelSettings struct {
	AdminsCanPost     bool `json:"admins_can_post"`
	LinkPreview       bool `json:"link_preview"`
	ForwardingAllowed bool `json:"forwarding_allowed"`
}

// SubscriberDTO is the subscriber data transfer object
type SubscriberDTO struct {
	UserID               string    `json:"user_id"`
	User                 *UserDTO  `json:"user,omitempty"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
	SubscribedAt         time.Time `json:"subscribed_at"`
}

// ChannelAdminDTO is the channel admin data transfer object
type ChannelAdminDTO struct {
	UserID      string                  `json:"user_id"`
	User        *UserDTO                `json:"user,omitempty"`
	Permissions *ChannelAdminPermissions `json:"permissions,omitempty"`
	AddedAt     time.Time               `json:"added_at"`
}

// ChannelAdminPermissions represents admin permissions in DTO
type ChannelAdminPermissions struct {
	CanPostMessages   bool `json:"can_post_messages"`
	CanEditMessages   bool `json:"can_edit_messages"`
	CanDeleteMessages bool `json:"can_delete_messages"`
	CanManageAdmins   bool `json:"can_manage_admins"`
}
