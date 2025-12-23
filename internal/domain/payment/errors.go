package payment

import "errors"

var (
	// ErrPaymentRequestNotFound is returned when payment request is not found
	ErrPaymentRequestNotFound = errors.New("payment request not found")

	// ErrPaymentRequestExpired is returned when payment request has expired
	ErrPaymentRequestExpired = errors.New("payment request has expired")

	// ErrPaymentRequestAlreadyProcessed is returned when payment request is already processed
	ErrPaymentRequestAlreadyProcessed = errors.New("payment request already processed")

	// ErrInvalidPaymentAmount is returned when payment amount is invalid
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")

	// ErrInvalidPaymentRecipient is returned when trying to send payment to self
	ErrInvalidPaymentRecipient = errors.New("cannot send payment to yourself")

	// ErrUnauthorizedPaymentAction is returned when user is not authorized to perform action
	ErrUnauthorizedPaymentAction = errors.New("unauthorized to perform this payment action")

	// ErrInsufficientBalance is returned when wallet has insufficient balance
	ErrInsufficientBalance = errors.New("insufficient wallet balance")

	// ErrPaymentTransactionFailed is returned when blockchain transaction fails
	ErrPaymentTransactionFailed = errors.New("payment transaction failed")
)
