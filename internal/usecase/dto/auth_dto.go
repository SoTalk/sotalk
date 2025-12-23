package dto

import "time"

// GenerateWalletRequest is the request for generating a new wallet
type GenerateWalletRequest struct {
	Username     string  `json:"username" validate:"required,min=3,max=50"`
	ReferralCode *string `json:"referral_code,omitempty"`
}

// GenerateWalletResponse is the response for generating a new wallet
type GenerateWalletResponse struct {
	Mnemonic      string    `json:"mnemonic"`       // 12-word mnemonic phrase
	WalletAddress string    `json:"wallet_address"` // Solana public key (base58)
	PublicKey     string    `json:"public_key"`     // Same as wallet address
	AccessToken   string    `json:"access_token"`
	RefreshToken  string    `json:"refresh_token"`
	ExpiresAt     time.Time `json:"expires_at"`
	User          UserDTO   `json:"user"`
}

// ChallengeRequest is the request for generating a challenge
type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address" validate:"required"`
}

// ChallengeResponse is the response containing the challenge to sign
type ChallengeResponse struct {
	Challenge string    `json:"challenge"`
	ExpiresAt time.Time `json:"expires_at"`
}

// VerifySignatureRequest is the request for verifying a signed challenge
type VerifySignatureRequest struct {
	WalletAddress string  `json:"wallet_address" validate:"required"`
	Signature     string  `json:"signature" validate:"required"` // base58 encoded signature
	Message       string  `json:"message" validate:"required"`   // the challenge message that was signed
	Username      *string `json:"username,omitempty"`            // Username for new users
	ReferralCode  *string `json:"referral_code,omitempty"`       // Referral code for new users
}

// VerifySignatureResponse is the response after successful signature verification
type VerifySignatureResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserDTO   `json:"user"`
}

// RefreshTokenRequest is the request for refreshing access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse is the response for refreshing access token
type RefreshTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// UserDTO is the data transfer object for user
type UserDTO struct {
	ID            string    `json:"id"`
	WalletAddress string    `json:"wallet_address"`
	Username      string    `json:"username"`
	Avatar        *string   `json:"avatar,omitempty"`
	PublicKey     string    `json:"public_key"`
	Status        string    `json:"status"`
	IsOnline      bool      `json:"is_online"`
	LastSeen      time.Time `json:"last_seen"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
