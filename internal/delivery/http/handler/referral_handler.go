package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/referral"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// ReferralHandler handles referral-related HTTP requests
type ReferralHandler struct {
	referralService referral.Service
}

// NewReferralHandler creates a new referral handler
func NewReferralHandler(referralService referral.Service) *ReferralHandler {
	return &ReferralHandler{
		referralService: referralService,
	}
}

// GetMyReferralCode handles GET /api/v1/referrals/my-code
// @Summary Get my referral code
// @Description Get the current user's referral code
// @Tags referrals
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.ReferralCodeResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/referrals/my-code [get]
func (h *ReferralHandler) GetMyReferralCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	result, err := h.referralService.GetMyReferralCode(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get referral code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_referral_code_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMyReferralStats handles GET /api/v1/referrals/stats
// @Summary Get my referral statistics
// @Description Get the current user's referral statistics
// @Tags referrals
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.ReferralStatsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/referrals/stats [get]
func (h *ReferralHandler) GetMyReferralStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	result, err := h.referralService.GetMyReferralStats(c.Request.Context(), userID.(string))
	if err != nil {
		logger.Error("Failed to get referral stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_referral_stats_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMyReferralHistory handles GET /api/v1/referrals/history
// @Summary Get my referral history
// @Description Get the current user's referral history
// @Tags referrals
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.ReferralHistoryResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/referrals/history [get]
func (h *ReferralHandler) GetMyReferralHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.referralService.GetMyReferralHistory(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		logger.Error("Failed to get referral history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_referral_history_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ValidateReferralCode handles POST /api/v1/referrals/validate
// @Summary Validate referral code
// @Description Validate if a referral code exists and is valid
// @Tags referrals
// @Accept json
// @Produce json
// @Param body body dto.ApplyReferralCodeRequest true "Referral code"
// @Success 200 {object} dto.ValidateReferralCodeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/referrals/validate [post]
func (h *ReferralHandler) ValidateReferralCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.referralService.ValidateReferralCode(c.Request.Context(), req.Code)
	if err != nil {
		logger.Error("Failed to validate referral code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "validate_referral_code_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
