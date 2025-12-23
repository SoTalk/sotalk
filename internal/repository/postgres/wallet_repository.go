package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/wallet"
	"gorm.io/gorm"
)

// walletRepository implements wallet.Repository interface
type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *gorm.DB) wallet.Repository {
	return &walletRepository{db: db}
}

// CreateWallet creates a new wallet
func (r *walletRepository) CreateWallet(ctx context.Context, w *wallet.Wallet) error {
	dbWallet := toWalletModel(w)
	result := r.db.WithContext(ctx).Create(dbWallet)
	if result.Error != nil {
		return result.Error
	}

	w.ID = dbWallet.ID
	w.CreatedAt = dbWallet.CreatedAt
	w.UpdatedAt = dbWallet.UpdatedAt

	return nil
}

// FindWalletByID finds a wallet by ID
func (r *walletRepository) FindWalletByID(ctx context.Context, id uuid.UUID) (*wallet.Wallet, error) {
	var dbWallet Wallet
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbWallet)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, wallet.ErrWalletNotFound
		}
		return nil, result.Error
	}

	return toDomainWallet(&dbWallet), nil
}

// FindWalletByAddress finds a wallet by address
func (r *walletRepository) FindWalletByAddress(ctx context.Context, address string) (*wallet.Wallet, error) {
	var dbWallet Wallet
	result := r.db.WithContext(ctx).Where("address = ?", address).First(&dbWallet)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, wallet.ErrWalletNotFound
		}
		return nil, result.Error
	}

	return toDomainWallet(&dbWallet), nil
}

// FindWalletsByUserID finds all wallets by user ID
func (r *walletRepository) FindWalletsByUserID(ctx context.Context, userID uuid.UUID) ([]*wallet.Wallet, error) {
	var dbWallets []Wallet

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at ASC").
		Find(&dbWallets)

	if result.Error != nil {
		return nil, result.Error
	}

	wallets := make([]*wallet.Wallet, len(dbWallets))
	for i, dbWallet := range dbWallets {
		wallets[i] = toDomainWallet(&dbWallet)
	}

	return wallets, nil
}

// FindDefaultWallet finds the default wallet for a user
func (r *walletRepository) FindDefaultWallet(ctx context.Context, userID uuid.UUID) (*wallet.Wallet, error) {
	var dbWallet Wallet
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", userID, true).
		First(&dbWallet)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, wallet.ErrDefaultWalletNotSet
		}
		return nil, result.Error
	}

	return toDomainWallet(&dbWallet), nil
}

// UpdateWallet updates a wallet
func (r *walletRepository) UpdateWallet(ctx context.Context, w *wallet.Wallet) error {
	dbWallet := toWalletModel(w)
	result := r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", w.ID).Updates(dbWallet)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return wallet.ErrWalletNotFound
	}

	return nil
}

// DeleteWallet soft deletes a wallet
func (r *walletRepository) DeleteWallet(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Wallet{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return wallet.ErrWalletNotFound
	}

	return nil
}

// CreateTransaction creates a new transaction
func (r *walletRepository) CreateTransaction(ctx context.Context, tx *wallet.Transaction) error {
	dbTx := toTransactionModel(tx)
	result := r.db.WithContext(ctx).Create(dbTx)
	if result.Error != nil {
		return result.Error
	}

	tx.ID = dbTx.ID
	tx.CreatedAt = dbTx.CreatedAt
	tx.UpdatedAt = dbTx.UpdatedAt

	return nil
}

// FindTransactionByID finds a transaction by ID
func (r *walletRepository) FindTransactionByID(ctx context.Context, id uuid.UUID) (*wallet.Transaction, error) {
	var dbTx Transaction
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbTx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, wallet.ErrTransactionNotFound
		}
		return nil, result.Error
	}

	return toDomainTransaction(&dbTx), nil
}

// FindTransactionBySignature finds a transaction by signature
func (r *walletRepository) FindTransactionBySignature(ctx context.Context, signature string) (*wallet.Transaction, error) {
	var dbTx Transaction
	result := r.db.WithContext(ctx).Where("signature = ?", signature).First(&dbTx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, wallet.ErrTransactionNotFound
		}
		return nil, result.Error
	}

	return toDomainTransaction(&dbTx), nil
}

// FindTransactionsByUserID finds transactions by user ID with pagination
func (r *walletRepository) FindTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*wallet.Transaction, error) {
	var dbTxs []Transaction

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbTxs)

	if result.Error != nil {
		return nil, result.Error
	}

	transactions := make([]*wallet.Transaction, len(dbTxs))
	for i, dbTx := range dbTxs {
		transactions[i] = toDomainTransaction(&dbTx)
	}

	return transactions, nil
}

// FindTransactionsByWallet finds transactions by wallet address with pagination
func (r *walletRepository) FindTransactionsByWallet(ctx context.Context, address string, limit, offset int) ([]*wallet.Transaction, error) {
	var dbTxs []Transaction

	result := r.db.WithContext(ctx).
		Where("from_address = ? OR to_address = ?", address, address).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbTxs)

	if result.Error != nil {
		return nil, result.Error
	}

	transactions := make([]*wallet.Transaction, len(dbTxs))
	for i, dbTx := range dbTxs {
		transactions[i] = toDomainTransaction(&dbTx)
	}

	return transactions, nil
}

// UpdateTransaction updates a transaction
func (r *walletRepository) UpdateTransaction(ctx context.Context, tx *wallet.Transaction) error {
	dbTx := toTransactionModel(tx)
	result := r.db.WithContext(ctx).Model(&Transaction{}).Where("id = ?", tx.ID).Updates(dbTx)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return wallet.ErrTransactionNotFound
	}

	return nil
}

// CountTransactionsByUserID counts transactions for a user
func (r *walletRepository) CountTransactionsByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&Transaction{}).Where("user_id = ?", userID).Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// Mapper functions

// toWalletModel converts domain Wallet to GORM Wallet model
func toWalletModel(w *wallet.Wallet) *Wallet {
	// Serialize token balances to JSON
	tokenBalancesJSON, _ := json.Marshal(w.TokenBalances)

	return &Wallet{
		ID:            w.ID,
		UserID:        w.UserID,
		Address:       w.Address,
		Label:         w.Label,
		Balance:       w.Balance,
		TokenBalances: string(tokenBalancesJSON),
		IsDefault:     w.IsDefault,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

// toDomainWallet converts GORM Wallet model to domain Wallet
func toDomainWallet(w *Wallet) *wallet.Wallet {
	// Deserialize token balances from JSON
	var tokenBalances map[string]wallet.TokenBalance
	if w.TokenBalances != "" {
		json.Unmarshal([]byte(w.TokenBalances), &tokenBalances)
	}
	if tokenBalances == nil {
		tokenBalances = make(map[string]wallet.TokenBalance)
	}

	return &wallet.Wallet{
		ID:            w.ID,
		UserID:        w.UserID,
		Address:       w.Address,
		Label:         w.Label,
		Balance:       w.Balance,
		TokenBalances: tokenBalances,
		IsDefault:     w.IsDefault,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

// toTransactionModel converts domain Transaction to GORM Transaction model
func toTransactionModel(tx *wallet.Transaction) *Transaction {
	// Serialize metadata to JSON
	metadataJSON, _ := json.Marshal(tx.Metadata)

	return &Transaction{
		ID:          tx.ID,
		UserID:      tx.UserID,
		Signature:   tx.Signature,
		FromAddress: tx.FromAddress,
		ToAddress:   tx.ToAddress,
		Amount:      tx.Amount,
		TokenMint:   tx.TokenMint,
		Type:        string(tx.Type),
		Status:      string(tx.Status),
		Fee:         tx.Fee,
		BlockTime:   tx.BlockTime,
		Metadata:    string(metadataJSON),
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
}

// toDomainTransaction converts GORM Transaction model to domain Transaction
func toDomainTransaction(tx *Transaction) *wallet.Transaction {
	// Deserialize metadata from JSON
	var metadata wallet.TransactionMetadata
	if tx.Metadata != "" {
		json.Unmarshal([]byte(tx.Metadata), &metadata)
	}

	return &wallet.Transaction{
		ID:          tx.ID,
		UserID:      tx.UserID,
		Signature:   tx.Signature,
		FromAddress: tx.FromAddress,
		ToAddress:   tx.ToAddress,
		Amount:      tx.Amount,
		TokenMint:   tx.TokenMint,
		Type:        wallet.TransactionType(tx.Type),
		Status:      wallet.TransactionStatus(tx.Status),
		Fee:         tx.Fee,
		BlockTime:   tx.BlockTime,
		Metadata:    metadata,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
}
