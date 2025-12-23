package wallet

import "errors"

var (
	// Wallet errors
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyExists = errors.New("wallet already exists")
	ErrInvalidAddress      = errors.New("invalid wallet address")
	ErrDefaultWalletNotSet = errors.New("no default wallet set")
	ErrCannotDeleteDefault = errors.New("cannot delete default wallet")

	// Transaction errors
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	ErrInvalidSignature         = errors.New("invalid transaction signature")
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrTransactionFailed        = errors.New("transaction failed")

	// Solana RPC errors
	ErrRPCConnectionFailed = errors.New("failed to connect to Solana RPC")
	ErrRPCRequestFailed    = errors.New("Solana RPC request failed")
)
