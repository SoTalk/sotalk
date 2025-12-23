package response

import "time"

// UserProfileResponse is the HTTP response for user profile
type UserProfileResponse struct {
	ID            string    `json:"id"`
	WalletAddress string    `json:"wallet_address"`
	Username      string    `json:"username"`
	Avatar        *string   `json:"avatar,omitempty"`
	Bio           *string   `json:"bio,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// PublicUserResponse is the HTTP response for public user information
type PublicUserResponse struct {
	ID            string  `json:"id"`
	WalletAddress string  `json:"wallet_address"`
	Username      string  `json:"username"`
	Avatar        *string `json:"avatar,omitempty"`
	Bio           *string `json:"bio,omitempty"`
	Status        string  `json:"status"`
}

// UserPreferencesResponse is the HTTP response for user preferences
type UserPreferencesResponse struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
}

// CheckWalletExistsResponse is the HTTP response for checking if wallet exists
type CheckWalletExistsResponse struct {
	Exists bool                `json:"exists"`
	User   *PublicUserResponse `json:"user,omitempty"`
}

// SendInvitationResponse is the HTTP response for sending invitation
type SendInvitationResponse struct {
	Message string `json:"message"`
	Sent    bool   `json:"sent"`
}
