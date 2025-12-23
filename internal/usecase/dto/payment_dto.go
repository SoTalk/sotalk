package dto

import (
	"time"
)

// PaymentRequestDTO represents a payment request for API responses
type PaymentRequestDTO struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	MessageID      *string   `json:"message_id,omitempty"`
	FromUserID     string    `json:"from_user_id"`
	ToUserID       string    `json:"to_user_id"`
	Amount         uint64    `json:"amount"`
	AmountSOL      float64   `json:"amount_sol"`
	TokenMint      *string   `json:"token_mint,omitempty"`
	Message        string    `json:"message"`
	Status         string    `json:"status"`
	TransactionSig *string   `json:"transaction_sig,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreatePaymentRequestDTO represents request to create payment
type CreatePaymentRequestDTO struct {
	ConversationID string  `json:"conversation_id"`
	ToUserID       string  `json:"to_user_id"`
	Amount         uint64  `json:"amount"`
	TokenMint      *string `json:"token_mint,omitempty"`
	Message        string  `json:"message"`
	ExpiryMinutes  int     `json:"expiry_minutes"` // Default: 15 minutes
}

// AcceptPaymentRequestDTO represents accepting a payment request
type AcceptPaymentRequestDTO struct {
	WalletID string `json:"wallet_id"` // Wallet to use for payment
}

// PaymentRequestResponse is the response containing a single payment request
type PaymentRequestResponse struct {
	PaymentRequest PaymentRequestDTO `json:"payment_request"`
}

// PaymentRequestsResponse is the response containing multiple payment requests
type PaymentRequestsResponse struct {
	PaymentRequests []PaymentRequestDTO `json:"payment_requests"`
	Total           int                 `json:"total"`
}

// PaymentSendDTO represents a direct payment (not a request)
type PaymentSendDTO struct {
	ConversationID string  `json:"conversation_id"`
	ToUserID       string  `json:"to_user_id"`
	ToAddress      string  `json:"to_address"` // Recipient wallet address
	Amount         uint64  `json:"amount"`
	TokenMint      *string `json:"token_mint,omitempty"`
	Message        string  `json:"message"`
	WalletID       string  `json:"wallet_id"` // Sender wallet ID
}

// PaymentSendResponse is the response for sending payment
type PaymentSendResponse struct {
	TransactionID   string  `json:"transaction_id"`   // Internal transaction ID
	UnsignedTx      string  `json:"unsigned_tx"`      // Base64 encoded unsigned transaction
	TransactionSig  string  `json:"transaction_sig"`  // Signature after signing (empty initially)
	FromAddress     string  `json:"from_address"`     // Sender address
	ToAddress       string  `json:"to_address"`       // Recipient address
	Amount          uint64  `json:"amount"`           // Amount in lamports
	TokenMint       *string `json:"token_mint"`       // Token mint if SPL token
	Status          string  `json:"status"`           // Transaction status
	Message         string  `json:"message"`          // Status message
	EstimatedFee    uint64  `json:"estimated_fee"`    // Estimated transaction fee
}
