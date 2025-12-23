package referral

import (
	"time"

	"github.com/google/uuid"
)

// Referral represents a referral relationship between users
type Referral struct {
	ID         uuid.UUID
	ReferrerID uuid.UUID // User who referred
	RefereeID  uuid.UUID // User who was referred
	Code       string    // The referral code used
	Status     Status
	Reward     *Reward // Optional reward details
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Status represents the status of a referral
type Status string

const (
	StatusPending   Status = "pending"   // Referee registered but not verified
	StatusCompleted Status = "completed" // Referee completed verification
	StatusRewarded  Status = "rewarded"  // Referrer has been rewarded
	StatusExpired   Status = "expired"   // Referral expired
)

// Reward represents referral reward details
type Reward struct {
	Type   RewardType `json:"type"`   // Type of reward
	Amount uint64     `json:"amount"` // Amount in lamports (for SOL rewards)
	TxSig  string     `json:"tx_sig"` // Transaction signature if reward was paid
}

// RewardType represents the type of referral reward
type RewardType string

const (
	RewardTypeSOL    RewardType = "sol"     // SOL token reward
	RewardTypePoints RewardType = "points"  // Points reward
	RewardTypeNone   RewardType = "none"    // No reward
)

// ReferralStats contains aggregated referral statistics
type ReferralStats struct {
	UserID         uuid.UUID
	TotalReferrals int
	PendingCount   int
	CompletedCount int
	RewardedCount  int
	TotalRewards   uint64 // Total rewards earned in lamports
}

// NewReferral creates a new referral
func NewReferral(referrerID, refereeID uuid.UUID, code string) *Referral {
	return &Referral{
		ID:         uuid.New(),
		ReferrerID: referrerID,
		RefereeID:  refereeID,
		Code:       code,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Complete marks the referral as completed
func (r *Referral) Complete() {
	r.Status = StatusCompleted
	r.UpdatedAt = time.Now()
}

// MarkRewarded marks the referral as rewarded
func (r *Referral) MarkRewarded(reward Reward) {
	r.Status = StatusRewarded
	r.Reward = &reward
	r.UpdatedAt = time.Now()
}

// Expire marks the referral as expired
func (r *Referral) Expire() {
	r.Status = StatusExpired
	r.UpdatedAt = time.Now()
}

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusCompleted, StatusRewarded, StatusExpired:
		return true
	default:
		return false
	}
}
