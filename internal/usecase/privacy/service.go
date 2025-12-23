package privacy

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/yourusername/sotalk/internal/domain/conversation"
	domainPrivacy "github.com/yourusername/sotalk/internal/domain/privacy"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

type service struct {
	privacyRepo      domainPrivacy.Repository
	conversationRepo conversation.Repository
}

// NewService creates a new privacy service
func NewService(privacyRepo domainPrivacy.Repository, conversationRepo conversation.Repository) Service {
	return &service{
		privacyRepo:      privacyRepo,
		conversationRepo: conversationRepo,
	}
}

// Privacy Settings

func (s *service) GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*dto.PrivacySettingsResponse, error) {
	settings, err := s.privacyRepo.GetPrivacySettings(ctx, userID)
	if err != nil {
		if err == domainPrivacy.ErrPrivacySettingsNotFound {
			// Create default settings
			settings = &domainPrivacy.PrivacySettings{
				UserID:                 userID,
				ProfilePhotoVisibility: domainPrivacy.VisibilityEveryone,
				LastSeenVisibility:     domainPrivacy.VisibilityEveryone,
				StatusVisibility:       domainPrivacy.VisibilityEveryone,
				ReadReceiptsEnabled:    true,
				TypingIndicatorEnabled: true,
			}

			if err := s.privacyRepo.CreatePrivacySettings(ctx, settings); err != nil {
				return nil, fmt.Errorf("failed to create default privacy settings: %w", err)
			}
		} else {
			return nil, err
		}
	}

	return &dto.PrivacySettingsResponse{
		UserID:                 settings.UserID.String(),
		ProfilePhotoVisibility: string(settings.ProfilePhotoVisibility),
		LastSeenVisibility:     string(settings.LastSeenVisibility),
		StatusVisibility:       string(settings.StatusVisibility),
		ReadReceiptsEnabled:    settings.ReadReceiptsEnabled,
		TypingIndicatorEnabled: settings.TypingIndicatorEnabled,
		UpdatedAt:              settings.UpdatedAt,
	}, nil
}

func (s *service) UpdatePrivacySettings(ctx context.Context, userID uuid.UUID, req *dto.PrivacySettingsRequest) (*dto.PrivacySettingsResponse, error) {
	// Get current settings
	settings, err := s.privacyRepo.GetPrivacySettings(ctx, userID)
	if err != nil {
		if err == domainPrivacy.ErrPrivacySettingsNotFound {
			// Create default settings first
			settings = &domainPrivacy.PrivacySettings{
				UserID:                 userID,
				ProfilePhotoVisibility: domainPrivacy.VisibilityEveryone,
				LastSeenVisibility:     domainPrivacy.VisibilityEveryone,
				StatusVisibility:       domainPrivacy.VisibilityEveryone,
				ReadReceiptsEnabled:    true,
				TypingIndicatorEnabled: true,
			}
		} else {
			return nil, err
		}
	}

	// Update only provided fields
	if req.ProfilePhotoVisibility != nil {
		settings.ProfilePhotoVisibility = domainPrivacy.Visibility(*req.ProfilePhotoVisibility)
	}
	if req.LastSeenVisibility != nil {
		settings.LastSeenVisibility = domainPrivacy.Visibility(*req.LastSeenVisibility)
	}
	if req.StatusVisibility != nil {
		settings.StatusVisibility = domainPrivacy.Visibility(*req.StatusVisibility)
	}
	if req.ReadReceiptsEnabled != nil {
		settings.ReadReceiptsEnabled = *req.ReadReceiptsEnabled
	}
	if req.TypingIndicatorEnabled != nil {
		settings.TypingIndicatorEnabled = *req.TypingIndicatorEnabled
	}

	if err := s.privacyRepo.UpdatePrivacySettings(ctx, settings); err != nil {
		return nil, err
	}

	logger.Info("Privacy settings updated",
		zap.String("user_id", userID.String()),
	)

	return s.GetPrivacySettings(ctx, userID)
}

// User Blocking

func (s *service) BlockUser(ctx context.Context, userID uuid.UUID, req *dto.BlockUserRequest) error {
	blockedUserID, err := uuid.Parse(req.BlockedUserID)
	if err != nil {
		return fmt.Errorf("invalid blocked user ID: %w", err)
	}

	// Prevent self-blocking
	if userID == blockedUserID {
		return domainPrivacy.ErrCannotBlockSelf
	}

	reason := ""
	if req.Reason != nil {
		reason = *req.Reason
	}

	blocked := &domainPrivacy.BlockedUser{
		UserID:        userID,
		BlockedUserID: blockedUserID,
		Reason:        reason,
	}

	if err := s.privacyRepo.BlockUser(ctx, blocked); err != nil {
		return err
	}

	logger.Info("User blocked",
		zap.String("user_id", userID.String()),
		zap.String("blocked_user_id", blockedUserID.String()),
	)

	return nil
}

func (s *service) UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error {
	if err := s.privacyRepo.UnblockUser(ctx, userID, blockedUserID); err != nil {
		return err
	}

	logger.Info("User unblocked",
		zap.String("user_id", userID.String()),
		zap.String("blocked_user_id", blockedUserID.String()),
	)

	return nil
}

func (s *service) IsUserBlocked(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error) {
	// Check both directions: is userID blocking targetUserID, or is targetUserID blocking userID
	blocked1, err := s.privacyRepo.IsUserBlocked(ctx, userID, targetUserID)
	if err != nil {
		return false, err
	}

	blocked2, err := s.privacyRepo.IsUserBlocked(ctx, targetUserID, userID)
	if err != nil {
		return false, err
	}

	return blocked1 || blocked2, nil
}

func (s *service) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]*dto.BlockedUserResponse, error) {
	blocked, err := s.privacyRepo.GetBlockedUsers(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.BlockedUserResponse, len(blocked))
	for i, b := range blocked {
		result[i] = &dto.BlockedUserResponse{
			ID:            b.ID.String(),
			UserID:        b.UserID.String(),
			BlockedUserID: b.BlockedUserID.String(),
			Reason:        b.Reason,
			BlockedAt:     b.BlockedAt,
		}
	}

	return result, nil
}

// Disappearing Messages

func (s *service) SetDisappearingMessages(ctx context.Context, userID uuid.UUID, req *dto.DisappearingMessagesRequest) (*dto.DisappearingMessagesResponse, error) {
	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation ID: %w", err)
	}

	// Validate duration
	if req.DurationSeconds < 0 {
		return nil, domainPrivacy.ErrInvalidDuration
	}

	// Verify user is participant in conversation
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant status: %w", err)
	}
	if !isParticipant {
		return nil, conversation.ErrNotParticipant
	}

	config := &domainPrivacy.DisappearingMessagesConfig{
		ConversationID:  conversationID,
		DurationSeconds: req.DurationSeconds,
		EnabledBy:       userID,
	}

	if err := s.privacyRepo.SetDisappearingMessages(ctx, config); err != nil {
		return nil, err
	}

	logger.Info("Disappearing messages configured",
		zap.String("conversation_id", conversationID.String()),
		zap.Int("duration_seconds", req.DurationSeconds),
	)

	return &dto.DisappearingMessagesResponse{
		ConversationID:  config.ConversationID.String(),
		DurationSeconds: config.DurationSeconds,
		EnabledBy:       config.EnabledBy.String(),
		EnabledAt:       config.EnabledAt,
	}, nil
}

func (s *service) GetDisappearingMessagesConfig(ctx context.Context, conversationID uuid.UUID) (*dto.DisappearingMessagesResponse, error) {
	config, err := s.privacyRepo.GetDisappearingMessagesConfig(ctx, conversationID)
	if err != nil {
		if err == domainPrivacy.ErrDisappearingNotEnabled {
			return &dto.DisappearingMessagesResponse{
				ConversationID:  conversationID.String(),
				DurationSeconds: 0,
			}, nil
		}
		return nil, err
	}

	return &dto.DisappearingMessagesResponse{
		ConversationID:  config.ConversationID.String(),
		DurationSeconds: config.DurationSeconds,
		EnabledBy:       config.EnabledBy.String(),
		EnabledAt:       config.EnabledAt,
	}, nil
}

func (s *service) DisableDisappearingMessages(ctx context.Context, userID, conversationID uuid.UUID) error {
	// Verify user is participant in conversation
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to check participant status: %w", err)
	}
	if !isParticipant {
		return conversation.ErrNotParticipant
	}

	if err := s.privacyRepo.DisableDisappearingMessages(ctx, conversationID); err != nil {
		return err
	}

	logger.Info("Disappearing messages disabled",
		zap.String("conversation_id", conversationID.String()),
	)

	return nil
}

// Two-Factor Authentication

func (s *service) SetupTwoFactorAuth(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorSetupResponse, error) {
	// Check if already enabled
	existing, err := s.privacyRepo.GetTwoFactorAuth(ctx, userID)
	if err == nil && existing.Enabled {
		return nil, domainPrivacy.ErrTwoFactorAlreadyEnabled
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Soldef",
		AccountName: userID.String(),
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	secret := key.Secret()

	// Generate backup codes
	backupCodes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		backupCodes[i] = code
	}

	// Store (but not enabled yet)
	twoFA := &domainPrivacy.TwoFactorAuth{
		UserID:      userID,
		Enabled:     false,
		Secret:      secret,
		BackupCodes: backupCodes,
	}

	if existing == nil {
		if err := s.privacyRepo.CreateTwoFactorAuth(ctx, twoFA); err != nil {
			return nil, err
		}
	} else {
		if err := s.privacyRepo.UpdateTwoFactorAuth(ctx, twoFA); err != nil {
			return nil, err
		}
	}

	logger.Info("2FA setup initiated",
		zap.String("user_id", userID.String()),
	)

	return &dto.TwoFactorSetupResponse{
		Secret:      secret,
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
		Enabled:     false,
	}, nil
}

func (s *service) VerifyAndEnableTwoFactor(ctx context.Context, userID uuid.UUID, req *dto.TwoFactorVerifyRequest) error {
	twoFA, err := s.privacyRepo.GetTwoFactorAuth(ctx, userID)
	if err != nil {
		return err
	}

	// Verify TOTP code
	valid := totp.Validate(req.Code, twoFA.Secret)
	if !valid {
		// Try backup codes
		validBackup := false
		remainingCodes := []string{}
		for _, code := range twoFA.BackupCodes {
			if code == req.Code {
				validBackup = true
				// Don't add used code back
			} else {
				remainingCodes = append(remainingCodes, code)
			}
		}

		if !validBackup {
			return domainPrivacy.ErrInvalidTwoFactorCode
		}

		twoFA.BackupCodes = remainingCodes
	}

	// Enable 2FA
	twoFA.Enabled = true
	twoFA.EnabledAt = time.Now()

	if err := s.privacyRepo.UpdateTwoFactorAuth(ctx, twoFA); err != nil {
		return err
	}

	logger.Info("2FA enabled",
		zap.String("user_id", userID.String()),
	)

	return nil
}

func (s *service) VerifyTwoFactorCode(ctx context.Context, userID uuid.UUID, req *dto.TwoFactorVerifyRequest) (bool, error) {
	twoFA, err := s.privacyRepo.GetTwoFactorAuth(ctx, userID)
	if err != nil {
		return false, err
	}

	if !twoFA.Enabled {
		return false, domainPrivacy.ErrTwoFactorNotEnabled
	}

	// Verify TOTP code
	valid := totp.Validate(req.Code, twoFA.Secret)
	if valid {
		// Update last used
		twoFA.LastUsedAt = time.Now()
		s.privacyRepo.UpdateTwoFactorAuth(ctx, twoFA)
		return true, nil
	}

	// Try backup codes
	for i, code := range twoFA.BackupCodes {
		if code == req.Code {
			// Remove used backup code
			twoFA.BackupCodes = append(twoFA.BackupCodes[:i], twoFA.BackupCodes[i+1:]...)
			twoFA.LastUsedAt = time.Now()
			s.privacyRepo.UpdateTwoFactorAuth(ctx, twoFA)
			return true, nil
		}
	}

	return false, nil
}

func (s *service) DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID, req *dto.TwoFactorVerifyRequest) error {
	// Verify code before disabling
	valid, err := s.VerifyTwoFactorCode(ctx, userID, req)
	if err != nil {
		return err
	}

	if !valid {
		return domainPrivacy.ErrInvalidTwoFactorCode
	}

	if err := s.privacyRepo.DisableTwoFactorAuth(ctx, userID); err != nil {
		return err
	}

	logger.Info("2FA disabled",
		zap.String("user_id", userID.String()),
	)

	return nil
}

func (s *service) GetTwoFactorStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorStatusResponse, error) {
	twoFA, err := s.privacyRepo.GetTwoFactorAuth(ctx, userID)
	if err != nil {
		if err == domainPrivacy.ErrTwoFactorNotEnabled {
			return &dto.TwoFactorStatusResponse{
				Enabled:              false,
				BackupCodesRemaining: 0,
			}, nil
		}
		return nil, err
	}

	response := &dto.TwoFactorStatusResponse{
		Enabled:              twoFA.Enabled,
		BackupCodesRemaining: len(twoFA.BackupCodes),
	}

	if !twoFA.EnabledAt.IsZero() {
		response.EnabledAt = &twoFA.EnabledAt
	}

	if !twoFA.LastUsedAt.IsZero() {
		response.LastUsedAt = &twoFA.LastUsedAt
	}

	return response, nil
}

// Helper functions

func generateBackupCode() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}
