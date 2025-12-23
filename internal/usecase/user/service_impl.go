package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domainUser "github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/email"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

type service struct {
	userRepo    domainUser.Repository
	emailClient *email.Client
}

// NewService creates a new user service
func NewService(userRepo domainUser.Repository, emailClient *email.Client) Service {
	return &service{
		userRepo:    userRepo,
		emailClient: emailClient,
	}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*dto.GetProfileResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Get user stats
	stats, err := s.userRepo.GetUserStats(ctx, uid)
	if err != nil {
		// If stats fail, just log and continue with zero stats
		stats = &domainUser.UserStats{
			UserID:           uid,
			MessageCount:     0,
			GroupCount:       0,
			ChannelCount:     0,
			ContactCount:     0,
			TransactionCount: 0,
		}
	}

	return &dto.GetProfileResponse{
		ID:               user.ID.String(),
		WalletAddress:    user.WalletAddress,
		Username:         user.Username,
		Avatar:           user.Avatar,
		Bio:              user.Bio,
		Status:           string(user.Status),
		CreatedAt:        user.CreatedAt,
		MessageCount:     stats.MessageCount,
		GroupCount:       stats.GroupCount,
		ChannelCount:     stats.ChannelCount,
		ContactCount:     stats.ContactCount,
		TransactionCount: stats.TransactionCount,
	}, nil
}

func (s *service) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error) {
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Username != nil {
		user.UpdateUsername(*req.Username)
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dto.UpdateProfileResponse{
		ID:            user.ID.String(),
		WalletAddress: user.WalletAddress,
		Username:      user.Username,
		Avatar:        user.Avatar,
		Bio:           user.Bio,
		Status:        string(user.Status),
		CreatedAt:     user.CreatedAt,
	}, nil
}

func (s *service) GetPreferences(ctx context.Context, userID string) (*dto.GetPreferencesResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get preferences from database (will create defaults if not found)
	prefs, err := s.userRepo.FindPreferencesByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	return &dto.GetPreferencesResponse{
		Language:             prefs.Language,
		Theme:                prefs.Theme,
		NotificationsEnabled: prefs.NotificationsEnabled,
		SoundEnabled:         prefs.SoundEnabled,
		EmailNotifications:   prefs.EmailNotifications,
	}, nil
}

func (s *service) UpdatePreferences(ctx context.Context, req *dto.UpdatePreferencesRequest) (*dto.UpdatePreferencesResponse, error) {
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get existing preferences
	prefs, err := s.userRepo.FindPreferencesByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	// Update fields if provided
	if req.Language != nil {
		prefs.UpdateLanguage(*req.Language)
	}
	if req.Theme != nil {
		prefs.UpdateTheme(*req.Theme)
	}
	if req.NotificationsEnabled != nil {
		prefs.UpdateNotifications(*req.NotificationsEnabled)
	}
	if req.SoundEnabled != nil {
		prefs.UpdateSound(*req.SoundEnabled)
	}
	if req.EmailNotifications != nil {
		prefs.UpdateEmailNotifications(*req.EmailNotifications)
	}

	// Save updated preferences
	if err := s.userRepo.UpdatePreferences(ctx, prefs); err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	return &dto.UpdatePreferencesResponse{
		Language:             prefs.Language,
		Theme:                prefs.Theme,
		NotificationsEnabled: prefs.NotificationsEnabled,
		SoundEnabled:         prefs.SoundEnabled,
		EmailNotifications:   prefs.EmailNotifications,
	}, nil
}

func (s *service) GetUserByID(ctx context.Context, userID string) (*dto.GetUserResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &dto.GetUserResponse{
		ID:            user.ID.String(),
		WalletAddress: user.WalletAddress,
		Username:      user.Username,
		Avatar:        user.Avatar,
		Bio:           user.Bio,
		Status:        string(user.Status),
	}, nil
}

func (s *service) CheckWalletExists(ctx context.Context, walletAddress string) (bool, *dto.GetUserResponse, error) {
	user, err := s.userRepo.FindByWalletAddress(ctx, walletAddress)
	if err != nil {
		// User not found - not an error, just return false
		return false, nil, nil
	}

	return true, &dto.GetUserResponse{
		ID:            user.ID.String(),
		WalletAddress: user.WalletAddress,
		Username:      user.Username,
		Avatar:        user.Avatar,
		Bio:           user.Bio,
		Status:        string(user.Status),
	}, nil
}

func (s *service) SendInvitation(ctx context.Context, req *dto.SendInvitationRequest) error {
	// Parse sender ID
	senderID, err := uuid.Parse(req.SenderID)
	if err != nil {
		return fmt.Errorf("invalid sender ID: %w", err)
	}

	// Get sender information
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// Determine sender name
	senderName := "A SoTalk User"
	if sender.Username != "" {
		senderName = sender.Username
	}

	// Generate invite link
	inviteLink := fmt.Sprintf("https://sotalk.com/invite?ref=%s", sender.ReferralCode)
	if req.InviteLink != "" {
		inviteLink = req.InviteLink
	}

	logger.Info("Sending invitation email",
		zap.String("sender_id", req.SenderID),
		zap.String("sender_name", senderName),
		zap.String("recipient_email", req.Email),
		zap.String("invite_link", inviteLink),
	)

	// Send invitation email via SMTP
	if s.emailClient == nil {
		logger.Warn("Email client is nil, skipping email send")
		return fmt.Errorf("email service not configured")
	}

	if err := s.emailClient.SendInvitation(req.Email, senderName, inviteLink); err != nil {
		logger.Error("Failed to send invitation email",
			zap.String("recipient", req.Email),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	logger.Info("Invitation email sent successfully",
		zap.String("recipient", req.Email),
		zap.String("sender", senderName),
	)

	return nil
}
