package request

// GenerateWalletRequest is the HTTP request for generating a new wallet
type GenerateWalletRequest struct {
	Username     string  `json:"username" binding:"required,min=3,max=50"`
	ReferralCode *string `json:"referral_code,omitempty" binding:"omitempty,len=8"`
}

// ChallengeRequest is the HTTP request for requesting a challenge
type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

// VerifySignatureRequest is the HTTP request for verifying a signed challenge
type VerifySignatureRequest struct {
	WalletAddress string  `json:"wallet_address" binding:"required"`
	Signature     string  `json:"signature" binding:"required"`
	Message       string  `json:"message" binding:"required"`
	Username      *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"` // For new users
	ReferralCode  *string `json:"referral_code,omitempty" binding:"omitempty,len=8"`    // For new users
}

// RefreshTokenRequest is the HTTP request for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
