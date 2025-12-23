package dto

import "time"

// UpdateProfileRequest is the request DTO for updating profile
type UpdateProfileRequest struct {
	UserID   string
	Username *string
	Avatar   *string
	Bio      *string
}

// UpdateProfileResponse is the response DTO for updating profile
type UpdateProfileResponse struct {
	ID            string
	WalletAddress string
	Username      string
	Avatar        *string
	Bio           *string
	Status        string
	CreatedAt     time.Time
}

// GetProfileResponse is the response DTO for getting profile
type GetProfileResponse struct {
	ID            string
	WalletAddress string
	Username      string
	Avatar        *string
	Bio           *string
	Status        string
	CreatedAt     time.Time
	// Activity Stats
	MessageCount     int `json:"message_count"`
	GroupCount       int `json:"group_count"`
	ChannelCount     int `json:"channel_count"`
	ContactCount     int `json:"contact_count"`
	TransactionCount int `json:"transaction_count"`
}

// UpdatePreferencesRequest is the request DTO for updating preferences
type UpdatePreferencesRequest struct {
	UserID               string
	Language             *string
	Theme                *string
	NotificationsEnabled *bool
	SoundEnabled         *bool
	EmailNotifications   *bool
}

// UpdatePreferencesResponse is the response DTO for updating preferences
type UpdatePreferencesResponse struct {
	Language             string
	Theme                string
	NotificationsEnabled bool
	SoundEnabled         bool
	EmailNotifications   bool
}

// GetPreferencesResponse is the response DTO for getting preferences
type GetPreferencesResponse struct {
	Language             string
	Theme                string
	NotificationsEnabled bool
	SoundEnabled         bool
	EmailNotifications   bool
}

// GetUserResponse is the response DTO for getting user by ID
type GetUserResponse struct {
	ID            string
	WalletAddress string
	Username      string
	Avatar        *string
	Bio           *string
	Status        string
}

// CheckWalletExistsResponse is the response DTO for checking if wallet exists
type CheckWalletExistsResponse struct {
	Exists bool
	User   *GetUserResponse
}

// SendInvitationRequest is the request DTO for sending invitation
type SendInvitationRequest struct {
	SenderID   string
	Email      string
	InviteLink string // Optional: custom invite link
}

// SendInvitationResponse is the response DTO for sending invitation
type SendInvitationResponse struct {
	Message string
	Sent    bool
}
