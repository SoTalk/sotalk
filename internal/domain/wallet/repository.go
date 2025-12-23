package wallet

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for wallet data operations
type Repository interface {
	// Wallet operations
	CreateWallet(ctx context.Context, wallet *Wallet) error
	FindWalletByID(ctx context.Context, id uuid.UUID) (*Wallet, error)
	FindWalletByAddress(ctx context.Context, address string) (*Wallet, error)
	FindWalletsByUserID(ctx context.Context, userID uuid.UUID) ([]*Wallet, error)
	FindDefaultWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error)
	UpdateWallet(ctx context.Context, wallet *Wallet) error
	DeleteWallet(ctx context.Context, id uuid.UUID) error

	// Transaction operations
	CreateTransaction(ctx context.Context, tx *Transaction) error
	FindTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	FindTransactionBySignature(ctx context.Context, signature string) (*Transaction, error)
	FindTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Transaction, error)
	FindTransactionsByWallet(ctx context.Context, address string, limit, offset int) ([]*Transaction, error)
	UpdateTransaction(ctx context.Context, tx *Transaction) error
	CountTransactionsByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
