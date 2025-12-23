package payment

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for payment data operations
type Repository interface {
	// CreatePaymentRequest creates a new payment request
	CreatePaymentRequest(ctx context.Context, payment *PaymentRequest) error

	// FindPaymentRequestByID retrieves a payment request by ID
	FindPaymentRequestByID(ctx context.Context, id uuid.UUID) (*PaymentRequest, error)

	// FindPaymentRequestsByUserID retrieves payment requests for a user (sent or received)
	FindPaymentRequestsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*PaymentRequest, error)

	// FindPendingPaymentRequestsByUserID retrieves pending payment requests for a user
	FindPendingPaymentRequestsByUserID(ctx context.Context, userID uuid.UUID) ([]*PaymentRequest, error)

	// FindPaymentRequestsByConversationID retrieves payment requests in a conversation
	FindPaymentRequestsByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*PaymentRequest, error)

	// CountPaymentRequestsByUserID counts payment requests for a user
	CountPaymentRequestsByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// UpdatePaymentRequest updates a payment request
	UpdatePaymentRequest(ctx context.Context, payment *PaymentRequest) error

	// DeletePaymentRequest deletes a payment request
	DeletePaymentRequest(ctx context.Context, id uuid.UUID) error

	// ExpireOldPaymentRequests marks old payment requests as expired
	ExpireOldPaymentRequests(ctx context.Context) error
}
