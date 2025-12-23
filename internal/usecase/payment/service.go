package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/payment"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/domain/wallet"
	"github.com/yourusername/sotalk/internal/infrastructure/solana"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// WSBroadcaster defines the interface for WebSocket broadcasting
type WSBroadcaster interface {
	BroadcastPaymentRequest(ctx context.Context, toUserID uuid.UUID, payment dto.PaymentRequestDTO) error
	BroadcastPaymentUpdate(ctx context.Context, userID uuid.UUID, eventType string, payment dto.PaymentRequestDTO) error
}

// service implements the Service interface
type service struct {
	paymentRepo   payment.Repository
	userRepo      user.Repository
	walletRepo    wallet.Repository
	solanaClient  *solana.Client
	wsBroadcaster WSBroadcaster
}

// NewService creates a new payment service
func NewService(
	paymentRepo payment.Repository,
	userRepo user.Repository,
	walletRepo wallet.Repository,
	solanaClient *solana.Client,
	wsBroadcaster WSBroadcaster,
) Service {
	return &service{
		paymentRepo:   paymentRepo,
		userRepo:      userRepo,
		walletRepo:    walletRepo,
		solanaClient:  solanaClient,
		wsBroadcaster: wsBroadcaster,
	}
}

// CreatePaymentRequest creates a new payment request
func (s *service) CreatePaymentRequest(ctx context.Context, fromUserID uuid.UUID, req *dto.CreatePaymentRequestDTO) (*dto.PaymentRequestResponse, error) {
	// Parse UUIDs
	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation ID: %w", err)
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient user ID: %w", err)
	}

	// Validate user cannot send payment to self
	if fromUserID == toUserID {
		return nil, payment.ErrInvalidPaymentRecipient
	}

	// Validate amount
	if req.Amount == 0 {
		return nil, payment.ErrInvalidPaymentAmount
	}

	// Verify recipient exists
	_, err = s.userRepo.FindByID(ctx, toUserID)
	if err != nil {
		return nil, fmt.Errorf("recipient not found: %w", err)
	}

	// Set expiry time (default 15 minutes)
	expiryMinutes := req.ExpiryMinutes
	if expiryMinutes <= 0 {
		expiryMinutes = 15
	}
	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	// Create payment request entity
	paymentReq := &payment.PaymentRequest{
		ConversationID: conversationID,
		FromUserID:     fromUserID,
		ToUserID:       toUserID,
		Amount:         req.Amount,
		TokenMint:      req.TokenMint,
		Message:        req.Message,
		Status:         payment.PaymentStatusPending,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	if err := s.paymentRepo.CreatePaymentRequest(ctx, paymentReq); err != nil {
		logger.Error("Failed to create payment request", zap.Error(err))
		return nil, err
	}

	logger.Info("Payment request created",
		zap.String("payment_id", paymentReq.ID.String()),
		zap.String("from_user", fromUserID.String()),
		zap.String("to_user", toUserID.String()),
		zap.Uint64("amount", req.Amount),
	)

	// Broadcast payment request via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastPaymentRequest(ctx, toUserID, toPaymentRequestDTO(paymentReq)); err != nil {
			logger.Error("Failed to broadcast payment request",
				zap.String("payment_id", paymentReq.ID.String()),
				zap.Error(err),
			)
		}
	}

	return &dto.PaymentRequestResponse{
		PaymentRequest: toPaymentRequestDTO(paymentReq),
	}, nil
}

// GetPaymentRequest retrieves a payment request by ID
func (s *service) GetPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID) (*dto.PaymentRequestResponse, error) {
	paymentReq, err := s.paymentRepo.FindPaymentRequestByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Verify user is involved in payment
	if paymentReq.FromUserID != userID && paymentReq.ToUserID != userID {
		return nil, payment.ErrUnauthorizedPaymentAction
	}

	return &dto.PaymentRequestResponse{
		PaymentRequest: toPaymentRequestDTO(paymentReq),
	}, nil
}

// GetUserPaymentRequests retrieves all payment requests for a user
func (s *service) GetUserPaymentRequests(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.PaymentRequestsResponse, error) {
	// Set default limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	payments, err := s.paymentRepo.FindPaymentRequestsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.paymentRepo.CountPaymentRequestsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	paymentDTOs := make([]dto.PaymentRequestDTO, len(payments))
	for i, p := range payments {
		paymentDTOs[i] = toPaymentRequestDTO(p)
	}

	return &dto.PaymentRequestsResponse{
		PaymentRequests: paymentDTOs,
		Total:           total,
	}, nil
}

// GetPendingPaymentRequests retrieves pending payment requests for a user
func (s *service) GetPendingPaymentRequests(ctx context.Context, userID uuid.UUID) (*dto.PaymentRequestsResponse, error) {
	payments, err := s.paymentRepo.FindPendingPaymentRequestsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	paymentDTOs := make([]dto.PaymentRequestDTO, len(payments))
	for i, p := range payments {
		paymentDTOs[i] = toPaymentRequestDTO(p)
	}

	return &dto.PaymentRequestsResponse{
		PaymentRequests: paymentDTOs,
		Total:           len(payments),
	}, nil
}

// AcceptPaymentRequest accepts a payment request and initiates payment
func (s *service) AcceptPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID, req *dto.AcceptPaymentRequestDTO) (*dto.PaymentRequestResponse, error) {
	// Get payment request
	paymentReq, err := s.paymentRepo.FindPaymentRequestByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Verify user is the recipient
	if paymentReq.ToUserID != userID {
		return nil, payment.ErrUnauthorizedPaymentAction
	}

	// Check if payment can be accepted
	if !paymentReq.CanAccept() {
		if paymentReq.IsExpired() {
			return nil, payment.ErrPaymentRequestExpired
		}
		return nil, payment.ErrPaymentRequestAlreadyProcessed
	}

	// Parse wallet ID
	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet ID: %w", err)
	}

	// Get recipient's wallet (the one paying)
	senderWallet, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("sender wallet not found: %w", err)
	}

	// Verify wallet belongs to user
	if senderWallet.UserID != userID {
		return nil, payment.ErrUnauthorizedPaymentAction
	}

	// Get requester's default wallet (to receive payment)
	recipientWallet, err := s.walletRepo.FindDefaultWallet(ctx, paymentReq.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("recipient wallet not found: %w", err)
	}

	// Check sender has sufficient balance
	if senderWallet.Balance < paymentReq.Amount {
		return nil, payment.ErrInsufficientBalance
	}

	// Mark as accepted (payment will be sent by frontend with signed transaction)
	paymentReq.Accept()

	// Update payment request
	if err := s.paymentRepo.UpdatePaymentRequest(ctx, paymentReq); err != nil {
		return nil, err
	}

	logger.Info("Payment request accepted",
		zap.String("payment_id", paymentID.String()),
		zap.String("user_id", userID.String()),
		zap.String("from_wallet", senderWallet.Address),
		zap.String("to_wallet", recipientWallet.Address),
		zap.Uint64("amount", paymentReq.Amount),
	)

	// Broadcast payment accepted via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastPaymentUpdate(ctx, paymentReq.FromUserID, "payment.accepted", toPaymentRequestDTO(paymentReq)); err != nil {
			logger.Error("Failed to broadcast payment accepted",
				zap.String("payment_id", paymentID.String()),
				zap.Error(err),
			)
		}
	}

	return &dto.PaymentRequestResponse{
		PaymentRequest: toPaymentRequestDTO(paymentReq),
	}, nil
}

// RejectPaymentRequest rejects a payment request
func (s *service) RejectPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID) error {
	paymentReq, err := s.paymentRepo.FindPaymentRequestByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Verify user can reject
	if !paymentReq.CanReject(userID) {
		return payment.ErrUnauthorizedPaymentAction
	}

	paymentReq.Reject()

	if err := s.paymentRepo.UpdatePaymentRequest(ctx, paymentReq); err != nil {
		return err
	}

	logger.Info("Payment request rejected",
		zap.String("payment_id", paymentID.String()),
		zap.String("user_id", userID.String()),
	)

	// Broadcast payment rejected via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastPaymentUpdate(ctx, paymentReq.FromUserID, "payment.rejected", toPaymentRequestDTO(paymentReq)); err != nil {
			logger.Error("Failed to broadcast payment rejected",
				zap.String("payment_id", paymentID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// CancelPaymentRequest cancels a payment request
func (s *service) CancelPaymentRequest(ctx context.Context, userID, paymentID uuid.UUID) error {
	paymentReq, err := s.paymentRepo.FindPaymentRequestByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Verify user can cancel
	if !paymentReq.CanCancel(userID) {
		return payment.ErrUnauthorizedPaymentAction
	}

	paymentReq.Cancel()

	if err := s.paymentRepo.UpdatePaymentRequest(ctx, paymentReq); err != nil {
		return err
	}

	logger.Info("Payment request cancelled",
		zap.String("payment_id", paymentID.String()),
		zap.String("user_id", userID.String()),
	)

	// Broadcast payment canceled via WebSocket
	if s.wsBroadcaster != nil {
		if err := s.wsBroadcaster.BroadcastPaymentUpdate(ctx, paymentReq.ToUserID, "payment.canceled", toPaymentRequestDTO(paymentReq)); err != nil {
			logger.Error("Failed to broadcast payment canceled",
				zap.String("payment_id", paymentID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// SendDirectPayment sends a payment directly without request
func (s *service) SendDirectPayment(ctx context.Context, fromUserID uuid.UUID, req *dto.PaymentSendDTO) (*dto.PaymentSendResponse, error) {
	// Parse UUIDs
	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient user ID: %w", err)
	}

	// Validate user cannot send payment to self
	if fromUserID == toUserID {
		return nil, payment.ErrInvalidPaymentRecipient
	}

	// Validate amount
	if req.Amount == 0 {
		return nil, payment.ErrInvalidPaymentAmount
	}

	// Step 1: Get sender's wallet by walletID
	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet ID: %w", err)
	}

	senderWallet, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("sender wallet not found: %w", err)
	}

	// Verify wallet belongs to the user
	if senderWallet.UserID != fromUserID {
		return nil, payment.ErrUnauthorizedPaymentAction
	}

	// Step 2: Verify sufficient balance
	if senderWallet.Balance < req.Amount {
		return nil, payment.ErrInsufficientBalance
	}

	// Step 3: Get recipient wallet address
	recipientAddress := req.ToAddress
	if recipientAddress == "" {
		// If no address provided, get recipient's default wallet
		recipientWallet, err := s.walletRepo.FindDefaultWallet(ctx, toUserID)
		if err != nil {
			return nil, fmt.Errorf("recipient wallet not found: %w", err)
		}
		recipientAddress = recipientWallet.Address
	}

	// Verify recipient exists
	_, err = s.userRepo.FindByID(ctx, toUserID)
	if err != nil {
		return nil, fmt.Errorf("recipient not found: %w", err)
	}

	logger.Info("Preparing direct payment transaction",
		zap.String("from_user", fromUserID.String()),
		zap.String("from_wallet", senderWallet.Address),
		zap.String("to_user", toUserID.String()),
		zap.String("to_address", recipientAddress),
		zap.Uint64("amount", req.Amount),
	)

	// Step 4: Create unsigned transaction
	// Note: The transaction will be signed on the frontend since we don't store private keys
	unsignedTx, err := s.solanaClient.CreateTransferTransaction(
		ctx,
		senderWallet.Address,
		recipientAddress,
		req.Amount,
		req.TokenMint,
	)
	if err != nil {
		logger.Error("Failed to create transaction",
			zap.String("from", senderWallet.Address),
			zap.String("to", recipientAddress),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Step 5: Create transaction record in pending state
	txRecord := &wallet.Transaction{
		UserID:      fromUserID,
		FromAddress: senderWallet.Address,
		ToAddress:   recipientAddress,
		Amount:      req.Amount,
		TokenMint:   req.TokenMint,
		Type:        wallet.TransactionTypeSend,
		Status:      wallet.TransactionStatusPending,
		Metadata: wallet.TransactionMetadata{
			Message:    req.Message,
			RecipientUserID: toUserID.String(),
		},
		CreatedAt: time.Now(),
	}

	if err := s.walletRepo.CreateTransaction(ctx, txRecord); err != nil {
		logger.Error("Failed to create transaction record", zap.Error(err))
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	logger.Info("Direct payment transaction prepared",
		zap.String("transaction_id", txRecord.ID.String()),
		zap.String("from_user", fromUserID.String()),
		zap.String("to_user", toUserID.String()),
		zap.Uint64("amount", req.Amount),
	)

	// Return transaction data for frontend to sign and broadcast
	// The frontend will sign the transaction with the user's wallet and then
	// call a confirmation endpoint with the transaction signature
	return &dto.PaymentSendResponse{
		TransactionID:   txRecord.ID.String(),
		UnsignedTx:      unsignedTx,
		TransactionSig:  "", // Will be filled after frontend signs and broadcasts
		FromAddress:     senderWallet.Address,
		ToAddress:       recipientAddress,
		Amount:          req.Amount,
		TokenMint:       req.TokenMint,
		Status:          string(wallet.TransactionStatusPending),
		Message:         "Transaction prepared. Please sign with your wallet.",
		EstimatedFee:    5000, // 5000 lamports (0.000005 SOL) - typical Solana fee
	}, nil
}

// ConfirmPayment confirms a payment with a transaction signature from frontend
func (s *service) ConfirmPayment(ctx context.Context, userID, paymentID uuid.UUID, transactionSig string) error {
	// Get payment request
	paymentReq, err := s.paymentRepo.FindPaymentRequestByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Verify user is involved (either sender or recipient)
	if paymentReq.FromUserID != userID && paymentReq.ToUserID != userID {
		return payment.ErrUnauthorizedPaymentAction
	}

	// Verify payment is in accepted status
	if paymentReq.Status != payment.PaymentStatusAccepted {
		return fmt.Errorf("payment not in accepted status")
	}

	// Verify transaction on Solana network
	confirmed, err := s.solanaClient.ConfirmTransaction(ctx, transactionSig)
	if err != nil {
		logger.Error("Failed to confirm transaction",
			zap.String("signature", transactionSig),
			zap.Error(err),
		)
		return fmt.Errorf("failed to confirm transaction: %w", err)
	}

	if !confirmed {
		return fmt.Errorf("transaction not confirmed or failed")
	}

	// Get transaction details
	txDetail, err := s.solanaClient.GetTransaction(ctx, transactionSig)
	if err != nil {
		logger.Warn("Could not get transaction details", zap.Error(err))
	}

	// Mark payment as completed
	paymentReq.Complete(transactionSig)

	// Update payment request
	if err := s.paymentRepo.UpdatePaymentRequest(ctx, paymentReq); err != nil {
		return err
	}

	logger.Info("Payment confirmed",
		zap.String("payment_id", paymentID.String()),
		zap.String("transaction_sig", transactionSig),
		zap.Uint64("fee", txDetail.Fee),
	)

	// Broadcast payment confirmed via WebSocket to both users
	if s.wsBroadcaster != nil {
		paymentDTO := toPaymentRequestDTO(paymentReq)
		// Notify payment requester (recipient of funds)
		if err := s.wsBroadcaster.BroadcastPaymentUpdate(ctx, paymentReq.FromUserID, "payment.confirmed", paymentDTO); err != nil {
			logger.Error("Failed to broadcast payment confirmed to requester",
				zap.String("payment_id", paymentID.String()),
				zap.Error(err),
			)
		}
		// Notify payment sender
		if err := s.wsBroadcaster.BroadcastPaymentUpdate(ctx, paymentReq.ToUserID, "payment.confirmed", paymentDTO); err != nil {
			logger.Error("Failed to broadcast payment confirmed to sender",
				zap.String("payment_id", paymentID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// ExpireOldPaymentRequests expires old payment requests
func (s *service) ExpireOldPaymentRequests(ctx context.Context) error {
	return s.paymentRepo.ExpireOldPaymentRequests(ctx)
}

// Helper function to convert domain entity to DTO
func toPaymentRequestDTO(p *payment.PaymentRequest) dto.PaymentRequestDTO {
	var messageID *string
	if p.MessageID != nil {
		msgID := p.MessageID.String()
		messageID = &msgID
	}

	return dto.PaymentRequestDTO{
		ID:             p.ID.String(),
		ConversationID: p.ConversationID.String(),
		MessageID:      messageID,
		FromUserID:     p.FromUserID.String(),
		ToUserID:       p.ToUserID.String(),
		Amount:         p.Amount,
		AmountSOL:      p.GetAmountSOL(),
		TokenMint:      p.TokenMint,
		Message:        p.Message,
		Status:         string(p.Status),
		TransactionSig: p.TransactionSig,
		ExpiresAt:      p.ExpiresAt,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
