package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/referral"
	"gorm.io/gorm"
)

// referralRepository implements referral.Repository interface
type referralRepository struct {
	db *gorm.DB
}

// NewReferralRepository creates a new referral repository
func NewReferralRepository(db *gorm.DB) referral.Repository {
	return &referralRepository{db: db}
}

// Create creates a new referral
func (r *referralRepository) Create(ctx context.Context, ref *referral.Referral) error {
	dbReferral := toReferralModel(ref)
	result := r.db.WithContext(ctx).Create(dbReferral)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return referral.ErrReferralAlreadyExists
		}
		return result.Error
	}

	// Update domain entity with generated values
	ref.ID = dbReferral.ID
	ref.CreatedAt = dbReferral.CreatedAt
	ref.UpdatedAt = dbReferral.UpdatedAt

	return nil
}

// FindByID finds a referral by ID
func (r *referralRepository) FindByID(ctx context.Context, id uuid.UUID) (*referral.Referral, error) {
	var dbReferral Referral
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbReferral)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, referral.ErrReferralNotFound
		}
		return nil, result.Error
	}

	return toDomainReferral(&dbReferral), nil
}

// FindByRefereeID finds a referral by referee (referred user) ID
func (r *referralRepository) FindByRefereeID(ctx context.Context, refereeID uuid.UUID) (*referral.Referral, error) {
	var dbReferral Referral
	result := r.db.WithContext(ctx).Where("referee_id = ?", refereeID).First(&dbReferral)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, referral.ErrReferralNotFound
		}
		return nil, result.Error
	}

	return toDomainReferral(&dbReferral), nil
}

// FindByReferrerID finds all referrals made by a referrer
func (r *referralRepository) FindByReferrerID(ctx context.Context, referrerID uuid.UUID, limit, offset int) ([]*referral.Referral, error) {
	var dbReferrals []Referral
	query := r.db.WithContext(ctx).
		Where("referrer_id = ?", referrerID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	result := query.Find(&dbReferrals)
	if result.Error != nil {
		return nil, result.Error
	}

	referrals := make([]*referral.Referral, len(dbReferrals))
	for i, dbRef := range dbReferrals {
		referrals[i] = toDomainReferral(&dbRef)
	}

	return referrals, nil
}

// Update updates a referral
func (r *referralRepository) Update(ctx context.Context, ref *referral.Referral) error {
	dbReferral := toReferralModel(ref)
	result := r.db.WithContext(ctx).Model(&Referral{}).
		Where("id = ?", ref.ID).
		Updates(dbReferral)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return referral.ErrReferralNotFound
	}

	return nil
}

// GetStats retrieves referral statistics for a user
func (r *referralRepository) GetStats(ctx context.Context, userID uuid.UUID) (*referral.ReferralStats, error) {
	var stats struct {
		TotalReferrals int
		PendingCount   int
		CompletedCount int
		RewardedCount  int
		TotalRewards   uint64
	}

	// Get counts by status
	err := r.db.WithContext(ctx).
		Model(&Referral{}).
		Select(`
			COUNT(*) as total_referrals,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
			COUNT(CASE WHEN status = 'rewarded' THEN 1 END) as rewarded_count,
			COALESCE(SUM(CASE WHEN reward_amount IS NOT NULL THEN reward_amount ELSE 0 END), 0) as total_rewards
		`).
		Where("referrer_id = ?", userID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return &referral.ReferralStats{
		UserID:         userID,
		TotalReferrals: stats.TotalReferrals,
		PendingCount:   stats.PendingCount,
		CompletedCount: stats.CompletedCount,
		RewardedCount:  stats.RewardedCount,
		TotalRewards:   stats.TotalRewards,
	}, nil
}

// CountByReferrerID counts referrals by referrer ID
func (r *referralRepository) CountByReferrerID(ctx context.Context, referrerID uuid.UUID) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&Referral{}).
		Where("referrer_id = ?", referrerID).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(count), nil
}

// ExistsByRefereeID checks if a referee already has a referral
func (r *referralRepository) ExistsByRefereeID(ctx context.Context, refereeID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&Referral{}).
		Where("referee_id = ?", refereeID).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// Mapper functions

// toReferralModel converts domain Referral to GORM Referral model
func toReferralModel(ref *referral.Referral) *Referral {
	model := &Referral{
		ID:         ref.ID,
		ReferrerID: ref.ReferrerID,
		RefereeID:  ref.RefereeID,
		Code:       ref.Code,
		Status:     string(ref.Status),
		CreatedAt:  ref.CreatedAt,
		UpdatedAt:  ref.UpdatedAt,
	}

	if ref.Reward != nil {
		rewardType := string(ref.Reward.Type)
		model.RewardType = &rewardType
		model.RewardAmount = &ref.Reward.Amount
		if ref.Reward.TxSig != "" {
			model.RewardTxSig = &ref.Reward.TxSig
		}
	}

	return model
}

// toDomainReferral converts GORM Referral model to domain Referral
func toDomainReferral(m *Referral) *referral.Referral {
	ref := &referral.Referral{
		ID:         m.ID,
		ReferrerID: m.ReferrerID,
		RefereeID:  m.RefereeID,
		Code:       m.Code,
		Status:     referral.Status(m.Status),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}

	if m.RewardType != nil && m.RewardAmount != nil {
		ref.Reward = &referral.Reward{
			Type:   referral.RewardType(*m.RewardType),
			Amount: *m.RewardAmount,
		}
		if m.RewardTxSig != nil {
			ref.Reward.TxSig = *m.RewardTxSig
		}
	}

	return ref
}
