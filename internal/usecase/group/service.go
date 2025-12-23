package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/conversation"
	"github.com/yourusername/sotalk/internal/domain/group"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Broadcaster defines methods for WebSocket broadcasting
type Broadcaster interface {
	BroadcastGroupCreated(ctx context.Context, conversationID uuid.UUID, group dto.GroupDTO) error
	BroadcastGroupUpdated(ctx context.Context, conversationID uuid.UUID, group dto.GroupDTO) error
	BroadcastGroupDeleted(ctx context.Context, conversationID uuid.UUID, groupID, deletedBy string) error
	BroadcastGroupSettingsUpdated(ctx context.Context, conversationID uuid.UUID, groupID, updatedBy string, settings dto.GroupSettings) error
	BroadcastGroupMemberJoined(ctx context.Context, conversationID uuid.UUID, groupID, userID, username, role, addedBy string) error
	BroadcastGroupMemberLeft(ctx context.Context, conversationID uuid.UUID, groupID, userID, username string) error
	BroadcastGroupMemberRemoved(ctx context.Context, conversationID uuid.UUID, groupID, userID, username, removedBy string) error
	BroadcastGroupMemberRoleChanged(ctx context.Context, conversationID uuid.UUID, groupID, userID, username, previousRole, newRole, changedBy string) error
}

type service struct {
	groupRepo        group.Repository
	conversationRepo conversation.Repository
	userRepo         user.Repository
	broadcaster      Broadcaster
}

// NewService creates a new group service
func NewService(
	groupRepo group.Repository,
	conversationRepo conversation.Repository,
	userRepo user.Repository,
	broadcaster Broadcaster,
) Service {
	return &service{
		groupRepo:        groupRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		broadcaster:      broadcaster,
	}
}

// CreateGroup creates a new group
func (s *service) CreateGroup(ctx context.Context, userID uuid.UUID, req *dto.CreateGroupRequest) (*dto.CreateGroupResponse, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("creator not found: %w", err)
	}

	// Create conversation for the group
	conv := conversation.NewGroupConversation()
	if err := s.conversationRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Create group
	grp := group.NewGroup(req.Name, req.Description, userID, conv.ID)
	if err := s.groupRepo.Create(ctx, grp); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Add creator as admin
	creatorMember := group.NewMember(grp.ID, userID, group.RoleAdmin)
	if err := s.groupRepo.AddMember(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	// Add creator to conversation
	participant := conversation.NewParticipant(conv.ID, userID, conversation.RoleAdmin)
	if err := s.conversationRepo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add creator to conversation: %w", err)
	}

	// Add initial members if provided
	for _, memberIDStr := range req.MemberIDs {
		memberID, err := uuid.Parse(memberIDStr)
		if err != nil {
			continue
		}

		// Skip if member is the creator
		if memberID == userID {
			continue
		}

		// Validate member exists
		if _, err := s.userRepo.FindByID(ctx, memberID); err != nil {
			continue
		}

		// Add member to group
		member := group.NewMember(grp.ID, memberID, group.RoleMember)
		if err := s.groupRepo.AddMember(ctx, member); err != nil {
			continue
		}

		// Add to conversation
		convParticipant := conversation.NewParticipant(conv.ID, memberID, conversation.RoleMember)
		s.conversationRepo.AddParticipant(ctx, convParticipant)
	}

	// Get member count
	memberCount, _ := s.groupRepo.CountMembers(ctx, grp.ID)

	groupDTO := toGroupDTO(grp, memberCount)

	// Broadcast group created event
	if s.broadcaster != nil {
		go s.broadcaster.BroadcastGroupCreated(context.Background(), conv.ID, groupDTO)
	}

	return &dto.CreateGroupResponse{
		Group: groupDTO,
	}, nil
}

// GetGroup gets group information
func (s *service) GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.GetGroupResponse, error) {
	// Check if user is a member
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, group.ErrNotMember
	}

	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get members
	members, err := s.groupRepo.FindMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	// Get member count
	memberCount := len(members)

	// Convert to DTOs
	memberDTOs := make([]dto.MemberDTO, len(members))
	for i, member := range members {
		u, err := s.userRepo.FindByID(ctx, member.UserID)
		var userDTO *dto.UserDTO
		if err == nil {
			userDTO = &dto.UserDTO{
				ID:            u.ID.String(),
				WalletAddress: u.WalletAddress,
				Username:      u.Username,
				Avatar:        u.Avatar,
				Status:        string(u.Status),
			}
		}

		memberDTOs[i] = dto.MemberDTO{
			UserID: member.UserID.String(),
			User:   userDTO,
			Role:   string(member.Role),
			Permissions: &dto.MemberPermissions{
				CanSendMessages:  member.Permissions.CanSendMessages,
				CanAddMembers:    member.Permissions.CanAddMembers,
				CanRemoveMembers: member.Permissions.CanRemoveMembers,
				CanEditInfo:      member.Permissions.CanEditInfo,
			},
			JoinedAt: member.JoinedAt,
		}
	}

	return &dto.GetGroupResponse{
		Group:   toGroupDTO(grp, memberCount),
		Members: memberDTOs,
	}, nil
}

// GetUserGroups gets all groups where user is a member
func (s *service) GetUserGroups(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.GetGroupsResponse, error) {
	// Set defaults
	if limit == 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	groups, err := s.groupRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	groupDTOs := make([]dto.GroupDTO, len(groups))
	for i, grp := range groups {
		memberCount, _ := s.groupRepo.CountMembers(ctx, grp.ID)
		groupDTOs[i] = toGroupDTO(grp, memberCount)
	}

	return &dto.GetGroupsResponse{
		Groups: groupDTOs,
	}, nil
}

// UpdateGroup updates group information
func (s *service) UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req *dto.UpdateGroupRequest) error {
	// Check if user is a member
	member, err := s.groupRepo.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	// Check permissions
	if !member.Permissions.CanEditInfo {
		return group.ErrNotAuthorized
	}

	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Update group info
	grp.UpdateInfo(req.Name, req.Description, req.Avatar)

	// Save
	if err := s.groupRepo.Update(ctx, grp); err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	// Broadcast group updated event
	if s.broadcaster != nil {
		memberCount, _ := s.groupRepo.CountMembers(ctx, groupID)
		groupDTO := toGroupDTO(grp, memberCount)
		go s.broadcaster.BroadcastGroupUpdated(context.Background(), grp.ConversationID, groupDTO)
	}

	return nil
}

// UpdateGroupSettings updates group settings
func (s *service) UpdateGroupSettings(ctx context.Context, userID, groupID uuid.UUID, req *dto.UpdateGroupSettingsRequest) error {
	// Check if user is a member
	member, err := s.groupRepo.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	// Only admins can update group settings
	if !member.IsAdmin() {
		return group.ErrNotAdmin
	}

	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Update group settings
	settings := &group.Settings{
		WhoCanMessage:    req.WhoCanMessage,
		WhoCanAddMembers: req.WhoCanAddMembers,
		WhoCanEditInfo:   req.WhoCanEditInfo,
	}
	grp.UpdateSettings(settings)

	// Save
	if err := s.groupRepo.Update(ctx, grp); err != nil {
		return fmt.Errorf("failed to update group settings: %w", err)
	}

	// Broadcast group settings updated event
	if s.broadcaster != nil {
		settingsDTO := dto.GroupSettings{
			WhoCanMessage:    settings.WhoCanMessage,
			WhoCanAddMembers: settings.WhoCanAddMembers,
			WhoCanEditInfo:   settings.WhoCanEditInfo,
		}
		go s.broadcaster.BroadcastGroupSettingsUpdated(context.Background(), grp.ConversationID, groupID.String(), userID.String(), settingsDTO)
	}

	return nil
}

// DeleteGroup deletes a group
func (s *service) DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Only creator can delete group
	if grp.CreatorID != userID {
		return group.ErrNotAuthorized
	}

	// Broadcast group deleted event before deletion
	if s.broadcaster != nil {
		go s.broadcaster.BroadcastGroupDeleted(context.Background(), grp.ConversationID, groupID.String(), userID.String())
	}

	// Delete group
	if err := s.groupRepo.Delete(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// AddMember adds a member to the group
func (s *service) AddMember(ctx context.Context, userID, groupID uuid.UUID, req *dto.AddMemberRequest) error {
	// Parse target user ID
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if requester is a member
	requester, err := s.groupRepo.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	// Check permissions
	if !requester.Permissions.CanAddMembers {
		return group.ErrNotAuthorized
	}

	// Check if target user exists
	if _, err := s.userRepo.FindByID(ctx, targetUserID); err != nil {
		return fmt.Errorf("target user not found: %w", err)
	}

	// Check if already a member
	isMember, err := s.groupRepo.IsMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}
	if isMember {
		return group.ErrAlreadyMember
	}

	// Check group capacity
	memberCount, err := s.groupRepo.CountMembers(ctx, groupID)
	if err != nil {
		return err
	}

	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if memberCount >= grp.MaxMembers {
		return group.ErrGroupFull
	}

	// Determine role
	role := group.RoleMember
	if req.Role != "" {
		role = group.Role(req.Role)
	}

	// Add member to group
	member := group.NewMember(groupID, targetUserID, role)
	if err := s.groupRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Add to conversation
	convRole := conversation.RoleMember
	if role == group.RoleAdmin {
		convRole = conversation.RoleAdmin
	}
	participant := conversation.NewParticipant(grp.ConversationID, targetUserID, convRole)
	if err := s.conversationRepo.AddParticipant(ctx, participant); err != nil {
		return fmt.Errorf("failed to add to conversation: %w", err)
	}

	// Broadcast member joined event
	if s.broadcaster != nil {
		targetUser, _ := s.userRepo.FindByID(ctx, targetUserID)
		username := ""
		if targetUser != nil {
			username = targetUser.Username
		}
		go s.broadcaster.BroadcastGroupMemberJoined(context.Background(), grp.ConversationID, groupID.String(), targetUserID.String(), username, string(role), userID.String())
	}

	return nil
}

// RemoveMember removes a member from the group
func (s *service) RemoveMember(ctx context.Context, userID, groupID, targetUserID uuid.UUID) error {
	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Cannot remove creator
	if grp.CreatorID == targetUserID {
		return group.ErrCannotRemoveCreator
	}

	// Check if requester is a member
	requester, err := s.groupRepo.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	// Check permissions
	if !requester.Permissions.CanRemoveMembers {
		return group.ErrNotAuthorized
	}

	// Remove member
	if err := s.groupRepo.RemoveMember(ctx, groupID, targetUserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Broadcast member removed event before removal
	if s.broadcaster != nil {
		targetUser, _ := s.userRepo.FindByID(ctx, targetUserID)
		username := ""
		if targetUser != nil {
			username = targetUser.Username
		}
		go s.broadcaster.BroadcastGroupMemberRemoved(context.Background(), grp.ConversationID, groupID.String(), targetUserID.String(), username, userID.String())
	}

	// Remove from conversation
	if err := s.conversationRepo.RemoveParticipant(ctx, grp.ConversationID, targetUserID); err != nil {
		return fmt.Errorf("failed to remove from conversation: %w", err)
	}

	return nil
}

// UpdateMemberRole updates a member's role
func (s *service) UpdateMemberRole(ctx context.Context, userID, groupID, targetUserID uuid.UUID, req *dto.UpdateMemberRoleRequest) error {
	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Check if requester is admin
	requester, err := s.groupRepo.FindMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	if !requester.IsAdmin() {
		return group.ErrNotAdmin
	}

	// Cannot change creator's role
	if grp.CreatorID == targetUserID {
		return group.ErrNotAuthorized
	}

	// Get current member to get previous role
	currentMember, err := s.groupRepo.FindMember(ctx, groupID, targetUserID)
	if err != nil {
		return err
	}
	previousRole := string(currentMember.Role)

	// Update role
	role := group.Role(req.Role)
	if err := s.groupRepo.UpdateMemberRole(ctx, groupID, targetUserID, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Broadcast member role changed event
	if s.broadcaster != nil {
		targetUser, _ := s.userRepo.FindByID(ctx, targetUserID)
		username := ""
		if targetUser != nil {
			username = targetUser.Username
		}
		go s.broadcaster.BroadcastGroupMemberRoleChanged(context.Background(), grp.ConversationID, groupID.String(), targetUserID.String(), username, previousRole, req.Role, userID.String())
	}

	return nil
}

// LeaveGroup allows a user to leave a group
func (s *service) LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	// Get group
	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	// Creator cannot leave
	if grp.CreatorID == userID {
		return fmt.Errorf("creator cannot leave group, must delete it instead")
	}

	// Get user info for broadcast
	var username string
	if usr, err := s.userRepo.FindByID(ctx, userID); err == nil {
		username = usr.Username
	}

	// Remove member
	if err := s.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return fmt.Errorf("failed to leave group: %w", err)
	}

	// Remove from conversation
	if err := s.conversationRepo.RemoveParticipant(ctx, grp.ConversationID, userID); err != nil {
		return fmt.Errorf("failed to leave conversation: %w", err)
	}

	// Broadcast member left event
	if s.broadcaster != nil {
		go s.broadcaster.BroadcastGroupMemberLeft(context.Background(), grp.ConversationID, groupID.String(), userID.String(), username)
	}

	return nil
}

// Helper function to convert group to DTO
func toGroupDTO(grp *group.Group, memberCount int) dto.GroupDTO {
	var settings *dto.GroupSettings
	if grp.Settings != nil {
		settings = &dto.GroupSettings{
			WhoCanMessage:    grp.Settings.WhoCanMessage,
			WhoCanAddMembers: grp.Settings.WhoCanAddMembers,
			WhoCanEditInfo:   grp.Settings.WhoCanEditInfo,
		}
	}

	return dto.GroupDTO{
		ID:             grp.ID.String(),
		ConversationID: grp.ConversationID.String(),
		Name:           grp.Name,
		Description:    grp.Description,
		Avatar:         grp.Avatar,
		CreatorID:      grp.CreatorID.String(),
		MaxMembers:     grp.MaxMembers,
		MemberCount:    memberCount,
		Settings:       settings,
		CreatedAt:      grp.CreatedAt,
		UpdatedAt:      grp.UpdatedAt,
	}
}
