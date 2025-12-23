package auth

import (
	"context"

	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the authentication use case interface
type Service interface {
	// GenerateWallet generates a new Solana wallet with mnemonic phrase
	GenerateWallet(ctx context.Context, req *dto.GenerateWalletRequest) (*dto.GenerateWalletResponse, error)

	// GenerateChallenge generates a challenge for wallet to sign
	GenerateChallenge(ctx context.Context, walletAddress string) (*dto.ChallengeResponse, error)

	// VerifySignature verifies the signed challenge and authenticates user
	VerifySignature(ctx context.Context, req *dto.VerifySignatureRequest) (*dto.VerifySignatureResponse, error)

	// RefreshToken refreshes the access token using refresh token
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)

	// GetCurrentUser gets the current user from token
	GetCurrentUser(ctx context.Context, userID string) (*dto.UserDTO, error)

	// Logout invalidates the current session
	Logout(ctx context.Context, userID string, token string) error

	// DeleteAccount permanently deletes the user account
	DeleteAccount(ctx context.Context, userID string) error
}
