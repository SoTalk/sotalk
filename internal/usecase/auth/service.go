package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/domain/wallet"
	"github.com/yourusername/sotalk/internal/infrastructure/solana"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/middleware"
	"github.com/yourusername/sotalk/pkg/redis"
	solanaUtil "github.com/yourusername/sotalk/pkg/solana"
)

// ReferralService is the interface for referral operations
type ReferralService interface {
	ApplyReferralCode(ctx context.Context, refereeID uuid.UUID, code string) error
	CompleteReferral(ctx context.Context, refereeID uuid.UUID) error
}

// service implements the Service interface
type service struct {
	userRepo        user.Repository
	walletRepo      wallet.Repository
	jwtManager      *middleware.JWTManager
	solanaClient    *solana.Client
	redisClient     *redis.Client
	referralService ReferralService
}

// NewService creates a new authentication service
func NewService(
	userRepo user.Repository,
	walletRepo wallet.Repository,
	jwtManager *middleware.JWTManager,
	solanaClient *solana.Client,
	redisClient *redis.Client,
	referralService ReferralService,
) Service {
	return &service{
		userRepo:        userRepo,
		walletRepo:      walletRepo,
		jwtManager:      jwtManager,
		solanaClient:    solanaClient,
		redisClient:     redisClient,
		referralService: referralService,
	}
}

// GenerateWallet generates a new Solana wallet with mnemonic phrase
func (s *service) GenerateWallet(ctx context.Context, req *dto.GenerateWalletRequest) (*dto.GenerateWalletResponse, error) {
	// Generate new mnemonic
	mnemonic, err := solanaUtil.GenerateMnemonic()
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Derive wallet address from mnemonic
	walletAddress, err := solanaUtil.GetPublicKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to derive wallet: %w", err)
	}

	// Check if wallet already exists
	existingUser, err := s.userRepo.FindByWalletAddress(ctx, walletAddress)
	if err != nil && err != user.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("wallet address already exists")
	}

	// Check if username already exists
	existingUsername, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil && err != user.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}

	if existingUsername != nil {
		return nil, fmt.Errorf("username already taken")
	}

	// Create new user
	userEntity := user.NewUser(walletAddress, req.Username, walletAddress)
	userEntity.UpdateStatus(user.StatusOnline)

	if err := s.userRepo.Create(ctx, userEntity); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create wallet entry
	walletEntity := wallet.NewWallet(userEntity.ID, walletAddress, "Default Wallet")
	walletEntity.SetDefault()
	if err := s.walletRepo.CreateWallet(ctx, walletEntity); err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Apply referral code if provided
	if req.ReferralCode != nil && *req.ReferralCode != "" {
		if err := s.referralService.ApplyReferralCode(ctx, userEntity.ID, *req.ReferralCode); err != nil {
			// Log error but don't fail registration
			fmt.Printf("Warning: failed to apply referral code: %v\n", err)
		}
	}

	// Generate access token
	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(userEntity.ID, userEntity.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(userEntity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Map to DTO
	userDTO := dto.UserDTO{
		ID:            userEntity.ID.String(),
		WalletAddress: userEntity.WalletAddress,
		Username:      userEntity.Username,
		Avatar:        userEntity.Avatar,
		PublicKey:     userEntity.PublicKey,
		Status:        string(userEntity.Status),
		LastSeen:      userEntity.LastSeen,
		CreatedAt:     userEntity.CreatedAt,
		UpdatedAt:     userEntity.UpdatedAt,
	}

	return &dto.GenerateWalletResponse{
		Mnemonic:      mnemonic,
		WalletAddress: walletAddress,
		PublicKey:     walletAddress,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresAt:     accessExpiresAt,
		User:          userDTO,
	}, nil
}

// RefreshToken refreshes the access token using refresh token
func (s *service) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	// Validate refresh token
	userID, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user from database
	userEntity, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new access token
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(userEntity.ID, userEntity.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

// GetCurrentUser gets the current user from token
func (s *service) GetCurrentUser(ctx context.Context, userID string) (*dto.UserDTO, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user from database
	userEntity, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Map to DTO
	return &dto.UserDTO{
		ID:            userEntity.ID.String(),
		WalletAddress: userEntity.WalletAddress,
		Username:      userEntity.Username,
		Avatar:        userEntity.Avatar,
		PublicKey:     userEntity.PublicKey,
		Status:        string(userEntity.Status),
		LastSeen:      userEntity.LastSeen,
		CreatedAt:     userEntity.CreatedAt,
		UpdatedAt:     userEntity.UpdatedAt,
	}, nil
}

// Logout invalidates the current session
func (s *service) Logout(ctx context.Context, userID string, token string) error {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Update user status to offline
	if err := s.userRepo.UpdateStatus(ctx, uid, user.StatusOffline); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Note: Token blacklisting would be implemented here with Redis
	// For now, we rely on short-lived tokens and client-side token removal
	// Example: s.redisClient.Blacklist(ctx, token, tokenExpiry)

	return nil
}

// DeleteAccount permanently deletes the user account
func (s *service) DeleteAccount(ctx context.Context, userID string) error {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if user exists
	userEntity, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Delete user's wallets
	// Note: This assumes cascade delete is set up in the database
	// Or we need to explicitly delete related entities
	if err := s.userRepo.Delete(ctx, userEntity.ID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GenerateChallenge generates a challenge for wallet to sign
func (s *service) GenerateChallenge(ctx context.Context, walletAddress string) (*dto.ChallengeResponse, error) {
	// Generate random challenge message
	challengeID := uuid.New().String()
	challenge := fmt.Sprintf("Sign this message to authenticate with SoTalk.\n\nChallenge ID: %s\n\nThis request will not trigger a blockchain transaction or cost any gas fees.", challengeID)

	expiresAt := time.Now().Add(5 * time.Minute)
	expiresAtStr := fmt.Sprintf("%d", expiresAt.Unix())

	// Include expiration time in the message itself
	challengeWithExpiry := fmt.Sprintf("%s\nExpires At: %s", challenge, expiresAtStr)

	// Store challenge in Redis with 5-minute expiration
	key := fmt.Sprintf("challenge:%s", walletAddress)
	if err := s.redisClient.Set(ctx, key, challengeWithExpiry, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to store challenge in cache: %w", err)
	}

	return &dto.ChallengeResponse{
		Challenge: challengeWithExpiry,
		ExpiresAt: expiresAt,
	}, nil
}

// VerifySignature verifies the signed challenge and authenticates user
// Note: Also supports wallet-only authentication (when signature is empty)
func (s *service) VerifySignature(ctx context.Context, req *dto.VerifySignatureRequest) (*dto.VerifySignatureResponse, error) {
	// Check if this is wallet-only authentication (no signature)
	if req.Signature == "" || req.Message == "" {
		fmt.Println("ℹ️ Using wallet-only authentication (no signature verification)")
		// Skip signature verification for wallet-only auth
		// Connection to Phantom alone proves wallet ownership
	} else {
		// Full signature verification flow
		fmt.Println("ℹ️ Using signature-based authentication")

		// Validate challenge from Redis
		key := fmt.Sprintf("challenge:%s", req.WalletAddress)
		storedChallenge, err := s.redisClient.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("challenge not found or expired")
		}

		// Verify that the message matches the stored challenge
		if storedChallenge != req.Message {
			return nil, fmt.Errorf("challenge mismatch")
		}

		// Verify the signature using Solana ed25519 verification
		isValid, err := solanaUtil.VerifySignature(req.WalletAddress, req.Message, req.Signature)
		if err != nil {
			return nil, fmt.Errorf("failed to verify signature: %w", err)
		}

		if !isValid {
			return nil, fmt.Errorf("invalid signature")
		}

		// Delete challenge after successful verification (one-time use)
		if err := s.redisClient.Del(ctx, key); err != nil {
			// Log error but don't fail the authentication
			fmt.Printf("Warning: failed to delete challenge from cache: %v\n", err)
		}
	}

	// Find or create user
	userEntity, err := s.userRepo.FindByWalletAddress(ctx, req.WalletAddress)
	isNewUser := false
	if err != nil {
		if err == user.ErrUserNotFound {
			// Auto-register new user
			isNewUser = true
			username := "User" + req.WalletAddress[:6]
			if req.Username != nil && *req.Username != "" {
				username = *req.Username
			}

			userEntity = user.NewUser(req.WalletAddress, username, req.WalletAddress)
			userEntity.UpdateStatus(user.StatusOnline)

			if err := s.userRepo.Create(ctx, userEntity); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			// Apply referral code if provided
			if req.ReferralCode != nil && *req.ReferralCode != "" {
				if err := s.referralService.ApplyReferralCode(ctx, userEntity.ID, *req.ReferralCode); err != nil {
					// Log error but don't fail registration
					fmt.Printf("Warning: failed to apply referral code: %v\n", err)
				}
			}

			// Create default wallet entry
			walletEntity := wallet.NewWallet(userEntity.ID, req.WalletAddress, "Default Wallet")
			walletEntity.SetDefault()

			// Fetch initial balance from Solana blockchain
			balance, err := s.solanaClient.GetBalance(ctx, req.WalletAddress)
			if err == nil {
				walletEntity.UpdateBalance(balance)
			}

			// Get token balances
			tokenAccounts, err := s.solanaClient.GetTokenAccounts(ctx, req.WalletAddress)
			if err == nil {
				for _, ta := range tokenAccounts {
					walletEntity.SetTokenBalance(ta.Mint, wallet.TokenBalance{
						Mint:     ta.Mint,
						Decimals: ta.Decimals,
						UiAmount: ta.UiAmount,
						Symbol:   ta.Symbol,
					})
				}
			}

			if err := s.walletRepo.CreateWallet(ctx, walletEntity); err != nil {
				return nil, fmt.Errorf("failed to create wallet: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to find user: %w", err)
		}
	} else {
		// Update existing user status to online
		userEntity.UpdateStatus(user.StatusOnline)
		if err := s.userRepo.Update(ctx, userEntity); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Generate access token
	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(userEntity.ID, userEntity.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(userEntity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Map to DTO
	userDTO := dto.UserDTO{
		ID:            userEntity.ID.String(),
		WalletAddress: userEntity.WalletAddress,
		Username:      userEntity.Username,
		Avatar:        userEntity.Avatar,
		PublicKey:     userEntity.PublicKey,
		Status:        string(userEntity.Status),
		LastSeen:      userEntity.LastSeen,
		CreatedAt:     userEntity.CreatedAt,
		UpdatedAt:     userEntity.UpdatedAt,
	}

	// Complete referral if this was a new user with a referral
	if isNewUser {
		if err := s.referralService.CompleteReferral(ctx, userEntity.ID); err != nil {
			// Log error but don't fail authentication
			fmt.Printf("Warning: failed to complete referral: %v\n", err)
		}
	}

	return &dto.VerifySignatureResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
		User:         userDTO,
	}, nil
}
