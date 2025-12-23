package wallet

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the wallet use case interface
type Service interface {
	// Wallet operations
	AddWallet(ctx context.Context, userID uuid.UUID, address, label string) (*dto.WalletResponse, error)
	GetWallet(ctx context.Context, userID, walletID uuid.UUID) (*dto.WalletResponse, error)
	GetWalletByAddress(ctx context.Context, userID uuid.UUID, address string) (*dto.WalletResponse, error)
	GetUserWallets(ctx context.Context, userID uuid.UUID) (*dto.WalletsResponse, error)
	GetDefaultWallet(ctx context.Context, userID uuid.UUID) (*dto.WalletResponse, error)
	SetDefaultWallet(ctx context.Context, userID, walletID uuid.UUID) error
	DeleteWallet(ctx context.Context, userID, walletID uuid.UUID) error
	RefreshBalance(ctx context.Context, userID, walletID uuid.UUID) (*dto.WalletResponse, error)
	RequestAirdrop(ctx context.Context, userID, walletID uuid.UUID, amountSOL float64) (*dto.AirdropResponse, error)

	// Transaction operations
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.TransactionsResponse, error)
	GetTransaction(ctx context.Context, userID uuid.UUID, txID uuid.UUID) (*dto.TransactionResponse, error)
	GetTransactionBySignature(ctx context.Context, userID uuid.UUID, signature string) (*dto.TransactionDTO, error)
	GetWalletTransactions(ctx context.Context, userID, walletID uuid.UUID, limit, offset int) (*dto.TransactionsResponse, error)
	SyncTransactions(ctx context.Context, userID, walletID uuid.UUID) error
}
