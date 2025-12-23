package solana

import (
	"crypto/ed25519"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic generates a new 12-word BIP39 mnemonic
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128) // 128 bits = 12 words
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// DeriveWalletFromMnemonic derives a Solana wallet from a BIP39 mnemonic
func DeriveWalletFromMnemonic(mnemonic string) (solana.PrivateKey, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return solana.PrivateKey{}, fmt.Errorf("invalid mnemonic phrase")
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "") // No passphrase

	// Use first 32 bytes as seed for ED25519 keypair
	if len(seed) < 32 {
		return solana.PrivateKey{}, fmt.Errorf("seed too short")
	}

	// Generate ED25519 keypair from seed
	// ED25519 private key is 64 bytes: 32-byte seed + 32-byte public key
	privateKeyED25519 := ed25519.NewKeyFromSeed(seed[:32])

	// Convert to solana.PrivateKey (which is []byte)
	privateKey := solana.PrivateKey(privateKeyED25519)

	return privateKey, nil
}

// GetPublicKeyFromMnemonic gets the Solana public key (wallet address) from mnemonic
func GetPublicKeyFromMnemonic(mnemonic string) (string, error) {
	privateKey, err := DeriveWalletFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}

	// Get public key from private key
	publicKey := privateKey.PublicKey()

	return publicKey.String(), nil
}

// VerifySignature verifies an ed25519 signature against a public key and message
// This is used for wallet authentication - the wallet signs a challenge message
// and we verify the signature using the wallet's public key
func VerifySignature(publicKeyBase58, message, signatureBase58 string) (bool, error) {
	// Decode public key from base58
	publicKeyBytes, err := base58.Decode(publicKeyBase58)
	if err != nil {
		return false, fmt.Errorf("invalid public key encoding: %w", err)
	}

	// Validate public key size (ed25519 public keys are 32 bytes)
	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid public key size: expected %d bytes, got %d", ed25519.PublicKeySize, len(publicKeyBytes))
	}

	// Decode signature from base58
	signatureBytes, err := base58.Decode(signatureBase58)
	if err != nil {
		return false, fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Validate signature size (ed25519 signatures are 64 bytes)
	if len(signatureBytes) != ed25519.SignatureSize {
		return false, fmt.Errorf("invalid signature size: expected %d bytes, got %d", ed25519.SignatureSize, len(signatureBytes))
	}

	// Convert message to bytes
	messageBytes := []byte(message)

	// Verify the signature
	isValid := ed25519.Verify(ed25519.PublicKey(publicKeyBytes), messageBytes, signatureBytes)

	return isValid, nil
}

// SignMessage signs a message with a private key (for testing purposes only)
// WARNING: In production, signing should ONLY happen on the client side with user's wallet
func SignMessage(privateKey []byte, message string) (string, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid private key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	messageBytes := []byte(message)
	signature := ed25519.Sign(ed25519.PrivateKey(privateKey), messageBytes)

	return base58.Encode(signature), nil
}
