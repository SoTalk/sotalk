package privacy

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for privacy data operations
type Repository interface {
	// Privacy Settings
	CreatePrivacySettings(ctx context.Context, settings *PrivacySettings) error
	GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*PrivacySettings, error)
	UpdatePrivacySettings(ctx context.Context, settings *PrivacySettings) error

	// Blocked Users
	BlockUser(ctx context.Context, blocked *BlockedUser) error
	UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error
	IsUserBlocked(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error)
	GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]*BlockedUser, error)
	GetBlockedByUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) // Who blocked this user

	// Disappearing Messages
	SetDisappearingMessages(ctx context.Context, config *DisappearingMessagesConfig) error
	GetDisappearingMessagesConfig(ctx context.Context, conversationID uuid.UUID) (*DisappearingMessagesConfig, error)
	DisableDisappearingMessages(ctx context.Context, conversationID uuid.UUID) error

	// Two-Factor Authentication
	CreateTwoFactorAuth(ctx context.Context, twoFA *TwoFactorAuth) error
	GetTwoFactorAuth(ctx context.Context, userID uuid.UUID) (*TwoFactorAuth, error)
	UpdateTwoFactorAuth(ctx context.Context, twoFA *TwoFactorAuth) error
	DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID) error
}
