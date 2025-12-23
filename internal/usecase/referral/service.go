package referral

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/referral"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// Service defines the interface for referral use cases
type Service interface {
	// GetMyReferralCode gets the user's referral code
	GetMyReferralCode(ctx context.Context, userID string) (*dto.ReferralCodeResponse, error)

	// GetMyReferralStats gets the user's referral statistics
	GetMyReferralStats(ctx context.Context, userID string) (*dto.ReferralStatsResponse, error)

	// GetMyReferralHistory gets the user's referral history
	GetMyReferralHistory(ctx context.Context, userID string, limit, offset int) (*dto.ReferralHistoryResponse, error)

	// ValidateReferralCode validates a referral code
	ValidateReferralCode(ctx context.Context, code string) (*dto.ValidateReferralCodeResponse, error)

	// ApplyReferralCode applies a referral code to a new user (called during registration)
	ApplyReferralCode(ctx context.Context, refereeID uuid.UUID, code string) error

	// CompleteReferral marks a referral as completed (when referee completes verification/action)
	CompleteReferral(ctx context.Context, refereeID uuid.UUID) error
}

type service struct {
	referralRepo referral.Repository
	userRepo     user.Repository
}

// NewService creates a new referral service
func NewService(referralRepo referral.Repository, userRepo user.Repository) Service {
	return &service{
		referralRepo: referralRepo,
		userRepo:     userRepo,
	}
}

// GetMyReferralCode gets the user's referral code
func (s *service) GetMyReferralCode(ctx context.Context, userID string) (*dto.ReferralCodeResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Get referral count
	count, err := s.referralRepo.CountByReferrerID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to count referrals: %w", err)
	}

	return &dto.ReferralCodeResponse{
		Code:          user.ReferralCode,
		ReferralCount: count,
		ShareURL:      fmt.Sprintf("https://sotalk.app/invite/%s", user.ReferralCode),
	}, nil
}

// GetMyReferralStats gets the user's referral statistics
func (s *service) GetMyReferralStats(ctx context.Context, userID string) (*dto.ReferralStatsResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	stats, err := s.referralRepo.GetStats(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &dto.ReferralStatsResponse{
		TotalReferrals: stats.TotalReferrals,
		PendingCount:   stats.PendingCount,
		CompletedCount: stats.CompletedCount,
		RewardedCount:  stats.RewardedCount,
		TotalRewards:   stats.TotalRewards,
	}, nil
}

// GetMyReferralHistory gets the user's referral history
func (s *service) GetMyReferralHistory(ctx context.Context, userID string, limit, offset int) (*dto.ReferralHistoryResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	referrals, err := s.referralRepo.FindByReferrerID(ctx, uid, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get referrals: %w", err)
	}

	items := make([]dto.ReferralItem, len(referrals))
	for i, ref := range referrals {
		// Get referee user details
		referee, err := s.userRepo.FindByID(ctx, ref.RefereeID)
		if err != nil {
			continue // Skip if user not found
		}

		items[i] = dto.ReferralItem{
			ID:           ref.ID.String(),
			RefereeID:    ref.RefereeID.String(),
			Username:     referee.Username,
			Status:       string(ref.Status),
			RewardAmount: 0,
			CreatedAt:    ref.CreatedAt,
		}

		if ref.Reward != nil {
			items[i].RewardAmount = ref.Reward.Amount
		}
	}

	return &dto.ReferralHistoryResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// ValidateReferralCode validates a referral code
func (s *service) ValidateReferralCode(ctx context.Context, code string) (*dto.ValidateReferralCodeResponse, error) {
	if code == "" {
		return &dto.ValidateReferralCodeResponse{
			Valid: false,
			Error: "Referral code is required",
		}, nil
	}

	// Check if code exists
	referrer, err := s.userRepo.FindByReferralCode(ctx, code)
	if err != nil {
		if err == user.ErrUserNotFound {
			return &dto.ValidateReferralCodeResponse{
				Valid: false,
				Error: "Invalid referral code",
			}, nil
		}
		return nil, fmt.Errorf("failed to validate code: %w", err)
	}

	return &dto.ValidateReferralCodeResponse{
		Valid:        true,
		ReferrerID:   referrer.ID.String(),
		ReferrerName: referrer.Username,
	}, nil
}

// ApplyReferralCode applies a referral code to a new user
func (s *service) ApplyReferralCode(ctx context.Context, refereeID uuid.UUID, code string) error {
	if code == "" {
		return nil // No referral code provided, skip
	}

	// Validate code and get referrer
	referrer, err := s.userRepo.FindByReferralCode(ctx, code)
	if err != nil {
		if err == user.ErrUserNotFound {
			return referral.ErrInvalidReferralCode
		}
		return fmt.Errorf("failed to find referrer: %w", err)
	}

	// Prevent self-referral
	if referrer.ID == refereeID {
		return referral.ErrSelfReferral
	}

	// Check if user already has a referral
	exists, err := s.referralRepo.ExistsByRefereeID(ctx, refereeID)
	if err != nil {
		return fmt.Errorf("failed to check existing referral: %w", err)
	}
	if exists {
		return referral.ErrReferralAlreadyExists
	}

	// Create referral
	ref := referral.NewReferral(referrer.ID, refereeID, code)
	if err := s.referralRepo.Create(ctx, ref); err != nil {
		return fmt.Errorf("failed to create referral: %w", err)
	}

	return nil
}

// CompleteReferral marks a referral as completed
func (s *service) CompleteReferral(ctx context.Context, refereeID uuid.UUID) error {
	// Find referral by referee ID
	ref, err := s.referralRepo.FindByRefereeID(ctx, refereeID)
	if err != nil {
		if err == referral.ErrReferralNotFound {
			return nil // No referral to complete, skip
		}
		return fmt.Errorf("failed to find referral: %w", err)
	}

	// Only complete if still pending
	if ref.Status != referral.StatusPending {
		return nil // Already completed or rewarded
	}

	// Mark as completed
	ref.Complete()

	// Update in database
	if err := s.referralRepo.Update(ctx, ref); err != nil {
		return fmt.Errorf("failed to update referral: %w", err)
	}

	// Trigger reward distribution
	if err := s.distributeReferralRewards(ctx, ref); err != nil {
		// Log error but don't fail the completion
		// Rewards can be retried via background job
		logger.Error("Failed to distribute referral rewards",
			zap.String("referral_id", ref.ID.String()),
			zap.String("referrer_id", ref.ReferrerID.String()),
			zap.String("referee_id", ref.RefereeID.String()),
			zap.Error(err),
		)
	}

	return nil
}

// distributeReferralRewards handles reward distribution for completed referrals
func (s *service) distributeReferralRewards(ctx context.Context, ref *referral.Referral) error {
	// Get referrer information
	referrer, err := s.userRepo.FindByID(ctx, ref.ReferrerID)
	if err != nil {
		return fmt.Errorf("failed to find referrer: %w", err)
	}

	// Get referee information
	referee, err := s.userRepo.FindByID(ctx, ref.RefereeID)
	if err != nil {
		return fmt.Errorf("failed to find referee: %w", err)
	}

	// Reward configuration (in a real system, this would come from config/database)
	const (
		referrerRewardLamports uint64 = 100_000_000 // 0.1 SOL for referrer
		refereeRewardLamports  uint64 = 50_000_000  // 0.05 SOL for referee
		rewardType                    = "referral_bonus"
	)

	logger.Info("Distributing referral rewards",
		zap.String("referral_id", ref.ID.String()),
		zap.String("referrer", referrer.Username),
		zap.String("referee", referee.Username),
		zap.Uint64("referrer_reward", referrerRewardLamports),
		zap.Uint64("referee_reward", refereeRewardLamports),
	)

	// TODO: In production, integrate with wallet service to actually send rewards
	// Example:
	// 1. Get referrer's default wallet
	// 2. Get referee's default wallet
	// 3. Send rewards from platform treasury wallet
	// 4. Record transaction signatures in referral record
	//
	// For now, we'll mark the referral as ready for reward distribution
	// A background job can process these and actually send the SOL

	// Update referral with reward information
	reward := referral.Reward{
		Type:   referral.RewardTypeSOL,
		Amount: referrerRewardLamports,
		TxSig:  "", // Will be set when actual blockchain transaction is made
	}
	ref.MarkRewarded(reward)

	if err := s.referralRepo.Update(ctx, ref); err != nil {
		return fmt.Errorf("failed to update referral with reward info: %w", err)
	}

	logger.Info("Referral rewards prepared for distribution",
		zap.String("referral_id", ref.ID.String()),
		zap.String("reward_type", rewardType),
		zap.Uint64("amount", referrerRewardLamports),
	)

	// In a production system, you would:
	// 1. Enqueue reward distribution job to a message queue (Redis, RabbitMQ, etc.)
	// 2. Have a background worker process the rewards
	// 3. Handle retries and failures gracefully
	// 4. Track transaction signatures for auditing

	return nil
}
