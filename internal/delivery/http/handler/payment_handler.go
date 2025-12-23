package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/payment"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// PaymentHandler handles payment HTTP requests
type PaymentHandler struct {
	paymentService payment.Service
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService payment.Service) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentRequest handles POST /api/v1/payments/request
func (h *PaymentHandler) CreatePaymentRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.CreatePaymentRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Convert to DTO
	createDTO := &dto.CreatePaymentRequestDTO{
		ConversationID: req.ConversationID,
		ToUserID:       req.ToUserID,
		Amount:         req.Amount,
		TokenMint:      req.TokenMint,
		Message:        req.Message,
		ExpiryMinutes:  req.ExpiryMinutes,
	}

	result, err := h.paymentService.CreatePaymentRequest(c.Request.Context(), userID, createDTO)
	if err != nil {
		logger.Error("Failed to create payment request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, mapPaymentRequestResponse(result))
}

// GetPaymentRequest handles GET /api/v1/payments/:id
func (h *PaymentHandler) GetPaymentRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Invalid payment ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.paymentService.GetPaymentRequest(c.Request.Context(), userID, paymentID)
	if err != nil {
		logger.Error("Failed to get payment request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapPaymentRequestResponse(result))
}

// GetPaymentHistory handles GET /api/v1/payments/history
func (h *PaymentHandler) GetPaymentHistory(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.GetPaymentHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.paymentService.GetUserPaymentRequests(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get payment history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_history_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapPaymentRequestsResponse(result))
}

// GetPendingPayments handles GET /api/v1/payments/pending
func (h *PaymentHandler) GetPendingPayments(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.paymentService.GetPendingPaymentRequests(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get pending payments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_pending_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapPaymentRequestsResponse(result))
}

// AcceptPaymentRequest handles POST /api/v1/payments/:id/accept
func (h *PaymentHandler) AcceptPaymentRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Invalid payment ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.AcceptPaymentRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	acceptDTO := &dto.AcceptPaymentRequestDTO{
		WalletID: req.WalletID,
	}

	result, err := h.paymentService.AcceptPaymentRequest(c.Request.Context(), userID, paymentID, acceptDTO)
	if err != nil {
		logger.Error("Failed to accept payment request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "accept_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapPaymentRequestResponse(result))
}

// RejectPaymentRequest handles POST /api/v1/payments/:id/reject
func (h *PaymentHandler) RejectPaymentRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Invalid payment ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.paymentService.RejectPaymentRequest(c.Request.Context(), userID, paymentID)
	if err != nil {
		logger.Error("Failed to reject payment request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "reject_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment request rejected"})
}

// CancelPaymentRequest handles POST /api/v1/payments/:id/cancel
func (h *PaymentHandler) CancelPaymentRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Invalid payment ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.paymentService.CancelPaymentRequest(c.Request.Context(), userID, paymentID)
	if err != nil {
		logger.Error("Failed to cancel payment request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "cancel_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment request cancelled"})
}

// ConfirmPayment handles POST /api/v1/payments/:id/confirm
func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Invalid payment ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.paymentService.ConfirmPayment(c.Request.Context(), userID, paymentID, req.TransactionSignature)
	if err != nil {
		logger.Error("Failed to confirm payment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "confirm_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment confirmed"})
}

// SendPayment handles POST /api/v1/payments/send
func (h *PaymentHandler) SendPayment(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.SendPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	sendDTO := &dto.PaymentSendDTO{
		ConversationID: req.ConversationID,
		ToUserID:       req.ToUserID,
		ToAddress:      req.ToAddress,
		Amount:         req.Amount,
		TokenMint:      req.TokenMint,
		Message:        req.Message,
		WalletID:       req.WalletID,
	}

	result, err := h.paymentService.SendDirectPayment(c.Request.Context(), userID, sendDTO)
	if err != nil {
		logger.Error("Failed to send payment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "send_payment_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.PaymentSendResponse{
		TransactionSig: result.TransactionSig,
		Status:         result.Status,
		Message:        result.Message,
	})
}

// Helper functions to map DTOs to HTTP responses

func mapPaymentRequestResponse(serviceDTO *dto.PaymentRequestResponse) response.PaymentRequestResponse {
	return response.PaymentRequestResponse{
		PaymentRequest: response.PaymentRequestDTO{
			ID:             serviceDTO.PaymentRequest.ID,
			ConversationID: serviceDTO.PaymentRequest.ConversationID,
			MessageID:      serviceDTO.PaymentRequest.MessageID,
			FromUserID:     serviceDTO.PaymentRequest.FromUserID,
			ToUserID:       serviceDTO.PaymentRequest.ToUserID,
			Amount:         serviceDTO.PaymentRequest.Amount,
			AmountSOL:      serviceDTO.PaymentRequest.AmountSOL,
			TokenMint:      serviceDTO.PaymentRequest.TokenMint,
			Message:        serviceDTO.PaymentRequest.Message,
			Status:         serviceDTO.PaymentRequest.Status,
			TransactionSig: serviceDTO.PaymentRequest.TransactionSig,
			ExpiresAt:      serviceDTO.PaymentRequest.ExpiresAt,
			CreatedAt:      serviceDTO.PaymentRequest.CreatedAt,
			UpdatedAt:      serviceDTO.PaymentRequest.UpdatedAt,
		},
	}
}

func mapPaymentRequestsResponse(serviceDTO *dto.PaymentRequestsResponse) response.PaymentRequestsResponse {
	payments := make([]response.PaymentRequestDTO, len(serviceDTO.PaymentRequests))
	for i, p := range serviceDTO.PaymentRequests {
		payments[i] = response.PaymentRequestDTO{
			ID:             p.ID,
			ConversationID: p.ConversationID,
			MessageID:      p.MessageID,
			FromUserID:     p.FromUserID,
			ToUserID:       p.ToUserID,
			Amount:         p.Amount,
			AmountSOL:      p.AmountSOL,
			TokenMint:      p.TokenMint,
			Message:        p.Message,
			Status:         p.Status,
			TransactionSig: p.TransactionSig,
			ExpiresAt:      p.ExpiresAt,
			CreatedAt:      p.CreatedAt,
			UpdatedAt:      p.UpdatedAt,
		}
	}

	return response.PaymentRequestsResponse{
		PaymentRequests: payments,
		Total:           serviceDTO.Total,
	}
}
