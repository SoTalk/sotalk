package response

import "time"

// CreateChannelResponse is the HTTP response for creating a channel
type CreateChannelResponse struct {
	Channel ChannelDTO `json:"channel"`
}

// GetChannelResponse is the HTTP response for getting a channel
type GetChannelResponse struct {
	Channel      ChannelDTO `json:"channel"`
	IsSubscribed bool       `json:"is_subscribed"`
	IsOwner      bool       `json:"is_owner"`
	IsAdmin      bool       `json:"is_admin"`
}

// GetChannelsResponse is the HTTP response for getting channels
type GetChannelsResponse struct {
	Channels []ChannelDTO `json:"channels"`
}

// ChannelDTO is the channel data in response
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

// ChannelSettings represents channel settings in response
type ChannelSettings struct {
	AdminsCanPost     bool `json:"admins_can_post"`
	LinkPreview       bool `json:"link_preview"`
	ForwardingAllowed bool `json:"forwarding_allowed"`
}
