package request

// CreatePaymentRequestRequest is the HTTP request for creating a payment request
type CreatePaymentRequestRequest struct {
	ConversationID string  `json:"conversation_id" binding:"required"`
	ToUserID       string  `json:"to_user_id" binding:"required"`
	Amount         uint64  `json:"amount" binding:"required"`
	TokenMint      *string `json:"token_mint"`
	Message        string  `json:"message"`
	ExpiryMinutes  int     `json:"expiry_minutes"` // Default: 15
}

// AcceptPaymentRequestRequest is the HTTP request for accepting a payment
type AcceptPaymentRequestRequest struct {
	WalletID string `json:"wallet_id" binding:"required"`
}

// SendPaymentRequest is the HTTP request for sending direct payment
type SendPaymentRequest struct {
	ConversationID string  `json:"conversation_id" binding:"required"`
	ToUserID       string  `json:"to_user_id" binding:"required"`
	ToAddress      string  `json:"to_address" binding:"required"`
	Amount         uint64  `json:"amount" binding:"required"`
	TokenMint      *string `json:"token_mint"`
	Message        string  `json:"message"`
	WalletID       string  `json:"wallet_id" binding:"required"`
}

// GetPaymentHistoryRequest is the HTTP request for getting payment history
type GetPaymentHistoryRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// ConfirmPaymentRequest is the HTTP request for confirming a payment
type ConfirmPaymentRequest struct {
	TransactionSignature string `json:"transaction_signature" binding:"required"`
}
