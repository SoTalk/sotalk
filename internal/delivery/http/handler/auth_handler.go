package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/auth"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// GenerateWallet handles POST /api/v1/auth/register
// @Summary Generate new Solana wallet
// @Description Generates a new Solana wallet with mnemonic phrase (shown ONCE!)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.GenerateWalletRequest true "Generate Wallet Request"
// @Success 200 {object} response.GenerateWalletResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) GenerateWallet(c *gin.Context) {
	var req request.GenerateWalletRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.authService.GenerateWallet(c.Request.Context(), &dto.GenerateWalletRequest{
		Username:     req.Username,
		ReferralCode: req.ReferralCode,
	})

	if err != nil {
		logger.Error("Failed to generate wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "wallet_generation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.GenerateWalletResponse{
		Mnemonic:      result.Mnemonic,
		WalletAddress: result.WalletAddress,
		PublicKey:     result.PublicKey,
		AccessToken:   result.AccessToken,
		RefreshToken:  result.RefreshToken,
		ExpiresAt:     result.ExpiresAt,
		User: response.UserDTO{
			ID:            result.User.ID,
			WalletAddress: result.User.WalletAddress,
			Username:      result.User.Username,
			Avatar:        result.User.Avatar,
			Status:        result.User.Status,
		},
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
// @Summary Refresh access token
// @Description Refreshes the access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "Refresh Request"
// @Success 200 {object} response.RefreshTokenResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.authService.RefreshToken(c.Request.Context(), &dto.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		logger.Error("Failed to refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "refresh_failed",
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, response.RefreshTokenResponse{
		AccessToken: result.AccessToken,
		ExpiresAt:   result.ExpiresAt,
	})
}

// GetMe handles GET /api/v1/auth/me
// @Summary Get current user
// @Description Gets the current authenticated user's information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.UserDTO
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Call use case
	result, err := h.authService.GetCurrentUser(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get current user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_user_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.UserDTO{
		ID:            result.ID,
		WalletAddress: result.WalletAddress,
		Username:      result.Username,
		Avatar:        result.Avatar,
		Status:        result.Status,
	})
}

// Logout handles POST /api/v1/auth/logout
// @Summary Logout
// @Description Logs out the current user and invalidates the session
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Get token from Authorization header
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Call use case
	if err := h.authService.Logout(c.Request.Context(), userID.(string), token); err != nil {
		logger.Error("Failed to logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "logout_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Logged out successfully",
	})
}

// DeleteAccount handles DELETE /api/v1/auth/account
// @Summary Delete account
// @Description Permanently deletes the user account and all associated data
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/account [delete]
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Call use case
	if err := h.authService.DeleteAccount(c.Request.Context(), userID.(string)); err != nil {
		logger.Error("Failed to delete account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_account_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Account deleted successfully",
	})
}

// GenerateChallenge handles POST /api/v1/auth/challenge
// @Summary Generate authentication challenge
// @Description Generates a challenge message for the wallet to sign
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ChallengeRequest true "Challenge Request"
// @Success 200 {object} response.ChallengeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/challenge [post]
func (h *AuthHandler) GenerateChallenge(c *gin.Context) {
	var req request.ChallengeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.authService.GenerateChallenge(c.Request.Context(), req.WalletAddress)
	if err != nil {
		logger.Error("Failed to generate challenge", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "challenge_generation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.ChallengeResponse{
		Challenge: result.Challenge,
		ExpiresAt: result.ExpiresAt,
	})
}

// VerifySignature handles POST /api/v1/auth/verify
// @Summary Verify signed challenge
// @Description Verifies the signed challenge and authenticates the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.VerifySignatureRequest true "Verify Signature Request"
// @Success 200 {object} response.VerifySignatureResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/verify [post]
func (h *AuthHandler) VerifySignature(c *gin.Context) {
	var req request.VerifySignatureRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.authService.VerifySignature(c.Request.Context(), &dto.VerifySignatureRequest{
		WalletAddress: req.WalletAddress,
		Signature:     req.Signature,
		Message:       req.Message,
		Username:      req.Username,
		ReferralCode:  req.ReferralCode,
	})

	if err != nil {
		logger.Error("Failed to verify signature", zap.Error(err))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, response.VerifySignatureResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		User: response.UserDTO{
			ID:            result.User.ID,
			WalletAddress: result.User.WalletAddress,
			Username:      result.User.Username,
			Avatar:        result.User.Avatar,
			Status:        result.User.Status,
		},
	})
}
