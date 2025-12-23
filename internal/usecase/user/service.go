package user

import (
	"context"

	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the user service interface
type Service interface {
	// Profile management
	GetProfile(ctx context.Context, userID string) (*dto.GetProfileResponse, error)
	UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error)

	// Preferences
	GetPreferences(ctx context.Context, userID string) (*dto.GetPreferencesResponse, error)
	UpdatePreferences(ctx context.Context, req *dto.UpdatePreferencesRequest) (*dto.UpdatePreferencesResponse, error)

	// Get user by ID (public information)
	GetUserByID(ctx context.Context, userID string) (*dto.GetUserResponse, error)

	// Check if wallet address exists
	CheckWalletExists(ctx context.Context, walletAddress string) (bool, *dto.GetUserResponse, error)

	// Send invitation to wallet address
	SendInvitation(ctx context.Context, req *dto.SendInvitationRequest) error
}
