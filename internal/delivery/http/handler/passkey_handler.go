package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/passkey"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// PasskeyHandler handles passkey HTTP requests
type PasskeyHandler struct {
	passkeyService passkey.Service
}

// NewPasskeyHandler creates a new passkey handler
func NewPasskeyHandler(passkeyService passkey.Service) *PasskeyHandler {
	return &PasskeyHandler{
		passkeyService: passkeyService,
	}
}

// RegisterBegin handles POST /api/v1/passkey/register/begin
// @Summary Begin passkey registration
// @Description Starts the passkey registration process and returns options
// @Tags passkey
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.RegisterPasskeyBeginResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/passkey/register/begin [post]
func (h *PasskeyHandler) RegisterBegin(c *gin.Context) {
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

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.passkeyService.BeginRegistration(c.Request.Context(), &dto.RegisterPasskeyBeginRequest{
		UserID: userUUID,
	})

	if err != nil {
		logger.Error("Failed to begin passkey registration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "registration_begin_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.RegisterPasskeyBeginResponse{
		Options: result.Options,
	})
}

// RegisterFinish handles POST /api/v1/passkey/register/finish
// @Summary Finish passkey registration
// @Description Completes the passkey registration process
// @Tags passkey
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.RegisterPasskeyFinishRequest true "Registration Credential"
// @Success 200 {object} response.RegisterPasskeyFinishResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/passkey/register/finish [post]
func (h *PasskeyHandler) RegisterFinish(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.RegisterPasskeyFinishRequest
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
	result, err := h.passkeyService.FinishRegistration(c.Request.Context(), &dto.RegisterPasskeyFinishRequest{
		UserID:     userUUID,
		Credential: req.Credential,
	})

	if err != nil {
		logger.Error("Failed to finish passkey registration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "registration_finish_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.RegisterPasskeyFinishResponse{
		CredentialID: result.CredentialID,
		CreatedAt:    result.CreatedAt,
		Message:      "Passkey registered successfully",
	})
}

// AuthenticateBegin handles POST /api/v1/passkey/authenticate/begin
// @Summary Begin passkey authentication
// @Description Starts the passkey authentication process and returns challenge
// @Tags passkey
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.AuthenticatePasskeyBeginResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/passkey/authenticate/begin [post]
func (h *PasskeyHandler) AuthenticateBegin(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.passkeyService.BeginAuthentication(c.Request.Context(), &dto.AuthenticatePasskeyBeginRequest{
		UserID: userUUID,
	})

	if err != nil {
		logger.Error("Failed to begin passkey authentication", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "authentication_begin_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.AuthenticatePasskeyBeginResponse{
		Options: result.Options,
	})
}

// AuthenticateFinish handles POST /api/v1/passkey/authenticate/finish
// @Summary Finish passkey authentication
// @Description Completes the passkey authentication process
// @Tags passkey
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.AuthenticatePasskeyFinishRequest true "Authentication Credential"
// @Success 200 {object} response.AuthenticatePasskeyFinishResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/passkey/authenticate/finish [post]
func (h *PasskeyHandler) AuthenticateFinish(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.AuthenticatePasskeyFinishRequest
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
	result, err := h.passkeyService.FinishAuthentication(c.Request.Context(), &dto.AuthenticatePasskeyFinishRequest{
		UserID:     userUUID,
		Credential: req.Credential,
	})

	if err != nil {
		logger.Error("Failed to finish passkey authentication", zap.Error(err))
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "authentication_failed",
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, response.AuthenticatePasskeyFinishResponse{
		Success: result.Success,
		Message: result.Message,
		UsedAt:  result.UsedAt,
	})
}

// GetPasskeys handles GET /api/v1/passkey
// @Summary Get user passkeys
// @Description Retrieves all passkeys registered for the current user
// @Tags passkey
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.GetPasskeysResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/passkey [get]
func (h *PasskeyHandler) GetPasskeys(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.passkeyService.GetUserPasskeys(c.Request.Context(), &dto.GetPasskeysRequest{
		UserID: userUUID,
	})

	if err != nil {
		logger.Error("Failed to get passkeys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_passkeys_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to response DTOs
	passkeys := make([]response.PasskeyDTO, len(result.Passkeys))
	for i, pk := range result.Passkeys {
		transports := make([]string, len(pk.Transports))
		for j, t := range pk.Transports {
			transports[j] = string(t)
		}

		passkeys[i] = response.PasskeyDTO{
			ID:              pk.ID,
			CredentialID:    pk.CredentialID,
			AttestationType: pk.AttestationType,
			Transports:      transports,
			BackupEligible:  pk.BackupEligible,
			BackupState:     pk.BackupState,
			CreatedAt:       pk.CreatedAt,
			LastUsedAt:      pk.LastUsedAt,
		}
	}

	c.JSON(http.StatusOK, response.GetPasskeysResponse{
		Passkeys: passkeys,
		Count:    len(passkeys),
	})
}

// DeletePasskey handles DELETE /api/v1/passkey/:credentialId
// @Summary Delete a passkey
// @Description Deletes a specific passkey credential
// @Tags passkey
// @Security BearerAuth
// @Produce json
// @Param credentialId path string true "Credential ID"
// @Success 200 {object} response.DeletePasskeyResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/passkey/{credentialId} [delete]
func (h *PasskeyHandler) DeletePasskey(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	credentialID := c.Param("credentialId")
	if credentialID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Credential ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call use case
	result, err := h.passkeyService.DeletePasskey(c.Request.Context(), &dto.DeletePasskeyRequest{
		UserID:       userUUID,
		CredentialID: credentialID,
	})

	if err != nil {
		logger.Error("Failed to delete passkey", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_passkey_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.DeletePasskeyResponse{
		Message: result.Message,
	})
}
