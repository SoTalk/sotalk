package dto

import "time"

// ReferralCodeResponse is the response for getting user's referral code
type ReferralCodeResponse struct {
	Code          string `json:"code"`
	ReferralCount int    `json:"referral_count"`
	ShareURL      string `json:"share_url"`
}

// ReferralStatsResponse is the response for getting referral statistics
type ReferralStatsResponse struct {
	TotalReferrals int    `json:"total_referrals"`
	PendingCount   int    `json:"pending_count"`
	CompletedCount int    `json:"completed_count"`
	RewardedCount  int    `json:"rewarded_count"`
	TotalRewards   uint64 `json:"total_rewards"` // Total rewards earned in lamports
}

// ReferralHistoryResponse is the response for getting referral history
type ReferralHistoryResponse struct {
	Items []ReferralItem `json:"items"`
	Total int            `json:"total"`
}

// ReferralItem represents a single referral in the history
type ReferralItem struct {
	ID           string    `json:"id"`
	RefereeID    string    `json:"referee_id"`
	Username     string    `json:"username"`
	Status       string    `json:"status"`
	RewardAmount uint64    `json:"reward_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

// ValidateReferralCodeResponse is the response for validating a referral code
type ValidateReferralCodeResponse struct {
	Valid        bool   `json:"valid"`
	ReferrerID   string `json:"referrer_id,omitempty"`
	ReferrerName string `json:"referrer_name,omitempty"`
	Error        string `json:"error,omitempty"`
}

// ApplyReferralCodeRequest is the request for applying a referral code
type ApplyReferralCodeRequest struct {
	Code string `json:"code" validate:"required,min=8,max=8"`
}
