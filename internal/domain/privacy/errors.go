package privacy

import "errors"

var (
	// Privacy Settings Errors
	ErrPrivacySettingsNotFound = errors.New("privacy settings not found")
	ErrInvalidVisibility       = errors.New("invalid visibility setting")

	// Blocking Errors
	ErrUserAlreadyBlocked = errors.New("user already blocked")
	ErrUserNotBlocked     = errors.New("user not blocked")
	ErrCannotBlockSelf    = errors.New("cannot block yourself")

	// Disappearing Messages Errors
	ErrInvalidDuration           = errors.New("invalid disappearing message duration")
	ErrDisappearingNotEnabled    = errors.New("disappearing messages not enabled")
	ErrUnauthorizedToModify      = errors.New("unauthorized to modify disappearing messages settings")

	// Two-Factor Authentication Errors
	ErrTwoFactorNotEnabled       = errors.New("two-factor authentication not enabled")
	ErrTwoFactorAlreadyEnabled   = errors.New("two-factor authentication already enabled")
	ErrInvalidTwoFactorCode      = errors.New("invalid two-factor authentication code")
	ErrInvalidBackupCode         = errors.New("invalid backup code")
	ErrNoBackupCodesRemaining    = errors.New("no backup codes remaining")

	// Rate Limiting Errors
	ErrRateLimitExceeded         = errors.New("rate limit exceeded")
	ErrTooManyRequests           = errors.New("too many requests")
)
