package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/domain/wallet"
	"github.com/yourusername/sotalk/internal/infrastructure/solana"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

type service struct {
	walletRepo  wallet.Repository
	userRepo    user.Repository
	solanaClient *solana.Client
}

// NewService creates a new wallet service
func NewService(
	walletRepo wallet.Repository,
	userRepo user.Repository,
	solanaClient *solana.Client,
) Service {
	return &service{
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		solanaClient: solanaClient,
	}
}

// AddWallet adds a new wallet for a user
func (s *service) AddWallet(ctx context.Context, userID uuid.UUID, address, label string) (*dto.WalletResponse, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate Solana address
	if !s.solanaClient.VerifyAddress(address) {
		return nil, wallet.ErrInvalidAddress
	}

	// Check if wallet already exists
	existingWallet, _ := s.walletRepo.FindWalletByAddress(ctx, address)
	if existingWallet != nil {
		return nil, wallet.ErrWalletAlreadyExists
	}

	// Create wallet entity
	w := wallet.NewWallet(userID, address, label)

	// Get initial balance from Solana
	balance, err := s.solanaClient.GetBalance(ctx, address)
	if err == nil {
		w.UpdateBalance(balance)
	}

	// Get token balances
	tokenAccounts, err := s.solanaClient.GetTokenAccounts(ctx, address)
	if err == nil {
		for _, ta := range tokenAccounts {
			w.SetTokenBalance(ta.Mint, wallet.TokenBalance{
				Mint:     ta.Mint,
				Decimals: ta.Decimals,
				UiAmount: ta.UiAmount,
				Symbol:   ta.Symbol,
			})
		}
	}

	// Check if this should be the default wallet
	wallets, _ := s.walletRepo.FindWalletsByUserID(ctx, userID)
	if len(wallets) == 0 {
		w.SetDefault()
	}

	// Save to database
	if err := s.walletRepo.CreateWallet(ctx, w); err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return &dto.WalletResponse{
		Wallet: toWalletDTO(w),
	}, nil
}

// GetWallet gets a wallet by ID
func (s *service) GetWallet(ctx context.Context, userID, walletID uuid.UUID) (*dto.WalletResponse, error) {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if w.UserID != userID {
		return nil, wallet.ErrWalletNotFound
	}

	return &dto.WalletResponse{
		Wallet: toWalletDTO(w),
	}, nil
}

// GetWalletByAddress gets a wallet by its Solana address
func (s *service) GetWalletByAddress(ctx context.Context, userID uuid.UUID, address string) (*dto.WalletResponse, error) {
	w, err := s.walletRepo.FindWalletByAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	// Check authorization - user must own the wallet
	if w.UserID != userID {
		return nil, wallet.ErrWalletNotFound
	}

	return &dto.WalletResponse{
		Wallet: toWalletDTO(w),
	}, nil
}

// GetUserWallets gets all wallets for a user
func (s *service) GetUserWallets(ctx context.Context, userID uuid.UUID) (*dto.WalletsResponse, error) {
	wallets, err := s.walletRepo.FindWalletsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	walletDTOs := make([]dto.WalletDTO, len(wallets))
	for i, w := range wallets {
		walletDTOs[i] = toWalletDTO(w)
	}

	return &dto.WalletsResponse{
		Wallets: walletDTOs,
		Total:   len(walletDTOs),
	}, nil
}

// GetDefaultWallet gets the default wallet for a user
func (s *service) GetDefaultWallet(ctx context.Context, userID uuid.UUID) (*dto.WalletResponse, error) {
	w, err := s.walletRepo.FindDefaultWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.WalletResponse{
		Wallet: toWalletDTO(w),
	}, nil
}

// SetDefaultWallet sets a wallet as default
func (s *service) SetDefaultWallet(ctx context.Context, userID, walletID uuid.UUID) error {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return err
	}

	// Check authorization
	if w.UserID != userID {
		return wallet.ErrWalletNotFound
	}

	// Unset current default
	currentDefault, err := s.walletRepo.FindDefaultWallet(ctx, userID)
	if err == nil && currentDefault != nil {
		currentDefault.IsDefault = false
		s.walletRepo.UpdateWallet(ctx, currentDefault)
	}

	// Set new default
	w.SetDefault()
	return s.walletRepo.UpdateWallet(ctx, w)
}

// DeleteWallet deletes a wallet
func (s *service) DeleteWallet(ctx context.Context, userID, walletID uuid.UUID) error {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return err
	}

	// Check authorization
	if w.UserID != userID {
		return wallet.ErrWalletNotFound
	}

	// Prevent deleting default wallet if there are other wallets
	if w.IsDefault {
		wallets, _ := s.walletRepo.FindWalletsByUserID(ctx, userID)
		if len(wallets) > 1 {
			return wallet.ErrCannotDeleteDefault
		}
	}

	return s.walletRepo.DeleteWallet(ctx, walletID)
}

// RefreshBalance refreshes wallet balance from Solana
func (s *service) RefreshBalance(ctx context.Context, userID, walletID uuid.UUID) (*dto.WalletResponse, error) {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if w.UserID != userID {
		return nil, wallet.ErrWalletNotFound
	}

	// Get balance from Solana
	balance, err := s.solanaClient.GetBalance(ctx, w.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	w.UpdateBalance(balance)

	// Get token balances
	tokenAccounts, err := s.solanaClient.GetTokenAccounts(ctx, w.Address)
	if err == nil {
		for _, ta := range tokenAccounts {
			w.SetTokenBalance(ta.Mint, wallet.TokenBalance{
				Mint:     ta.Mint,
				Decimals: ta.Decimals,
				UiAmount: ta.UiAmount,
				Symbol:   ta.Symbol,
			})
		}
	}

	// Update in database
	if err := s.walletRepo.UpdateWallet(ctx, w); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &dto.WalletResponse{
		Wallet: toWalletDTO(w),
	}, nil
}

// RequestAirdrop requests an airdrop of SOL from the faucet (devnet/testnet only)
func (s *service) RequestAirdrop(ctx context.Context, userID, walletID uuid.UUID, amountSOL float64) (*dto.AirdropResponse, error) {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if w.UserID != userID {
		return nil, wallet.ErrWalletNotFound
	}

	// Convert SOL to lamports (1 SOL = 1,000,000,000 lamports)
	amountLamports := uint64(amountSOL * 1_000_000_000)

	// Request airdrop from Solana
	signature, err := s.solanaClient.RequestAirdrop(ctx, w.Address, amountLamports)
	if err != nil {
		return nil, fmt.Errorf("failed to request airdrop: %w", err)
	}

	// Wait a moment for the transaction to process, then refresh balance
	// Note: In production, you might want to poll for confirmation
	// For now, we'll just return the signature and the UI can refresh

	return &dto.AirdropResponse{
		Signature:     signature,
		Amount:        amountLamports,
		AmountSOL:     amountSOL,
		WalletAddress: w.Address,
		Network:       s.solanaClient.GetNetwork(),
		Message:       fmt.Sprintf("Airdrop of %.2f SOL requested. Signature: %s", amountSOL, signature),
	}, nil
}

// GetTransactionHistory gets transaction history for a user
func (s *service) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.TransactionsResponse, error) {
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	transactions, err := s.walletRepo.FindTransactionsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	txDTOs := make([]dto.TransactionDTO, len(transactions))
	for i, tx := range transactions {
		txDTOs[i] = toTransactionDTO(tx)
	}

	total, _ := s.walletRepo.CountTransactionsByUserID(ctx, userID)

	return &dto.TransactionsResponse{
		Transactions: txDTOs,
		Total:        int(total),
	}, nil
}

// GetTransaction gets a transaction by ID
func (s *service) GetTransaction(ctx context.Context, userID uuid.UUID, txID uuid.UUID) (*dto.TransactionResponse, error) {
	tx, err := s.walletRepo.FindTransactionByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if tx.UserID != userID {
		return nil, wallet.ErrTransactionNotFound
	}

	return &dto.TransactionResponse{
		Transaction: toTransactionDTO(tx),
	}, nil
}

// GetTransactionBySignature gets a transaction by its Solana signature
func (s *service) GetTransactionBySignature(ctx context.Context, userID uuid.UUID, signature string) (*dto.TransactionDTO, error) {
	tx, err := s.walletRepo.FindTransactionBySignature(ctx, signature)
	if err != nil {
		return nil, err
	}

	// Check authorization - user must own the transaction
	if tx.UserID != userID {
		return nil, wallet.ErrTransactionNotFound
	}

	txDTO := toTransactionDTO(tx)
	return &txDTO, nil
}

// GetWalletTransactions gets transactions for a specific wallet
func (s *service) GetWalletTransactions(ctx context.Context, userID, walletID uuid.UUID, limit, offset int) (*dto.TransactionsResponse, error) {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if w.UserID != userID {
		return nil, wallet.ErrWalletNotFound
	}

	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	transactions, err := s.walletRepo.FindTransactionsByWallet(ctx, w.Address, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	txDTOs := make([]dto.TransactionDTO, len(transactions))
	for i, tx := range transactions {
		txDTOs[i] = toTransactionDTO(tx)
	}

	return &dto.TransactionsResponse{
		Transactions: txDTOs,
		Total:        len(txDTOs),
	}, nil
}

// SyncTransactions syncs transactions from Solana blockchain
func (s *service) SyncTransactions(ctx context.Context, userID, walletID uuid.UUID) error {
	w, err := s.walletRepo.FindWalletByID(ctx, walletID)
	if err != nil {
		return err
	}

	// Check authorization
	if w.UserID != userID {
		return wallet.ErrWalletNotFound
	}

	// Get transaction signatures from Solana
	signatures, err := s.solanaClient.GetTransactionSignatures(ctx, w.Address, 50)
	if err != nil {
		return fmt.Errorf("failed to get signatures: %w", err)
	}

	// Process each signature
	for _, sig := range signatures {
		// Check if transaction already exists
		_, err := s.walletRepo.FindTransactionBySignature(ctx, sig.Signature)
		if err == nil {
			continue // Already synced
		}

		// Get transaction details with parsed amounts and addresses
		detail, err := s.solanaClient.GetTransaction(ctx, sig.Signature)
		if err != nil {
			logger.Warn("Failed to get transaction details",
				zap.String("signature", sig.Signature),
				zap.Error(err),
			)
			continue // Skip failed transactions
		}

		// Determine transaction type based on wallet address
		var txType wallet.TransactionType
		if detail.FromAddress == w.Address {
			// User sent SOL
			txType = wallet.TransactionTypeSend
		} else if detail.ToAddress == w.Address {
			// User received SOL
			txType = wallet.TransactionTypeReceive
		} else {
			// Other transaction types (could be contract interaction, etc.)
			txType = wallet.TransactionTypeOther
		}

		// Create transaction entity with real data
		tx := wallet.NewTransaction(
			userID,
			sig.Signature,
			detail.FromAddress, // Real sender address
			detail.ToAddress,   // Real receiver address
			detail.Amount,      // Real transfer amount in lamports
			txType,
		)

		tx.SetFee(detail.Fee)
		if detail.Success {
			tx.MarkAsConfirmed(detail.BlockTime)
		} else {
			tx.MarkAsFailed("Transaction failed on chain")
		}

		// Save transaction to database
		if err := s.walletRepo.CreateTransaction(ctx, tx); err != nil {
			logger.Warn("Failed to save transaction",
				zap.String("signature", sig.Signature),
				zap.Error(err),
			)
		}
	}

	return nil
}

// Helper functions

func toWalletDTO(w *wallet.Wallet) dto.WalletDTO {
	tokenBalances := make(map[string]dto.TokenBalanceDTO)
	for mint, balance := range w.TokenBalances {
		tokenBalances[mint] = dto.TokenBalanceDTO{
			Mint:     balance.Mint,
			Amount:   balance.Amount,
			Decimals: balance.Decimals,
			Symbol:   balance.Symbol,
			UiAmount: balance.UiAmount,
		}
	}

	return dto.WalletDTO{
		ID:            w.ID.String(),
		UserID:        w.UserID.String(),
		Address:       w.Address,
		Label:         w.Label,
		Balance:       w.Balance,
		BalanceSOL:    w.GetBalanceInSOL(),
		TokenBalances: tokenBalances,
		IsDefault:     w.IsDefault,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

func toTransactionDTO(tx *wallet.Transaction) dto.TransactionDTO {
	amountSOL := float64(tx.Amount) / 1_000_000_000
	feeSOL := float64(tx.Fee) / 1_000_000_000

	return dto.TransactionDTO{
		ID:          tx.ID.String(),
		UserID:      tx.UserID.String(),
		Signature:   tx.Signature,
		FromAddress: tx.FromAddress,
		ToAddress:   tx.ToAddress,
		Amount:      tx.Amount,
		AmountSOL:   amountSOL,
		TokenMint:   tx.TokenMint,
		Type:        string(tx.Type),
		Status:      string(tx.Status),
		Fee:         tx.Fee,
		FeeSOL:      feeSOL,
		BlockTime:   tx.BlockTime,
		Metadata: dto.TransactionMetadataDTO{
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
