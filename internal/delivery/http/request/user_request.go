package request

// UpdateProfileRequest is the HTTP request for updating user profile
type UpdateProfileRequest struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Avatar   *string `json:"avatar" binding:"omitempty"`
	Bio      *string `json:"bio" binding:"omitempty,max=500"`
}

// UpdatePreferencesRequest is the HTTP request for updating user preferences
type UpdatePreferencesRequest struct {
	Language *string `json:"language" binding:"omitempty"`
	Theme    *string `json:"theme" binding:"omitempty,oneof=light dark system"`
}

// CheckWalletExistsRequest is the HTTP request for checking if wallet exists
type CheckWalletExistsRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

// SendInvitationRequest is the HTTP request for sending invitation
type SendInvitationRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
}
