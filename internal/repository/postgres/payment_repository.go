package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/payment"
	"gorm.io/gorm"
)

// PaymentRequestRepository implements payment.Repository
type PaymentRequestRepository struct {
	db *gorm.DB
}

// NewPaymentRequestRepository creates a new payment repository
func NewPaymentRequestRepository(db *gorm.DB) payment.Repository {
	return &PaymentRequestRepository{db: db}
}

// CreatePaymentRequest creates a new payment request
func (r *PaymentRequestRepository) CreatePaymentRequest(ctx context.Context, paymentReq *payment.PaymentRequest) error {
	model := toPaymentRequestModel(paymentReq)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create payment request: %w", err)
	}

	// Update the domain entity with generated values
	paymentReq.ID = model.ID
	paymentReq.CreatedAt = model.CreatedAt
	paymentReq.UpdatedAt = model.UpdatedAt

	return nil
}

// FindPaymentRequestByID retrieves a payment request by ID
func (r *PaymentRequestRepository) FindPaymentRequestByID(ctx context.Context, id uuid.UUID) (*payment.PaymentRequest, error) {
	var model PaymentRequest
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, payment.ErrPaymentRequestNotFound
		}
		return nil, fmt.Errorf("failed to find payment request: %w", err)
	}

	return toDomainPaymentRequest(&model), nil
}

// FindPaymentRequestsByUserID retrieves payment requests for a user (sent or received)
func (r *PaymentRequestRepository) FindPaymentRequestsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*payment.PaymentRequest, error) {
	var models []PaymentRequest

	query := r.db.WithContext(ctx).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find payment requests: %w", err)
	}

	payments := make([]*payment.PaymentRequest, len(models))
	for i, model := range models {
		payments[i] = toDomainPaymentRequest(&model)
	}

	return payments, nil
}

// FindPendingPaymentRequestsByUserID retrieves pending payment requests for a user
func (r *PaymentRequestRepository) FindPendingPaymentRequestsByUserID(ctx context.Context, userID uuid.UUID) ([]*payment.PaymentRequest, error) {
	var models []PaymentRequest

	if err := r.db.WithContext(ctx).
		Where("to_user_id = ? AND status = ?", userID, string(payment.PaymentStatusPending)).
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find pending payment requests: %w", err)
	}

	payments := make([]*payment.PaymentRequest, len(models))
	for i, model := range models {
		payments[i] = toDomainPaymentRequest(&model)
	}

	return payments, nil
}

// FindPaymentRequestsByConversationID retrieves payment requests in a conversation
func (r *PaymentRequestRepository) FindPaymentRequestsByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*payment.PaymentRequest, error) {
	var models []PaymentRequest

	query := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find payment requests: %w", err)
	}

	payments := make([]*payment.PaymentRequest, len(models))
	for i, model := range models {
		payments[i] = toDomainPaymentRequest(&model)
	}

	return payments, nil
}

// CountPaymentRequestsByUserID counts payment requests for a user
func (r *PaymentRequestRepository) CountPaymentRequestsByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&PaymentRequest{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payment requests: %w", err)
	}

	return int(count), nil
}

// UpdatePaymentRequest updates a payment request
func (r *PaymentRequestRepository) UpdatePaymentRequest(ctx context.Context, paymentReq *payment.PaymentRequest) error {
	model := toPaymentRequestModel(paymentReq)

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return fmt.Errorf("failed to update payment request: %w", err)
	}

	// Update timestamps
	paymentReq.UpdatedAt = model.UpdatedAt

	return nil
}

// DeletePaymentRequest deletes a payment request
func (r *PaymentRequestRepository) DeletePaymentRequest(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&PaymentRequest{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete payment request: %w", err)
	}

	return nil
}

// ExpireOldPaymentRequests marks old payment requests as expired
func (r *PaymentRequestRepository) ExpireOldPaymentRequests(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Model(&PaymentRequest{}).
		Where("status = ? AND expires_at < ?", string(payment.PaymentStatusPending), time.Now()).
		Update("status", string(payment.PaymentStatusExpired)).Error; err != nil {
		return fmt.Errorf("failed to expire old payment requests: %w", err)
	}

	return nil
}

// Helper functions to map between domain and GORM models

func toPaymentRequestModel(domainPayment *payment.PaymentRequest) *PaymentRequest {
	return &PaymentRequest{
		ID:             domainPayment.ID,
		ConversationID: domainPayment.ConversationID,
		MessageID:      domainPayment.MessageID,
		FromUserID:     domainPayment.FromUserID,
		ToUserID:       domainPayment.ToUserID,
		Amount:         domainPayment.Amount,
		TokenMint:      domainPayment.TokenMint,
		Message:        domainPayment.Message,
		Status:         string(domainPayment.Status),
		TransactionSig: domainPayment.TransactionSig,
		ExpiresAt:      domainPayment.ExpiresAt,
		CreatedAt:      domainPayment.CreatedAt,
		UpdatedAt:      domainPayment.UpdatedAt,
	}
}

func toDomainPaymentRequest(model *PaymentRequest) *payment.PaymentRequest {
	return &payment.PaymentRequest{
		ID:             model.ID,
		ConversationID: model.ConversationID,
		MessageID:      model.MessageID,
		FromUserID:     model.FromUserID,
		ToUserID:       model.ToUserID,
		Amount:         model.Amount,
		TokenMint:      model.TokenMint,
		Message:        model.Message,
		Status:         payment.PaymentStatus(model.Status),
		TransactionSig: model.TransactionSig,
		ExpiresAt:      model.ExpiresAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}
