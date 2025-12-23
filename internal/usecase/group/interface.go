package group

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the group use case interface
type Service interface {
	// CreateGroup creates a new group
	CreateGroup(ctx context.Context, userID uuid.UUID, req *dto.CreateGroupRequest) (*dto.CreateGroupResponse, error)

	// GetGroup gets group information
	GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.GetGroupResponse, error)

	// GetUserGroups gets all groups where user is a member
	GetUserGroups(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetGroupsResponse, error)

	// UpdateGroup updates group information
	UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req *dto.UpdateGroupRequest) error

	// UpdateGroupSettings updates group settings
	UpdateGroupSettings(ctx context.Context, userID, groupID uuid.UUID, req *dto.UpdateGroupSettingsRequest) error

	// DeleteGroup deletes a group
	DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error

	// AddMember adds a member to the group
	AddMember(ctx context.Context, userID, groupID uuid.UUID, req *dto.AddMemberRequest) error

	// RemoveMember removes a member from the group
	RemoveMember(ctx context.Context, userID, groupID, targetUserID uuid.UUID) error

	// UpdateMemberRole updates a member's role
	UpdateMemberRole(ctx context.Context, userID, groupID, targetUserID uuid.UUID, req *dto.UpdateMemberRoleRequest) error

	// LeaveGroup allows a user to leave a group
	LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error
}
