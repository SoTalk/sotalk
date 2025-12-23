package payment

import (
	"time"

	"github.com/google/uuid"
)

// PaymentRequest represents a payment request in chat
type PaymentRequest struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	MessageID      *uuid.UUID
	FromUserID     uuid.UUID
	ToUserID       uuid.UUID
	Amount         uint64 // lamports
	TokenMint      *string
	Message        string
	Status         PaymentStatus
	TransactionSig *string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// PaymentStatus represents payment request status
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusAccepted  PaymentStatus = "accepted"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusRejected  PaymentStatus = "rejected"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// IsValid checks if payment status is valid
func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending, PaymentStatusAccepted, PaymentStatusCompleted,
		PaymentStatusRejected, PaymentStatusExpired, PaymentStatusCancelled:
		return true
	}
	return false
}

// IsFinal checks if payment status is final
func (s PaymentStatus) IsFinal() bool {
	switch s {
	case PaymentStatusCompleted, PaymentStatusRejected, PaymentStatusExpired, PaymentStatusCancelled:
		return true
	}
	return false
}

// Accept marks payment request as accepted
func (p *PaymentRequest) Accept() {
	p.Status = PaymentStatusAccepted
	p.UpdatedAt = time.Now()
}

// Complete marks payment request as completed with transaction signature
func (p *PaymentRequest) Complete(txSig string) {
	p.Status = PaymentStatusCompleted
	p.TransactionSig = &txSig
	p.UpdatedAt = time.Now()
}

// Reject marks payment request as rejected
func (p *PaymentRequest) Reject() {
	p.Status = PaymentStatusRejected
	p.UpdatedAt = time.Now()
}

// Cancel marks payment request as cancelled
func (p *PaymentRequest) Cancel() {
	p.Status = PaymentStatusCancelled
	p.UpdatedAt = time.Now()
}

// Expire marks payment request as expired
func (p *PaymentRequest) Expire() {
	p.Status = PaymentStatusExpired
	p.UpdatedAt = time.Now()
}

// IsExpired checks if payment request has expired
func (p *PaymentRequest) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// CanAccept checks if payment can be accepted
func (p *PaymentRequest) CanAccept() bool {
	return p.Status == PaymentStatusPending && !p.IsExpired()
}

// CanCancel checks if payment can be cancelled
func (p *PaymentRequest) CanCancel(userID uuid.UUID) bool {
	return p.Status == PaymentStatusPending && p.FromUserID == userID
}

// CanReject checks if payment can be rejected
func (p *PaymentRequest) CanReject(userID uuid.UUID) bool {
	return p.Status == PaymentStatusPending && p.ToUserID == userID
}

// GetAmountSOL returns amount in SOL
func (p *PaymentRequest) GetAmountSOL() float64 {
	return float64(p.Amount) / 1_000_000_000
}
