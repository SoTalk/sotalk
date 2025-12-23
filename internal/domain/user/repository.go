package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for user data operations
type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByWalletAddress(ctx context.Context, walletAddress string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByReferralCode(ctx context.Context, referralCode string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByWalletAddress(ctx context.Context, walletAddress string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByReferralCode(ctx context.Context, referralCode string) (bool, error)
	UpdateStatus(ctx context.Context, userID uuid.UUID, status Status) error

	// Preferences operations
	CreatePreferences(ctx context.Context, preferences *Preferences) error
	FindPreferencesByUserID(ctx context.Context, userID uuid.UUID) (*Preferences, error)
	UpdatePreferences(ctx context.Context, preferences *Preferences) error

	// Stats operations
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserStats, error)
}
