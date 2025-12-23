package solana

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// Client represents a Solana RPC client
type Client struct {
	rpcClient *rpc.Client
	network   string
}

// Config holds configuration for Solana client
type Config struct {
	RPCEndpoint string
	Network     string
}

// NewClient creates a new Solana RPC client
func NewClient(config Config) (*Client, error) {
	rpcClient := rpc.New(config.RPCEndpoint)

	logger.Info("Solana RPC client initialized",
		zap.String("endpoint", config.RPCEndpoint),
		zap.String("network", config.Network),
	)

	return &Client{
		rpcClient: rpcClient,
		network:   config.Network,
	}, nil
}

// GetBalance gets the SOL balance for an address
func (c *Client) GetBalance(ctx context.Context, address string) (uint64, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %w", err)
	}

	balance, err := c.rpcClient.GetBalance(
		ctx,
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance.Value, nil
}

// GetTokenAccounts gets all token accounts for an address
func (c *Client) GetTokenAccounts(ctx context.Context, address string) ([]TokenAccount, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get token accounts
	resp, err := c.rpcClient.GetTokenAccountsByOwner(
		ctx,
		pubKey,
		&rpc.GetTokenAccountsConfig{
			ProgramId: &solana.TokenProgramID,
		},
		&rpc.GetTokenAccountsOpts{
			Commitment: rpc.CommitmentFinalized,
			Encoding:   solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token accounts: %w", err)
	}

	tokenAccounts := make([]TokenAccount, 0, len(resp.Value))
	for _, account := range resp.Value {
		// Parse account data
		if account.Account.Data != nil {
			rawJSON := account.Account.Data.GetRawJSON()
			if rawJSON != nil {
				var parsed map[string]interface{}
				if err := json.Unmarshal(rawJSON, &parsed); err == nil {
					info, ok := parsed["parsed"].(map[string]interface{})
					if ok {
						infoData, ok := info["info"].(map[string]interface{})
						if ok {
							mint, _ := infoData["mint"].(string)
							tokenAmount, _ := infoData["tokenAmount"].(map[string]interface{})

							if tokenAmount != nil {
								amount, _ := tokenAmount["amount"].(string)
								decimals, _ := tokenAmount["decimals"].(float64)
								uiAmount, _ := tokenAmount["uiAmount"].(float64)

								tokenAccounts = append(tokenAccounts, TokenAccount{
									Mint:     mint,
									Amount:   amount,
									Decimals: int(decimals),
									UiAmount: uiAmount,
								})
							}
						}
					}
				}
			}
		}
	}

	return tokenAccounts, nil
}

// GetTransactionSignatures gets transaction signatures for an address
func (c *Client) GetTransactionSignatures(ctx context.Context, address string, limit int) ([]TransactionSignature, error) {
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	if limit <= 0 || limit > 1000 {
		limit = 100 // Default limit
	}

	resp, err := c.rpcClient.GetSignaturesForAddress(
		ctx,
		pubKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %w", err)
	}

	signatures := make([]TransactionSignature, len(resp))
	for i, sig := range resp {
		signatures[i] = TransactionSignature{
			Signature: sig.Signature.String(),
			Slot:      sig.Slot,
			BlockTime: sig.BlockTime.Time(),
			Err:       sig.Err != nil,
		}
	}

	return signatures, nil
}

// GetTransaction gets transaction details with parsed transfer information
func (c *Client) GetTransaction(ctx context.Context, signature string) (*TransactionDetail, error) {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	tx, err := c.rpcClient.GetTransaction(
		ctx,
		sig,
		&rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentFinalized,
			Encoding:   solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if tx == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	detail := &TransactionDetail{
		Signature: signature,
		Slot:      tx.Slot,
		BlockTime: tx.BlockTime.Time(),
		Fee:       tx.Meta.Fee,
		Success:   tx.Meta.Err == nil,
		Type:      "other",
	}

	// Parse transaction to extract transfer details
	if tx.Transaction != nil {
		parseTransactionTransfer(tx, detail)
	}

	return detail, nil
}

// parseTransactionTransfer extracts SOL transfer information from a transaction
func parseTransactionTransfer(tx *rpc.GetTransactionResult, detail *TransactionDetail) {
	// Get account keys from the transaction message
	var accountKeys []solana.PublicKey

	// Extract account keys from the transaction
	if tx.Transaction != nil {
		decodedTx, err := tx.Transaction.GetTransaction()
		if err == nil && decodedTx != nil && decodedTx.Message.AccountKeys != nil {
			accountKeys = decodedTx.Message.AccountKeys
		}
	}

	// Also include loaded addresses from metadata
	if tx.Meta != nil {
		if tx.Meta.LoadedAddresses.Writable != nil {
			accountKeys = append(accountKeys, tx.Meta.LoadedAddresses.Writable...)
		}
		if tx.Meta.LoadedAddresses.ReadOnly != nil {
			accountKeys = append(accountKeys, tx.Meta.LoadedAddresses.ReadOnly...)
		}
	}

	// Parse pre and post balances to detect SOL transfers
	// This works for all transaction types (System Transfer, Airdrop, etc.)
	if tx.Meta != nil && len(tx.Meta.PreBalances) > 0 && len(tx.Meta.PostBalances) > 0 {
		// Track receiver
		var receiverIndex = -1
		var maxBalanceIncrease uint64 = 0

		for i := 0; i < len(tx.Meta.PreBalances) && i < len(tx.Meta.PostBalances); i++ {
			preBalance := tx.Meta.PreBalances[i]
			postBalance := tx.Meta.PostBalances[i]

			// First account is the fee payer/signer (sender)
			if i == 0 {
				if i < len(accountKeys) {
					detail.FromAddress = accountKeys[i].String()
				}

				// Calculate sender's balance change (excluding fee)
				// If sender balance decreased more than just the fee, they sent SOL
				if preBalance > postBalance {
					totalDecrease := preBalance - postBalance
					if totalDecrease > detail.Fee {
						detail.Amount = totalDecrease - detail.Fee
					}
				}
				continue
			}

			// Detect balance increase (receiver)
			if postBalance > preBalance {
				balanceChange := postBalance - preBalance
				// Track the account with the largest balance increase
				if balanceChange > maxBalanceIncrease {
					maxBalanceIncrease = balanceChange
					receiverIndex = i

					// If we didn't detect amount from sender, use receiver's increase
					if detail.Amount == 0 {
						detail.Amount = balanceChange
					}
				}
			}
		}

		// Set receiver address
		if receiverIndex > 0 && receiverIndex < len(accountKeys) {
			detail.ToAddress = accountKeys[receiverIndex].String()
		}

		// Determine transaction type
		if detail.Amount > 0 && detail.ToAddress != "" && detail.FromAddress != "" {
			detail.Type = "transfer"
		}
	}
}

// VerifyAddress verifies if a Solana address is valid
func (c *Client) VerifyAddress(address string) bool {
	_, err := solana.PublicKeyFromBase58(address)
	return err == nil
}

// GetNetwork returns the network name
func (c *Client) GetNetwork() string {
	return c.network
}

// SendTransaction sends a pre-signed transaction to the Solana network
// NOTE: Transaction must be signed by the client before calling this
func (c *Client) SendTransaction(ctx context.Context, signedTx []byte) (*TransactionResult, error) {
	// Send the transaction
	sig, err := c.rpcClient.SendRawTransaction(ctx, signedTx)
	if err != nil {
		return &TransactionResult{
			Success: false,
			Error:   err.Error(),
		}, fmt.Errorf("failed to send transaction: %w", err)
	}

	logger.Info("Transaction sent to Solana network",
		zap.String("signature", sig.String()),
		zap.String("network", c.network),
	)

	return &TransactionResult{
		Signature: sig.String(),
		Success:   true,
	}, nil
}

// ConfirmTransaction confirms a transaction and waits for finality
func (c *Client) ConfirmTransaction(ctx context.Context, signature string) (bool, error) {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	// Get transaction status
	tx, err := c.rpcClient.GetTransaction(
		ctx,
		sig,
		&rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentConfirmed,
		},
	)
	if err != nil {
		return false, fmt.Errorf("failed to get transaction: %w", err)
	}

	if tx == nil {
		return false, fmt.Errorf("transaction not found")
	}

	// Check if transaction was successful
	success := tx.Meta != nil && tx.Meta.Err == nil

	logger.Info("Transaction confirmation checked",
		zap.String("signature", signature),
		zap.Bool("success", success),
	)

	return success, nil
}

// GetRecentBlockhash gets a recent blockhash for transaction creation
func (c *Client) GetRecentBlockhash(ctx context.Context) (string, error) {
	resp, err := c.rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	return resp.Value.Blockhash.String(), nil
}

// EstimateTransactionFee estimates the fee for a transaction
func (c *Client) EstimateTransactionFee(ctx context.Context, message []byte) (uint64, error) {
	// Get fee for message
	resp, err := c.rpcClient.GetFeeForMessage(ctx, string(message), rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get fee: %w", err)
	}

	if resp.Value == nil {
		return 5000, nil // Default fee (5000 lamports)
	}

	return *resp.Value, nil
}

// RequestAirdrop requests an airdrop of SOL (devnet/testnet only)
func (c *Client) RequestAirdrop(ctx context.Context, address string, amountLamports uint64) (string, error) {
	// Check if network allows airdrops
	if c.network == "mainnet" {
		return "", fmt.Errorf("airdrops not available on mainnet")
	}

	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}

	// Request airdrop (max 2 SOL per request on devnet)
	if amountLamports > 2_000_000_000 {
		amountLamports = 2_000_000_000
	}

	sig, err := c.rpcClient.RequestAirdrop(
		ctx,
		pubKey,
		amountLamports,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return "", fmt.Errorf("failed to request airdrop: %w", err)
	}

	logger.Info("Airdrop requested",
		zap.String("address", address),
		zap.Uint64("amount_lamports", amountLamports),
		zap.String("signature", sig.String()),
		zap.String("network", c.network),
	)

	return sig.String(), nil
}

// CreateTransferTransaction creates an unsigned SOL/SPL token transfer transaction
// Returns base64 encoded transaction that needs to be signed by the client
func (c *Client) CreateTransferTransaction(ctx context.Context, fromAddress, toAddress string, amount uint64, tokenMint *string) (string, error) {
	// Validate addresses
	if _, err := solana.PublicKeyFromBase58(fromAddress); err != nil {
		return "", fmt.Errorf("invalid from address: %w", err)
	}

	if _, err := solana.PublicKeyFromBase58(toAddress); err != nil {
		return "", fmt.Errorf("invalid to address: %w", err)
	}

	// Get recent blockhash
	recentBlockhash, err := c.GetRecentBlockhash(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	blockhash, err := solana.HashFromBase58(recentBlockhash)
	if err != nil {
		return "", fmt.Errorf("invalid blockhash: %w", err)
	}

	// For now, we'll return transaction instruction data that the frontend can use
	// The actual transaction building and signing should happen on the frontend
	// This is a placeholder that returns the necessary data
	txData := map[string]interface{}{
		"from":            fromAddress,
		"to":              toAddress,
		"amount":          amount,
		"token_mint":      tokenMint,
		"recent_blockhash": recentBlockhash,
	}

	// In a real implementation, you would:
	// 1. Create a System Transfer instruction (for SOL) or Token Transfer (for SPL)
	// 2. Build a transaction with the instruction
	// 3. Serialize it to base64
	// 4. Return it for client-side signing

	// For now, return a JSON representation that frontend can use to build the transaction
	txJSON, err := json.Marshal(txData)
	if err != nil {
		return "", fmt.Errorf("failed to encode transaction data: %w", err)
	}

	logger.Info("Transfer transaction prepared",
		zap.String("from", fromAddress),
		zap.String("to", toAddress),
		zap.Uint64("amount", amount),
		zap.String("blockhash", blockhash.String()),
	)

	// Return base64 encoded transaction data
	// Note: Frontend should use @solana/web3.js to properly build and sign the transaction
	return string(txJSON), nil
}
