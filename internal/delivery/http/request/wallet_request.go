package request

// AddWalletRequest is the HTTP request for adding a wallet
type AddWalletRequest struct {
	Address string `json:"address" binding:"required"`
	Label   string `json:"label"`
}

// GetTransactionHistoryRequest is the HTTP request for getting transaction history
type GetTransactionHistoryRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// AirdropRequest is the HTTP request for requesting an airdrop
type AirdropRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}
