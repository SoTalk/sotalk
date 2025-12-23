package response

import "time"

// PaymentRequestResponse is the HTTP response for a single payment request
type PaymentRequestResponse struct {
	PaymentRequest PaymentRequestDTO `json:"payment_request"`
}

// PaymentRequestsResponse is the HTTP response for multiple payment requests
type PaymentRequestsResponse struct {
	PaymentRequests []PaymentRequestDTO `json:"payment_requests"`
	Total           int                 `json:"total"`
}

// PaymentRequestDTO is the payment request data in response
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

// PaymentSendResponse is the HTTP response for sending a payment
type PaymentSendResponse struct {
	TransactionSig string `json:"transaction_sig"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}
