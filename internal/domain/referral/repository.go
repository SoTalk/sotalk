package referral

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Repository defines the interface for referral data operations
type Repository interface {
	// Create creates a new referral
	Create(ctx context.Context, referral *Referral) error

	// FindByID finds a referral by ID
	FindByID(ctx context.Context, id uuid.UUID) (*Referral, error)

	// FindByRefereeID finds a referral by referee (referred user) ID
	FindByRefereeID(ctx context.Context, refereeID uuid.UUID) (*Referral, error)

	// FindByReferrerID finds all referrals made by a referrer
	FindByReferrerID(ctx context.Context, referrerID uuid.UUID, limit, offset int) ([]*Referral, error)

	// Update updates a referral
	Update(ctx context.Context, referral *Referral) error

	// GetStats retrieves referral statistics for a user
	GetStats(ctx context.Context, userID uuid.UUID) (*ReferralStats, error)

	// CountByReferrerID counts referrals by referrer ID
	CountByReferrerID(ctx context.Context, referrerID uuid.UUID) (int, error)

	// ExistsByRefereeID checks if a referee already has a referral
	ExistsByRefereeID(ctx context.Context, refereeID uuid.UUID) (bool, error)
}

// Errors
var (
	ErrReferralNotFound     = errors.New("referral not found")
	ErrReferralAlreadyExists = errors.New("user already has a referral")
	ErrInvalidReferralCode  = errors.New("invalid referral code")
	ErrSelfReferral         = errors.New("cannot refer yourself")
	ErrReferralExpired      = errors.New("referral has expired")
)
