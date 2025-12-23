package user

import "github.com/google/uuid"

// UserStats represents user activity statistics
type UserStats struct {
	UserID          uuid.UUID
	MessageCount    int // Total messages sent
	GroupCount      int // Number of groups user is in
	ChannelCount    int // Number of channels subscribed to
	ContactCount    int // Number of contacts/conversations
	TransactionCount int // Number of transactions
}

// NewUserStats creates a new UserStats instance
func NewUserStats(userID uuid.UUID) *UserStats {
	return &UserStats{
		UserID:          userID,
		MessageCount:    0,
		GroupCount:      0,
		ChannelCount:    0,
		ContactCount:    0,
		TransactionCount: 0,
	}
}
