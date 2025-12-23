package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/user"
	"gorm.io/gorm"
)

// userRepository implements user.Repository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) user.Repository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	dbUser := toUserModel(u)
	result := r.db.WithContext(ctx).Create(dbUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return user.ErrUserAlreadyExists
		}
		return result.Error
	}

	// Update domain entity with generated ID if needed
	u.ID = dbUser.ID
	u.CreatedAt = dbUser.CreatedAt
	u.UpdatedAt = dbUser.UpdatedAt

	return nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var dbUser User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}
		return nil, result.Error
	}

	return toDomainUser(&dbUser), nil
}

// FindByWalletAddress finds a user by wallet address
func (r *userRepository) FindByWalletAddress(ctx context.Context, walletAddress string) (*user.User, error) {
	var dbUser User
	result := r.db.WithContext(ctx).Where("wallet_address = ?", walletAddress).First(&dbUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}
		return nil, result.Error
	}

	return toDomainUser(&dbUser), nil
}

// FindByUsername finds a user by username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	var dbUser User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&dbUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}
		return nil, result.Error
	}

	return toDomainUser(&dbUser), nil
}


// Update updates a user
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	dbUser := toUserModel(u)
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", u.ID).Updates(dbUser)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// ExistsByWalletAddress checks if a user exists by wallet address
func (r *userRepository) ExistsByWalletAddress(ctx context.Context, walletAddress string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&User{}).Where("wallet_address = ?", walletAddress).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// ExistsByUsername checks if a user exists by username
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&User{}).Where("username = ?", username).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}


// UpdateStatus updates user status
func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status user.Status) error {
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"status":     string(status),
			"last_seen":  gorm.Expr("CURRENT_TIMESTAMP"),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// FindByReferralCode finds a user by referral code
func (r *userRepository) FindByReferralCode(ctx context.Context, referralCode string) (*user.User, error) {
	var dbUser User
	result := r.db.WithContext(ctx).Where("referral_code = ?", referralCode).First(&dbUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}
		return nil, result.Error
	}

	return toDomainUser(&dbUser), nil
}

// ExistsByReferralCode checks if a user exists by referral code
func (r *userRepository) ExistsByReferralCode(ctx context.Context, referralCode string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&User{}).Where("referral_code = ?", referralCode).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// Mapper functions

// toUserModel converts domain User to GORM User model
func toUserModel(u *user.User) *User {
	return &User{
		ID:            u.ID,
		WalletAddress: u.WalletAddress,
		Username:      u.Username,
		Avatar:        u.Avatar,
		Bio:           u.Bio,
		PublicKey:     u.PublicKey,
		ReferralCode:  u.ReferralCode,
		Status:        string(u.Status),
		LastSeen:      u.LastSeen,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// toDomainUser converts GORM User model to domain User
func toDomainUser(m *User) *user.User {
	return &user.User{
		ID:            m.ID,
		WalletAddress: m.WalletAddress,
		Username:      m.Username,
		Avatar:        m.Avatar,
		Bio:           m.Bio,
		PublicKey:     m.PublicKey,
		ReferralCode:  m.ReferralCode,
		Status:        user.Status(m.Status),
		LastSeen:      m.LastSeen,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// Preferences repository methods

// CreatePreferences creates new user preferences
func (r *userRepository) CreatePreferences(ctx context.Context, prefs *user.Preferences) error {
	dbPrefs := toPreferencesModel(prefs)
	result := r.db.WithContext(ctx).Create(dbPrefs)
	if result.Error != nil {
		return result.Error
	}

	// Update domain entity with generated fields
	prefs.ID = dbPrefs.ID
	prefs.CreatedAt = dbPrefs.CreatedAt
	prefs.UpdatedAt = dbPrefs.UpdatedAt

	return nil
}

// FindPreferencesByUserID finds preferences by user ID
func (r *userRepository) FindPreferencesByUserID(ctx context.Context, userID uuid.UUID) (*user.Preferences, error) {
	var dbPrefs UserPreferences
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&dbPrefs)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create default preferences if not found
			prefs := user.NewPreferences(userID)
			if err := r.CreatePreferences(ctx, prefs); err != nil {
				return nil, err
			}
			return prefs, nil
		}
		return nil, result.Error
	}

	return toDomainPreferences(&dbPrefs), nil
}

// UpdatePreferences updates user preferences
func (r *userRepository) UpdatePreferences(ctx context.Context, prefs *user.Preferences) error {
	dbPrefs := toPreferencesModel(prefs)
	result := r.db.WithContext(ctx).Model(&UserPreferences{}).Where("user_id = ?", prefs.UserID).Updates(dbPrefs)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Preferences mapper functions

// toPreferencesModel converts domain Preferences to GORM model
func toPreferencesModel(prefs *user.Preferences) *UserPreferences {
	return &UserPreferences{
		ID:                   prefs.ID,
		UserID:               prefs.UserID,
		Language:             prefs.Language,
		Theme:                prefs.Theme,
		NotificationsEnabled: prefs.NotificationsEnabled,
		SoundEnabled:         prefs.SoundEnabled,
		EmailNotifications:   prefs.EmailNotifications,
		CreatedAt:            prefs.CreatedAt,
		UpdatedAt:            prefs.UpdatedAt,
	}
}

// toDomainPreferences converts GORM model to domain Preferences
func toDomainPreferences(m *UserPreferences) *user.Preferences {
	return &user.Preferences{
		ID:                   m.ID,
		UserID:               m.UserID,
		Language:             m.Language,
		Theme:                m.Theme,
		NotificationsEnabled: m.NotificationsEnabled,
		SoundEnabled:         m.SoundEnabled,
		EmailNotifications:   m.EmailNotifications,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

// GetUserStats retrieves user activity statistics
func (r *userRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*user.UserStats, error) {
	stats := user.NewUserStats(userID)

	// Count messages sent by user
	var messageCount int64
	if err := r.db.WithContext(ctx).
		Table("messages").
		Where("sender_id = ?", userID).
		Count(&messageCount).Error; err != nil {
		return nil, err
	}
	stats.MessageCount = int(messageCount)

	// Count groups user is a member of
	var groupCount int64
	if err := r.db.WithContext(ctx).
		Table("group_members").
		Where("user_id = ?", userID).
		Count(&groupCount).Error; err != nil {
		return nil, err
	}
	stats.GroupCount = int(groupCount)

	// Count channels user is subscribed to
	var channelCount int64
	if err := r.db.WithContext(ctx).
		Table("channel_subscribers").
		Where("user_id = ?", userID).
		Count(&channelCount).Error; err != nil {
		return nil, err
	}
	stats.ChannelCount = int(channelCount)

	// Count distinct conversations (contacts)
	var contactCount int64
	if err := r.db.WithContext(ctx).
		Table("conversation_participants").
		Where("user_id = ?", userID).
		Count(&contactCount).Error; err != nil {
		return nil, err
	}
	stats.ContactCount = int(contactCount)

	// Count transactions
	var transactionCount int64
	if err := r.db.WithContext(ctx).
		Table("transactions").
		Where("user_id = ?", userID).
		Count(&transactionCount).Error; err != nil {
		return nil, err
	}
	stats.TransactionCount = int(transactionCount)

	return stats, nil
}
