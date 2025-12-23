package user

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID
	WalletAddress string
	Username      string
	Avatar        *string
	Bio           *string // User biography/about
	PublicKey     string  // Solana wallet public key for authentication
	ReferralCode  string  // Unique referral code for this user
	Status        Status
	LastSeen      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Status represents user online status
type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
	StatusAway    Status = "away"
)

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusOnline, StatusOffline, StatusAway:
		return true
	default:
		return false
	}
}

// NewUser creates a new user
func NewUser(walletAddress, username, publicKey string) *User {
	return &User{
		ID:            uuid.New(),
		WalletAddress: walletAddress,
		Username:      username,
		PublicKey:     publicKey,
		ReferralCode:  generateReferralCode(),
		Status:        StatusOffline,
		LastSeen:      time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// generateReferralCode generates a unique 8-character referral code
func generateReferralCode() string {
	// Generate 5 random bytes (40 bits)
	b := make([]byte, 5)
	rand.Read(b)

	// Encode to base32 and take first 8 characters
	code := base32.StdEncoding.EncodeToString(b)[:8]

	// Remove padding and make uppercase
	return strings.ToUpper(strings.TrimRight(code, "="))
}

// UpdateUsername updates the username
func (u *User) UpdateUsername(username string) {
	u.Username = username
	u.UpdatedAt = time.Now()
}

// SetAvatar sets the avatar URL
func (u *User) SetAvatar(avatar string) {
	u.Avatar = &avatar
	u.UpdatedAt = time.Now()
}

// UpdateStatus updates the user status
func (u *User) UpdateStatus(status Status) {
	if status.IsValid() {
		u.Status = status
		u.LastSeen = time.Now()
		u.UpdatedAt = time.Now()
	}
}

// IsOnline checks if user is online
func (u *User) IsOnline() bool {
	return u.Status == StatusOnline
}
