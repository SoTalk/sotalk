package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/delivery/http/request"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/internal/usecase/group"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// GroupHandler handles group HTTP requests
type GroupHandler struct {
	groupService group.Service
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(groupService group.Service) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// CreateGroup handles POST /api/v1/groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	result, err := h.groupService.CreateGroup(c.Request.Context(), userID, &dto.CreateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		MemberIDs:   req.MemberIDs,
	})

	if err != nil {
		logger.Error("Failed to create group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_group_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, response.CreateGroupResponse{
		Group: mapGroupDTO(result.Group),
	})
}

// GetGroup handles GET /api/v1/groups/:id
func (h *GroupHandler) GetGroup(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	result, err := h.groupService.GetGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		logger.Error("Failed to get group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_group_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, response.GetGroupResponse{
		Group:   mapGroupDTO(result.Group),
		Members: mapMemberDTOs(result.Members),
	})
}

// GetUserGroups handles GET /api/v1/groups
func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.GetGroupsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	result, err := h.groupService.GetUserGroups(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get user groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "get_groups_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	groups := make([]response.GroupDTO, len(result.Groups))
	for i, grp := range result.Groups {
		groups[i] = mapGroupDTO(grp)
	}

	c.JSON(http.StatusOK, response.GetGroupsResponse{
		Groups: groups,
	})
}

// UpdateGroup handles PUT /api/v1/groups/:id
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.UpdateGroup(c.Request.Context(), userID, groupID, &dto.UpdateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
	})

	if err != nil {
		logger.Error("Failed to update group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_group_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group updated"})
}

// DeleteGroup handles DELETE /api/v1/groups/:id
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.DeleteGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		logger.Error("Failed to delete group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_group_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

// AddMember handles POST /api/v1/groups/:id/members
func (h *GroupHandler) AddMember(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.AddMember(c.Request.Context(), userID, groupID, &dto.AddMemberRequest{
		UserID: req.UserID,
		Role:   req.Role,
	})

	if err != nil {
		logger.Error("Failed to add member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "add_member_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

// RemoveMember handles DELETE /api/v1/groups/:id/members/:userId
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get target user ID from URL
	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid target user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.RemoveMember(c.Request.Context(), userID, groupID, targetUserID)
	if err != nil {
		logger.Error("Failed to remove member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "remove_member_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// LeaveGroup handles POST /api/v1/groups/:id/leave
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.LeaveGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		logger.Error("Failed to leave group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "leave_group_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "left group"})
}

// UpdateMemberRole handles PUT /api/v1/groups/:id/members/:userId/role
func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get target user ID from URL
	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid target user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.UpdateMemberRole(c.Request.Context(), userID, groupID, targetUserID, &dto.UpdateMemberRoleRequest{
		Role: req.Role,
	})

	if err != nil {
		logger.Error("Failed to update member role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_role_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

// UpdateGroupSettings handles PUT /api/v1/groups/:id/settings
func (h *GroupHandler) UpdateGroupSettings(c *gin.Context) {
	// Get user ID from context
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
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get group ID from URL
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_group_id",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req request.UpdateGroupSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service
	err = h.groupService.UpdateGroupSettings(c.Request.Context(), userID, groupID, &dto.UpdateGroupSettingsRequest{
		WhoCanMessage:    req.WhoCanMessage,
		WhoCanAddMembers: req.WhoCanAddMembers,
		WhoCanEditInfo:   req.WhoCanEditInfo,
	})

	if err != nil {
		logger.Error("Failed to update group settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_settings_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group settings updated"})
}

// Helper functions to map DTOs
func mapGroupDTO(grp dto.GroupDTO) response.GroupDTO {
	var settings *response.GroupSettings
	if grp.Settings != nil {
		settings = &response.GroupSettings{
			WhoCanMessage:    grp.Settings.WhoCanMessage,
			WhoCanAddMembers: grp.Settings.WhoCanAddMembers,
			WhoCanEditInfo:   grp.Settings.WhoCanEditInfo,
		}
	}

	return response.GroupDTO{
		ID:             grp.ID,
		ConversationID: grp.ConversationID,
		Name:           grp.Name,
		Description:    grp.Description,
		Avatar:         grp.Avatar,
		CreatorID:      grp.CreatorID,
		MaxMembers:     grp.MaxMembers,
		MemberCount:    grp.MemberCount,
		Settings:       settings,
		CreatedAt:      grp.CreatedAt,
		UpdatedAt:      grp.UpdatedAt,
	}
}

func mapMemberDTOs(members []dto.MemberDTO) []response.MemberDTO {
	result := make([]response.MemberDTO, len(members))
	for i, m := range members {
		var userDTO *response.UserDTO
		if m.User != nil {
			userDTO = &response.UserDTO{
				ID:            m.User.ID,
				WalletAddress: m.User.WalletAddress,
				Username:      m.User.Username,
				Avatar:        m.User.Avatar,
				Status:        m.User.Status,
			}
		}

		var permissions *response.MemberPermissions
		if m.Permissions != nil {
			permissions = &response.MemberPermissions{
				CanSendMessages:  m.Permissions.CanSendMessages,
				CanAddMembers:    m.Permissions.CanAddMembers,
				CanRemoveMembers: m.Permissions.CanRemoveMembers,
				CanEditInfo:      m.Permissions.CanEditInfo,
			}
		}

		result[i] = response.MemberDTO{
			UserID:      m.UserID,
			User:        userDTO,
			Role:        m.Role,
			Permissions: permissions,
			JoinedAt:    m.JoinedAt,
		}
	}
	return result
}
