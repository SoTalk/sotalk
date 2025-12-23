package dto

import "time"

// PrivacySettingsRequest represents a request to update privacy settings
type PrivacySettingsRequest struct {
	ProfilePhotoVisibility *string `json:"profile_photo_visibility,omitempty"` // "everyone", "contacts", "nobody"
	LastSeenVisibility     *string `json:"last_seen_visibility,omitempty"`
	StatusVisibility       *string `json:"status_visibility,omitempty"`
	ReadReceiptsEnabled    *bool   `json:"read_receipts_enabled,omitempty"`
	TypingIndicatorEnabled *bool   `json:"typing_indicator_enabled,omitempty"`
}

// PrivacySettingsResponse represents privacy settings
type PrivacySettingsResponse struct {
	UserID                 string    `json:"user_id"`
	ProfilePhotoVisibility string    `json:"profile_photo_visibility"`
	LastSeenVisibility     string    `json:"last_seen_visibility"`
	StatusVisibility       string    `json:"status_visibility"`
	ReadReceiptsEnabled    bool      `json:"read_receipts_enabled"`
	TypingIndicatorEnabled bool      `json:"typing_indicator_enabled"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// BlockUserRequest represents a request to block a user
type BlockUserRequest struct {
	BlockedUserID string  `json:"blocked_user_id" binding:"required"`
	Reason        *string `json:"reason,omitempty"`
}

// BlockedUserResponse represents a blocked user
type BlockedUserResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	BlockedUserID string    `json:"blocked_user_id"`
	Reason        string    `json:"reason,omitempty"`
	BlockedAt     time.Time `json:"blocked_at"`
}

// DisappearingMessagesRequest represents a request to configure disappearing messages
type DisappearingMessagesRequest struct {
	ConversationID  string `json:"conversation_id" binding:"required"`
	DurationSeconds int    `json:"duration_seconds" binding:"required,min=0"`
}

// DisappearingMessagesResponse represents disappearing messages configuration
type DisappearingMessagesResponse struct {
	ConversationID  string    `json:"conversation_id"`
	DurationSeconds int       `json:"duration_seconds"`
	EnabledBy       string    `json:"enabled_by"`
	EnabledAt       time.Time `json:"enabled_at"`
}

// TwoFactorSetupRequest represents a request to setup 2FA
type TwoFactorSetupRequest struct {
	Code string `json:"code" binding:"required"` // TOTP code to verify
}

// TwoFactorSetupResponse represents 2FA setup result
type TwoFactorSetupResponse struct {
	Secret      string   `json:"secret"`       // TOTP secret (only returned on setup)
	QRCodeURL   string   `json:"qr_code_url"`  // QR code for authenticator apps
	BackupCodes []string `json:"backup_codes"` // One-time backup codes
	Enabled     bool     `json:"enabled"`
}

// TwoFactorVerifyRequest represents a request to verify 2FA code
type TwoFactorVerifyRequest struct {
	Code string `json:"code" binding:"required"`
}

// TwoFactorStatusResponse represents 2FA status
type TwoFactorStatusResponse struct {
	Enabled       bool       `json:"enabled"`
	EnabledAt     *time.Time `json:"enabled_at,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	BackupCodesRemaining int `json:"backup_codes_remaining"`
}

// RateLimitInfoResponse represents rate limit information
type RateLimitInfoResponse struct {
	Key           string    `json:"key"`
	Count         int       `json:"count"`
	Limit         int       `json:"limit"`
	Remaining     int       `json:"remaining"`
	WindowSeconds int       `json:"window_seconds"`
	ResetAt       time.Time `json:"reset_at"`
}
