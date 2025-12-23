package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the interface for payment operations
type Service interface {
	// CreatePaymentRequest creates a new payment request
	CreatePaymentRequest(ctx context.Context, fromUserID uuid.UUID, req *dto.CreatePaymentRequestDTO) (*dto.PaymentRequestResponse, error)

	// GetPaymentRequest retrieves a payment request by ID
	GetPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID) (*dto.PaymentRequestResponse, error)

	// GetUserPaymentRequests retrieves all payment requests for a user
	GetUserPaymentRequests(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.PaymentRequestsResponse, error)

	// GetPendingPaymentRequests retrieves pending payment requests for a user
	GetPendingPaymentRequests(ctx context.Context, userID uuid.UUID) (*dto.PaymentRequestsResponse, error)

	// AcceptPaymentRequest accepts a payment request and initiates payment
	AcceptPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID, req *dto.AcceptPaymentRequestDTO) (*dto.PaymentRequestResponse, error)

	// RejectPaymentRequest rejects a payment request
	RejectPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID) error

	// CancelPaymentRequest cancels a payment request
	CancelPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID) error

	// SendDirectPayment sends a payment directly without request
	SendDirectPayment(ctx context.Context, fromUserID uuid.UUID, req *dto.PaymentSendDTO) (*dto.PaymentSendResponse, error)

	// ConfirmPayment confirms a payment with a transaction signature from frontend
	ConfirmPayment(ctx context.Context, userID, paymentID uuid.UUID, transactionSig string) error

	// ExpireOldPaymentRequests expires old payment requests
	ExpireOldPaymentRequests(ctx context.Context) error
}
