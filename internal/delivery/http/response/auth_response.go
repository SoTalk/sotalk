package response

import "time"

// GenerateWalletResponse is the HTTP response for generating a new wallet
type GenerateWalletResponse struct {
	Mnemonic      string    `json:"mnemonic"`       // 12-word mnemonic phrase (shown ONCE!)
	WalletAddress string    `json:"wallet_address"` // Solana public key
	PublicKey     string    `json:"public_key"`     // Same as wallet address
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	ExpiresAt     time.Time `json:"expires_at"`
	User          UserDTO   `json:"user"`
}

// ChallengeResponse is the HTTP response containing the challenge to sign
type ChallengeResponse struct {
	Challenge string    `json:"challenge"`
	ExpiresAt time.Time `json:"expires_at"`
}

// VerifySignatureResponse is the HTTP response after successful signature verification
type VerifySignatureResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserDTO   `json:"user"`
}

// RefreshTokenResponse is the HTTP response for refreshing token
type RefreshTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// UserDTO is the user data in response
type UserDTO struct {
	ID            string     `json:"id"`
	WalletAddress string     `json:"wallet_address"`
	Username      string     `json:"username"`
	Avatar        *string    `json:"avatar,omitempty"`
	Status        string     `json:"status"`
	IsOnline      bool       `json:"is_online"`
	LastSeen      *time.Time `json:"last_seen,omitempty"`
	// Privacy fields - whether current user can see these fields
	CanSeeAvatar   *bool `json:"can_see_avatar,omitempty"`
	CanSeeLastSeen *bool `json:"can_see_last_seen,omitempty"`
	CanSeeStatus   *bool `json:"can_see_status,omitempty"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse is the standard success response
type SuccessResponse struct {
	Message string `json:"message"`
}
