package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/wallet"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// WalletHandler handles wallet HTTP requests
type WalletHandler struct {
	walletService wallet.Service
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(walletService wallet.Service) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// AddWallet handles POST /api/v1/wallet
func (h *WalletHandler) AddWallet(c *gin.Context) {
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

	var req request.AddWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.AddWallet(c.Request.Context(), userID, req.Address, req.Label)
	if err != nil {
		logger.Error("Failed to add wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "add_wallet_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, mapWalletResponse(result))
}

// GetWallets handles GET /api/v1/wallet
func (h *WalletHandler) GetWallets(c *gin.Context) {
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

	result, err := h.walletService.GetUserWallets(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get wallets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_wallets_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapWalletsResponse(result))
}

// GetWallet handles GET /api/v1/wallet/:id
func (h *WalletHandler) GetWallet(c *gin.Context) {
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

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_wallet_id",
			Message: "Invalid wallet ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.GetWallet(c.Request.Context(), userID, walletID)
	if err != nil {
		logger.Error("Failed to get wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_wallet_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapWalletResponse(result))
}

// RefreshBalance handles POST /api/v1/wallet/:id/refresh
func (h *WalletHandler) RefreshBalance(c *gin.Context) {
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

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_wallet_id",
			Message: "Invalid wallet ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.RefreshBalance(c.Request.Context(), userID, walletID)
	if err != nil {
		logger.Error("Failed to refresh balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "refresh_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Auto-sync transactions from blockchain after refreshing balance
	// This ensures transactions are always up to date
	// Errors are logged but don't fail the request
	if syncErr := h.walletService.SyncTransactions(c.Request.Context(), userID, walletID); syncErr != nil {
		logger.Warn("Failed to auto-sync transactions during balance refresh",
			zap.Error(syncErr),
			zap.String("wallet_id", walletID.String()),
		)
		// Continue anyway - balance was refreshed successfully
	}

	c.JSON(http.StatusOK, mapWalletResponse(result))
}

// SetDefault handles POST /api/v1/wallet/:id/default
func (h *WalletHandler) SetDefault(c *gin.Context) {
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

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_wallet_id",
			Message: "Invalid wallet ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.walletService.SetDefaultWallet(c.Request.Context(), userID, walletID)
	if err != nil {
		logger.Error("Failed to set default wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "set_default_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default wallet updated"})
}

// DeleteWallet handles DELETE /api/v1/wallet/:id
func (h *WalletHandler) DeleteWallet(c *gin.Context) {
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

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_wallet_id",
			Message: "Invalid wallet ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.walletService.DeleteWallet(c.Request.Context(), userID, walletID)
	if err != nil {
		logger.Error("Failed to delete wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_wallet_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wallet deleted"})
}

// GetWalletByAddress handles GET /api/v1/wallet/by-address/:address
func (h *WalletHandler) GetWalletByAddress(c *gin.Context) {
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

	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_address",
			Message: "Wallet address is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.GetWalletByAddress(c.Request.Context(), userID, address)
	if err != nil {
		logger.Error("Failed to get wallet by address", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_wallet_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapWalletResponse(result))
}

// GetTransactionHistory handles GET /api/v1/wallet/transactions
func (h *WalletHandler) GetTransactionHistory(c *gin.Context) {
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

	var req request.GetTransactionHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.GetTransactionHistory(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_transactions_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, mapTransactionsResponse(result))
}

// GetTransactionBySignature handles GET /api/v1/wallet/transactions/by-signature/:signature
func (h *WalletHandler) GetTransactionBySignature(c *gin.Context) {
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

	signature := c.Param("signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_signature",
			Message: "Transaction signature is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.GetTransactionBySignature(c.Request.Context(), userID, signature)
	if err != nil {
		logger.Error("Failed to get transaction by signature", zap.Error(err))
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "transaction_not_found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response.TransactionResponse{
		Transaction: response.TransactionDTO{
			ID:          result.ID,
			UserID:      result.UserID,
			Signature:   result.Signature,
			FromAddress: result.FromAddress,
			ToAddress:   result.ToAddress,
			Amount:      result.Amount,
			AmountSOL:   result.AmountSOL,
			TokenMint:   result.TokenMint,
			Type:        result.Type,
			Status:      result.Status,
			Fee:         result.Fee,
			FeeSOL:      result.FeeSOL,
			BlockTime:   result.BlockTime,
			Metadata: response.TransactionMetadataDTO{
				Memo:      result.Metadata.Memo,
				Program:   result.Metadata.Program,
				Slot:      result.Metadata.Slot,
				ErrorMsg:  result.Metadata.ErrorMsg,
				MessageID: result.Metadata.MessageID,
			},
			CreatedAt: result.CreatedAt,
			UpdatedAt: result.UpdatedAt,
		},
	})
}

// SyncTransactions handles POST /api/v1/wallet/:id/sync
func (h *WalletHandler) SyncTransactions(c *gin.Context) {
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

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_wallet_id",
			Message: "Invalid wallet ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = h.walletService.SyncTransactions(c.Request.Context(), userID, walletID)
	if err != nil {
		logger.Error("Failed to sync transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transactions synced"})
}

// RequestAirdrop handles POST /api/v1/wallet/:id/airdrop
func (h *WalletHandler) RequestAirdrop(c *gin.Context) {
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

	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_wallet_id",
			Message: "Invalid wallet ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.AirdropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to 1 SOL if no amount specified
		req.Amount = 1.0
	}

	// Validate amount (max 2 SOL per request on devnet)
	if req.Amount <= 0 || req.Amount > 2.0 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_amount",
			Message: "Amount must be between 0 and 2 SOL",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.walletService.RequestAirdrop(c.Request.Context(), userID, walletID, req.Amount)
	if err != nil {
		logger.Error("Failed to request airdrop", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "airdrop_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Auto-sync transactions after airdrop to record the transaction
	// Give it a moment for the transaction to be confirmed on-chain
	if syncErr := h.walletService.SyncTransactions(c.Request.Context(), userID, walletID); syncErr != nil {
		logger.Warn("Failed to auto-sync transactions after airdrop",
			zap.Error(syncErr),
			zap.String("wallet_id", walletID.String()),
		)
		// Continue anyway - airdrop was requested successfully
	}

	c.JSON(http.StatusOK, result)
}

// Helper functions to map service DTOs to HTTP responses
func mapWalletResponse(serviceDTO *dto.WalletResponse) response.WalletResponse {
	tokenBalances := make(map[string]response.TokenBalanceDTO)
	for mint, balance := range serviceDTO.Wallet.TokenBalances {
		tokenBalances[mint] = response.TokenBalanceDTO{
			Mint:     balance.Mint,
			Amount:   balance.Amount,
			Decimals: balance.Decimals,
			Symbol:   balance.Symbol,
			UiAmount: balance.UiAmount,
		}
	}

	return response.WalletResponse{
		Wallet: response.WalletDTO{
			ID:            serviceDTO.Wallet.ID,
			UserID:        serviceDTO.Wallet.UserID,
			Address:       serviceDTO.Wallet.Address,
			Label:         serviceDTO.Wallet.Label,
			Balance:       serviceDTO.Wallet.Balance,
			BalanceSOL:    serviceDTO.Wallet.BalanceSOL,
			TokenBalances: tokenBalances,
			IsDefault:     serviceDTO.Wallet.IsDefault,
			CreatedAt:     serviceDTO.Wallet.CreatedAt,
			UpdatedAt:     serviceDTO.Wallet.UpdatedAt,
		},
	}
}

func mapWalletsResponse(serviceDTO *dto.WalletsResponse) response.WalletsResponse {
	wallets := make([]response.WalletDTO, len(serviceDTO.Wallets))
	for i, w := range serviceDTO.Wallets {
		tokenBalances := make(map[string]response.TokenBalanceDTO)
		for mint, balance := range w.TokenBalances {
			tokenBalances[mint] = response.TokenBalanceDTO{
				Mint:     balance.Mint,
				Amount:   balance.Amount,
				Decimals: balance.Decimals,
				Symbol:   balance.Symbol,
				UiAmount: balance.UiAmount,
			}
		}

		wallets[i] = response.WalletDTO{
			ID:            w.ID,
			UserID:        w.UserID,
			Address:       w.Address,
			Label:         w.Label,
			Balance:       w.Balance,
			BalanceSOL:    w.BalanceSOL,
			TokenBalances: tokenBalances,
			IsDefault:     w.IsDefault,
			CreatedAt:     w.CreatedAt,
			UpdatedAt:     w.UpdatedAt,
		}
	}
	return response.WalletsResponse{
		Wallets: wallets,
		Total:   serviceDTO.Total,
	}
}

func mapTransactionsResponse(serviceDTO *dto.TransactionsResponse) response.TransactionsResponse {
	transactions := make([]response.TransactionDTO, len(serviceDTO.Transactions))
	for i, tx := range serviceDTO.Transactions {
		transactions[i] = response.TransactionDTO{
			ID:          tx.ID,
			UserID:      tx.UserID,
			Signature:   tx.Signature,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			Amount:      tx.Amount,
			AmountSOL:   tx.AmountSOL,
			TokenMint:   tx.TokenMint,
			Type:        tx.Type,
			Status:      tx.Status,
			Fee:         tx.Fee,
			FeeSOL:      tx.FeeSOL,
			BlockTime:   tx.BlockTime,
			Metadata: response.TransactionMetadataDTO{
				Memo:      tx.Metadata.Memo,
				Program:   tx.Metadata.Program,
				Slot:      tx.Metadata.Slot,
				ErrorMsg:  tx.Metadata.ErrorMsg,
				MessageID: tx.Metadata.MessageID,
			},
			CreatedAt: tx.CreatedAt,
			UpdatedAt: tx.UpdatedAt,
		}
	}
	return response.TransactionsResponse{
		Transactions: transactions,
		Total:        serviceDTO.Total,
	}
}
