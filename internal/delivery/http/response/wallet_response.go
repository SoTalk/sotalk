package response

import "time"

// WalletResponse is the HTTP response for wallet operations
type WalletResponse struct {
	Wallet WalletDTO `json:"wallet"`
}

// WalletsResponse is the HTTP response for multiple wallets
type WalletsResponse struct {
	Wallets []WalletDTO `json:"wallets"`
	Total   int         `json:"total"`
}

// WalletDTO is the wallet data in response
type WalletDTO struct {
	ID            string                     `json:"id"`
	UserID        string                     `json:"user_id"`
	Address       string                     `json:"address"`
	Label         string                     `json:"label"`
	Balance       uint64                     `json:"balance"`
	BalanceSOL    float64                    `json:"balance_sol"`
	TokenBalances map[string]TokenBalanceDTO `json:"token_balances"`
	IsDefault     bool                       `json:"is_default"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

// TokenBalanceDTO represents token balance in response
type TokenBalanceDTO struct {
	Mint     string  `json:"mint"`
	Amount   uint64  `json:"amount"`
	Decimals int     `json:"decimals"`
	Symbol   string  `json:"symbol"`
	UiAmount float64 `json:"ui_amount"`
}

// TransactionResponse is the HTTP response for transaction operations
type TransactionResponse struct {
	Transaction TransactionDTO `json:"transaction"`
}

// TransactionsResponse is the HTTP response for multiple transactions
type TransactionsResponse struct {
	Transactions []TransactionDTO `json:"transactions"`
	Total        int              `json:"total"`
}

// TransactionDTO is the transaction data in response
type TransactionDTO struct {
	ID          string                  `json:"id"`
	UserID      string                  `json:"user_id"`
	Signature   string                  `json:"signature"`
	FromAddress string                  `json:"from_address"`
	ToAddress   string                  `json:"to_address"`
	Amount      uint64                  `json:"amount"`
	AmountSOL   float64                 `json:"amount_sol,omitempty"`
	TokenMint   *string                 `json:"token_mint,omitempty"`
	Type        string                  `json:"type"`
	Status      string                  `json:"status"`
	Fee         uint64                  `json:"fee"`
	FeeSOL      float64                 `json:"fee_sol"`
	BlockTime   *time.Time              `json:"block_time,omitempty"`
	Metadata    TransactionMetadataDTO  `json:"metadata,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// TransactionMetadataDTO represents transaction metadata in response
type TransactionMetadataDTO struct {
	Memo      string `json:"memo,omitempty"`
	Program   string `json:"program,omitempty"`
	Slot      uint64 `json:"slot,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}
