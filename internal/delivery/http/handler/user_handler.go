package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/user"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	userService user.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile handles GET /api/v1/users/profile
// @Summary Get user profile
// @Description Gets user profile information
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.UserProfileResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	result, err := h.userService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_profile_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.UserProfileResponse{
		ID:            result.ID,
		WalletAddress: result.WalletAddress,
		Username:      result.Username,
		Avatar:        result.Avatar,
		Bio:           result.Bio,
		Status:        result.Status,
		CreatedAt:     result.CreatedAt,
	})
}

// UpdateProfile handles PUT /api/v1/users/profile
// @Summary Update user profile
// @Description Updates user profile information
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdateProfileRequest true "Update Profile Request"
// @Success 200 {object} response.UserProfileResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.userService.UpdateProfile(c.Request.Context(), &dto.UpdateProfileRequest{
		UserID:   userID.(string),
		Username: req.Username,
		Avatar:   req.Avatar,
		Bio:      req.Bio,
	})

	if err != nil {
		logger.Error("Failed to update profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_profile_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.UserProfileResponse{
		ID:            result.ID,
		WalletAddress: result.WalletAddress,
		Username:      result.Username,
		Avatar:        result.Avatar,
		Bio:           result.Bio,
		Status:        result.Status,
		CreatedAt:     result.CreatedAt,
	})
}

// GetPreferences handles GET /api/v1/users/preferences
// @Summary Get user preferences
// @Description Gets user preferences (language, theme, etc.)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.UserPreferencesResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/preferences [get]
func (h *UserHandler) GetPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	result, err := h.userService.GetPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get preferences", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_preferences_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.UserPreferencesResponse{
		Language: result.Language,
		Theme:    result.Theme,
	})
}

// UpdatePreferences handles PUT /api/v1/users/preferences
// @Summary Update user preferences
// @Description Updates user preferences (language, theme, etc.)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdatePreferencesRequest true "Update Preferences Request"
// @Success 200 {object} response.UserPreferencesResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/preferences [put]
func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req request.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.userService.UpdatePreferences(c.Request.Context(), &dto.UpdatePreferencesRequest{
		UserID:   userID.(string),
		Language: req.Language,
		Theme:    req.Theme,
	})

	if err != nil {
		logger.Error("Failed to update preferences", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_preferences_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.UserPreferencesResponse{
		Language: result.Language,
		Theme:    result.Theme,
	})
}

// GetUserByID handles GET /api/v1/users/:id
// @Summary Get user by ID
// @Description Gets public user information by user ID
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.PublicUserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "user_not_found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, response.PublicUserResponse{
		ID:            result.ID,
		WalletAddress: result.WalletAddress,
		Username:      result.Username,
		Avatar:        result.Avatar,
		Bio:           result.Bio,
		Status:        result.Status,
	})
}

// CheckWalletExists handles POST /api/v1/users/check-wallet
// @Summary Check if wallet address exists
// @Description Checks if a wallet address is registered in the system
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CheckWalletExistsRequest true "Check Wallet Request"
// @Success 200 {object} response.CheckWalletExistsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/users/check-wallet [post]
func (h *UserHandler) CheckWalletExists(c *gin.Context) {
	var req request.CheckWalletExistsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	exists, user, err := h.userService.CheckWalletExists(c.Request.Context(), req.WalletAddress)
	if err != nil {
		logger.Error("Failed to check wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "check_wallet_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var userResponse *response.PublicUserResponse
	if exists && user != nil {
		userResponse = &response.PublicUserResponse{
			ID:            user.ID,
			WalletAddress: user.WalletAddress,
			Username:      user.Username,
			Avatar:        user.Avatar,
			Bio:           user.Bio,
			Status:        user.Status,
		}
	}

	c.JSON(http.StatusOK, response.CheckWalletExistsResponse{
		Exists: exists,
		User:   userResponse,
	})
}

// SendInvitation handles POST /api/v1/invitations
// @Summary Send invitation to wallet address
// @Description Sends an email invitation to a wallet address that is not registered
// @Tags invitations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.SendInvitationRequest true "Send Invitation Request"
// @Success 200 {object} response.SendInvitationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/invitations [post]
func (h *UserHandler) SendInvitation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req request.SendInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check if wallet already exists
	exists_wallet, _, _ := h.userService.CheckWalletExists(c.Request.Context(), req.WalletAddress)
	if exists_wallet {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "wallet_already_exists",
			Message: "This wallet address is already registered",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err := h.userService.SendInvitation(c.Request.Context(), &dto.SendInvitationRequest{
		SenderID:   userID.(string),
		Email:      req.Email,
		InviteLink: "", // Will be generated from sender's referral code
	})

	if err != nil {
		logger.Error("Failed to send invitation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "send_invitation_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.SendInvitationResponse{
		Message: "Invitation sent successfully",
		Sent:    true,
	})
}
