package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/contact"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// ContactHandler handles contact HTTP requests
type ContactHandler struct {
	contactService contact.Service
}

// NewContactHandler creates a new contact handler
func NewContactHandler(contactService contact.Service) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
	}
}

// AddContact handles POST /api/v1/contacts
func (h *ContactHandler) AddContact(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.AddContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.contactService.AddContact(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("Failed to add contact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to add contact",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// RemoveContact handles DELETE /api/v1/contacts/:contactId
func (h *ContactHandler) RemoveContact(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	contactID, err := uuid.Parse(c.Param("contactId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid contact ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.contactService.RemoveContact(c.Request.Context(), userID, contactID); err != nil {
		logger.Error("Failed to remove contact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to remove contact",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contact removed"})
}

// GetContacts handles GET /api/v1/contacts
func (h *ContactHandler) GetContacts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse pagination parameters
	var limit, offset int
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	contacts, err := h.contactService.GetContacts(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get contacts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get contacts",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}

// GetFavoriteContacts handles GET /api/v1/contacts/favorites
func (h *ContactHandler) GetFavoriteContacts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	contacts, err := h.contactService.GetFavoriteContacts(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get favorite contacts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get favorite contacts",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}

// SetFavorite handles PUT /api/v1/contacts/:contactId/favorite
func (h *ContactHandler) SetFavorite(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	contactID, err := uuid.Parse(c.Param("contactId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid contact ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req struct {
		Favorite bool `json:"favorite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.contactService.SetFavorite(c.Request.Context(), userID, contactID, req.Favorite); err != nil {
		logger.Error("Failed to set favorite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to set favorite",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "favorite updated"})
}

// SendInvite handles POST /api/v1/contacts/invites
func (h *ContactHandler) SendInvite(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	senderID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.SendInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.contactService.SendInvite(c.Request.Context(), senderID, &req)
	if err != nil {
		logger.Error("Failed to send invite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to send invite",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPendingInvites handles GET /api/v1/contacts/invites/pending
func (h *ContactHandler) GetPendingInvites(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	invites, err := h.contactService.GetPendingInvites(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get pending invites", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get pending invites",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}

// AcceptInvite handles POST /api/v1/contacts/invites/:inviteId/accept
func (h *ContactHandler) AcceptInvite(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	recipientID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	inviteID, err := uuid.Parse(c.Param("inviteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid invite ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.contactService.AcceptInvite(c.Request.Context(), recipientID, inviteID); err != nil {
		logger.Error("Failed to accept invite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to accept invite",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invite accepted"})
}

// RejectInvite handles POST /api/v1/contacts/invites/:inviteId/reject
func (h *ContactHandler) RejectInvite(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	recipientID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	inviteID, err := uuid.Parse(c.Param("inviteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid invite ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.contactService.RejectInvite(c.Request.Context(), recipientID, inviteID); err != nil {
		logger.Error("Failed to reject invite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to reject invite",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invite rejected"})
}

// SearchContacts handles GET /api/v1/contacts/search
func (h *ContactHandler) SearchContacts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "validation_error",
			Message: "Query parameter 'q' is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var limit int
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	} else {
		limit = 20
	}

	contacts, err := h.contactService.SearchContacts(c.Request.Context(), userID, query, limit)
	if err != nil {
		logger.Error("Failed to search contacts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to search contacts",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}
