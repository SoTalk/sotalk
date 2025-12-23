package wallet

import (
	"time"

	"github.com/google/uuid"
)

// Wallet represents a Solana wallet
type Wallet struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Address       string
	Label         string
	Balance       uint64 // lamports (1 SOL = 1,000,000,000 lamports)
	TokenBalances map[string]TokenBalance
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TokenBalance represents a token balance
type TokenBalance struct {
	Mint     string  // Token mint address
	Amount   uint64  // Token amount (raw)
	Decimals int     // Token decimals
	Symbol   string  // Token symbol (e.g., USDC)
	UiAmount float64 // Human-readable amount
}

// Transaction represents a Solana transaction
type Transaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Signature   string
	FromAddress string
	ToAddress   string
	Amount      uint64
	TokenMint   *string // nil for SOL, token mint for SPL tokens
	Type        TransactionType
	Status      TransactionStatus
	Fee         uint64
	BlockTime   *time.Time
	Metadata    TransactionMetadata
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeSend    TransactionType = "send"
	TransactionTypeReceive TransactionType = "receive"
	TransactionTypeSwap    TransactionType = "swap"
	TransactionTypeStake   TransactionType = "stake"
	TransactionTypeUnstake TransactionType = "unstake"
	TransactionTypeOther   TransactionType = "other"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusDropped   TransactionStatus = "dropped"
)

// TransactionMetadata contains additional transaction information
type TransactionMetadata struct {
	Memo            string `json:"memo,omitempty"`
	Message         string `json:"message,omitempty"`         // User message with payment
	Program         string `json:"program,omitempty"`
	Slot            uint64 `json:"slot,omitempty"`
	Commitment      string `json:"commitment,omitempty"`
	ErrorMsg        string `json:"error_msg,omitempty"`
	MessageID       string `json:"message_id,omitempty"`       // Link to chat message
	RecipientUserID string `json:"recipient_user_id,omitempty"` // Recipient user ID
}

// NewWallet creates a new wallet entity
func NewWallet(userID uuid.UUID, address, label string) *Wallet {
	now := time.Now()
	return &Wallet{
		ID:            uuid.New(),
		UserID:        userID,
		Address:       address,
		Label:         label,
		Balance:       0,
		TokenBalances: make(map[string]TokenBalance),
		IsDefault:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewTransaction creates a new transaction entity
func NewTransaction(userID uuid.UUID, signature, fromAddress, toAddress string, amount uint64, txType TransactionType) *Transaction {
	now := time.Now()
	return &Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Signature:   signature,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Type:        txType,
		Status:      TransactionStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SetDefault sets this wallet as the default
func (w *Wallet) SetDefault() {
	w.IsDefault = true
	w.UpdatedAt = time.Now()
}

// UpdateBalance updates the wallet balance
func (w *Wallet) UpdateBalance(balance uint64) {
	w.Balance = balance
	w.UpdatedAt = time.Now()
}

// SetTokenBalance sets or updates a token balance
func (w *Wallet) SetTokenBalance(mint string, balance TokenBalance) {
	if w.TokenBalances == nil {
		w.TokenBalances = make(map[string]TokenBalance)
	}
	w.TokenBalances[mint] = balance
	w.UpdatedAt = time.Now()
}

// GetBalanceInSOL returns balance in SOL
func (w *Wallet) GetBalanceInSOL() float64 {
	return float64(w.Balance) / 1_000_000_000
}

// MarkAsConfirmed marks transaction as confirmed
func (t *Transaction) MarkAsConfirmed(blockTime time.Time) {
	t.Status = TransactionStatusConfirmed
	t.BlockTime = &blockTime
	t.UpdatedAt = time.Now()
}

// MarkAsFailed marks transaction as failed
func (t *Transaction) MarkAsFailed(errorMsg string) {
	t.Status = TransactionStatusFailed
	t.Metadata.ErrorMsg = errorMsg
	t.UpdatedAt = time.Now()
}

// SetFee sets the transaction fee
func (t *Transaction) SetFee(fee uint64) {
	t.Fee = fee
	t.UpdatedAt = time.Now()
}
