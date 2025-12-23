package solana

import "time"

// TokenAccount represents a token account
type TokenAccount struct {
	Mint     string
	Amount   string
	Decimals int
	UiAmount float64
	Symbol   string
}

// TransactionSignature represents a transaction signature
type TransactionSignature struct {
	Signature string
	Slot      uint64
	BlockTime time.Time
	Err       bool
}

// TransactionDetail represents detailed transaction information
type TransactionDetail struct {
	Signature   string
	Slot        uint64
	BlockTime   time.Time
	Fee         uint64
	Success     bool
	FromAddress string // Actual sender address
	ToAddress   string // Actual receiver address
	Amount      uint64 // Transfer amount in lamports
	Type        string // "send", "receive", or "other"
}

// TransactionResult represents the result of sending a transaction
type TransactionResult struct {
	Signature string
	Success   bool
	Error     string
}
