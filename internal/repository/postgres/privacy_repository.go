package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	domainPrivacy "github.com/yourusername/sotalk/internal/domain/privacy"
	"gorm.io/gorm"
)

type privacyRepository struct {
	db *gorm.DB
}

// NewPrivacyRepository creates a new privacy repository
func NewPrivacyRepository(db *gorm.DB) domainPrivacy.Repository {
	return &privacyRepository{db: db}
}

// Privacy Settings

func (r *privacyRepository) CreatePrivacySettings(ctx context.Context, settings *domainPrivacy.PrivacySettings) error {
	model := &PrivacySettings{
		UserID:                 settings.UserID,
		ProfilePhotoVisibility: string(settings.ProfilePhotoVisibility),
		LastSeenVisibility:     string(settings.LastSeenVisibility),
		StatusVisibility:       string(settings.StatusVisibility),
		ReadReceiptsEnabled:    settings.ReadReceiptsEnabled,
		TypingIndicatorEnabled: settings.TypingIndicatorEnabled,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	settings.CreatedAt = model.CreatedAt
	settings.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *privacyRepository) GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*domainPrivacy.PrivacySettings, error) {
	var model PrivacySettings
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainPrivacy.ErrPrivacySettingsNotFound
		}
		return nil, err
	}

	return &domainPrivacy.PrivacySettings{
		UserID:                 model.UserID,
		ProfilePhotoVisibility: domainPrivacy.Visibility(model.ProfilePhotoVisibility),
		LastSeenVisibility:     domainPrivacy.Visibility(model.LastSeenVisibility),
		StatusVisibility:       domainPrivacy.Visibility(model.StatusVisibility),
		ReadReceiptsEnabled:    model.ReadReceiptsEnabled,
		TypingIndicatorEnabled: model.TypingIndicatorEnabled,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}, nil
}

func (r *privacyRepository) UpdatePrivacySettings(ctx context.Context, settings *domainPrivacy.PrivacySettings) error {
	updates := map[string]interface{}{
		"profile_photo_visibility": string(settings.ProfilePhotoVisibility),
		"last_seen_visibility":     string(settings.LastSeenVisibility),
		"status_visibility":        string(settings.StatusVisibility),
		"read_receipts_enabled":    settings.ReadReceiptsEnabled,
		"typing_indicator_enabled": settings.TypingIndicatorEnabled,
	}

	result := r.db.WithContext(ctx).
		Model(&PrivacySettings{}).
		Where("user_id = ?", settings.UserID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainPrivacy.ErrPrivacySettingsNotFound
	}

	return nil
}

// Blocked Users

func (r *privacyRepository) BlockUser(ctx context.Context, blocked *domainPrivacy.BlockedUser) error {
	// Check if already blocked
	var existing BlockedUser
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND blocked_user_id = ?", blocked.UserID, blocked.BlockedUserID).
		First(&existing).Error

	if err == nil {
		return domainPrivacy.ErrUserAlreadyBlocked
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := &BlockedUser{
		UserID:        blocked.UserID,
		BlockedUserID: blocked.BlockedUserID,
		Reason:        blocked.Reason,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	blocked.ID = model.ID
	blocked.BlockedAt = model.BlockedAt
	return nil
}

func (r *privacyRepository) UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND blocked_user_id = ?", userID, blockedUserID).
		Delete(&BlockedUser{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainPrivacy.ErrUserNotBlocked
	}

	return nil
}

func (r *privacyRepository) IsUserBlocked(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&BlockedUser{}).
		Where("user_id = ? AND blocked_user_id = ?", userID, targetUserID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *privacyRepository) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]*domainPrivacy.BlockedUser, error) {
	var models []BlockedUser
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("blocked_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainPrivacy.BlockedUser, len(models))
	for i, m := range models {
		result[i] = &domainPrivacy.BlockedUser{
			ID:            m.ID,
			UserID:        m.UserID,
			BlockedUserID: m.BlockedUserID,
			Reason:        m.Reason,
			BlockedAt:     m.BlockedAt,
		}
	}

	return result, nil
}

func (r *privacyRepository) GetBlockedByUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&BlockedUser{}).
		Where("blocked_user_id = ?", userID).
		Pluck("user_id", &userIDs).Error

	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

// Disappearing Messages

func (r *privacyRepository) SetDisappearingMessages(ctx context.Context, config *domainPrivacy.DisappearingMessagesConfig) error {
	model := &DisappearingMessagesConfig{
		ConversationID:  config.ConversationID,
		DurationSeconds: config.DurationSeconds,
		EnabledBy:       config.EnabledBy,
	}

	// Upsert: Update if exists, create if not
	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", config.ConversationID).
		Assign(map[string]interface{}{
			"duration_seconds": config.DurationSeconds,
			"enabled_by":       config.EnabledBy,
		}).
		FirstOrCreate(model)

	if result.Error != nil {
		return result.Error
	}

	config.ID = model.ID
	config.EnabledAt = model.EnabledAt
	config.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *privacyRepository) GetDisappearingMessagesConfig(ctx context.Context, conversationID uuid.UUID) (*domainPrivacy.DisappearingMessagesConfig, error) {
	var model DisappearingMessagesConfig
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainPrivacy.ErrDisappearingNotEnabled
		}
		return nil, err
	}

	return &domainPrivacy.DisappearingMessagesConfig{
		ID:              model.ID,
		ConversationID:  model.ConversationID,
		DurationSeconds: model.DurationSeconds,
		EnabledBy:       model.EnabledBy,
		EnabledAt:       model.EnabledAt,
		UpdatedAt:       model.UpdatedAt,
	}, nil
}

func (r *privacyRepository) DisableDisappearingMessages(ctx context.Context, conversationID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Delete(&DisappearingMessagesConfig{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainPrivacy.ErrDisappearingNotEnabled
	}

	return nil
}

// Two-Factor Authentication

func (r *privacyRepository) CreateTwoFactorAuth(ctx context.Context, twoFA *domainPrivacy.TwoFactorAuth) error {
	// Convert backup codes to JSON string for storage
	backupCodesJSON := ""
	if len(twoFA.BackupCodes) > 0 {
		// In production, this should be encrypted
		// For now, we'll store as comma-separated
		for i, code := range twoFA.BackupCodes {
			if i > 0 {
				backupCodesJSON += ","
			}
			backupCodesJSON += code
		}
	}

	model := &TwoFactorAuth{
		UserID:      twoFA.UserID,
		Enabled:     twoFA.Enabled,
		Secret:      twoFA.Secret,
		BackupCodes: backupCodesJSON,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	twoFA.CreatedAt = model.CreatedAt
	twoFA.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *privacyRepository) GetTwoFactorAuth(ctx context.Context, userID uuid.UUID) (*domainPrivacy.TwoFactorAuth, error) {
	var model TwoFactorAuth
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainPrivacy.ErrTwoFactorNotEnabled
		}
		return nil, err
	}

	// Parse backup codes
	backupCodes := []string{}
	if model.BackupCodes != "" {
		// Simple split for now
		for _, code := range splitString(model.BackupCodes, ",") {
			if code != "" {
				backupCodes = append(backupCodes, code)
			}
		}
	}

	return &domainPrivacy.TwoFactorAuth{
		UserID:      model.UserID,
		Enabled:     model.Enabled,
		Secret:      model.Secret,
		BackupCodes: backupCodes,
		EnabledAt:   model.EnabledAt,
		LastUsedAt:  model.LastUsedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (r *privacyRepository) UpdateTwoFactorAuth(ctx context.Context, twoFA *domainPrivacy.TwoFactorAuth) error {
	// Convert backup codes to JSON string
	backupCodesJSON := ""
	if len(twoFA.BackupCodes) > 0 {
		for i, code := range twoFA.BackupCodes {
			if i > 0 {
				backupCodesJSON += ","
			}
			backupCodesJSON += code
		}
	}

	updates := map[string]interface{}{
		"enabled":      twoFA.Enabled,
		"secret":       twoFA.Secret,
		"backup_codes": backupCodesJSON,
	}

	if twoFA.Enabled && !twoFA.EnabledAt.IsZero() {
		updates["enabled_at"] = twoFA.EnabledAt
	}

	if !twoFA.LastUsedAt.IsZero() {
		updates["last_used_at"] = twoFA.LastUsedAt
	}

	result := r.db.WithContext(ctx).
		Model(&TwoFactorAuth{}).
		Where("user_id = ?", twoFA.UserID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainPrivacy.ErrTwoFactorNotEnabled
	}

	return nil
}

func (r *privacyRepository) DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&TwoFactorAuth{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainPrivacy.ErrTwoFactorNotEnabled
	}

	return nil
}

// Helper function
func splitString(s string, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
