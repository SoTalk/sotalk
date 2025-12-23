package privacy

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the privacy use case interface
type Service interface {
	// Privacy Settings
	GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*dto.PrivacySettingsResponse, error)
	UpdatePrivacySettings(ctx context.Context, userID uuid.UUID, req *dto.PrivacySettingsRequest) (*dto.PrivacySettingsResponse, error)

	// User Blocking
	BlockUser(ctx context.Context, userID uuid.UUID, req *dto.BlockUserRequest) error
	UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error
	IsUserBlocked(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error)
	GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]*dto.BlockedUserResponse, error)

	// Disappearing Messages
	SetDisappearingMessages(ctx context.Context, userID uuid.UUID, req *dto.DisappearingMessagesRequest) (*dto.DisappearingMessagesResponse, error)
	GetDisappearingMessagesConfig(ctx context.Context, conversationID uuid.UUID) (*dto.DisappearingMessagesResponse, error)
	DisableDisappearingMessages(ctx context.Context, userID, conversationID uuid.UUID) error

	// Two-Factor Authentication
	SetupTwoFactorAuth(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSetupResponse, error)
	VerifyAndEnableTwoFactor(ctx context.Context, userID uuid.UUID, req *dto.TwoFactorVerifyRequest) error
	VerifyTwoFactorCode(ctx context.Context, userID uuid.UUID, req *dto.TwoFactorVerifyRequest) (bool, error)
	DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID, req *dto.TwoFactorVerifyRequest) error
	GetTwoFactorStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorStatusResponse, error)
}
