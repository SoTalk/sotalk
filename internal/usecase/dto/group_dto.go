package dto

import "time"

// CreateGroupRequest is the request DTO for creating a group
type CreateGroupRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"member_ids"` // Initial members
}

// CreateGroupResponse is the response DTO for creating a group
type CreateGroupResponse struct {
	Group GroupDTO `json:"group"`
}

// GetGroupResponse is the response DTO for getting a group
type GetGroupResponse struct {
	Group   GroupDTO     `json:"group"`
	Members []MemberDTO  `json:"members"`
}

// UpdateGroupRequest is the request DTO for updating a group
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// UpdateGroupSettingsRequest is the request DTO for updating group settings
type UpdateGroupSettingsRequest struct {
	WhoCanMessage    string `json:"who_can_message"`
	WhoCanAddMembers string `json:"who_can_add_members"`
	WhoCanEditInfo   string `json:"who_can_edit_info"`
}

// AddMemberRequest is the request DTO for adding a member
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"` // admin, moderator, member (default: member)
}

// UpdateMemberRoleRequest is the request DTO for updating member role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"` // admin, moderator, member
}

// GetGroupsResponse is the response DTO for getting user's groups
type GetGroupsResponse struct {
	Groups []GroupDTO `json:"groups"`
}

// GroupDTO is the group data transfer object
type GroupDTO struct {
	ID             string          `json:"id"`
	ConversationID string          `json:"conversation_id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Avatar         string          `json:"avatar,omitempty"`
	CreatorID      string          `json:"creator_id"`
	MaxMembers     int             `json:"max_members"`
	MemberCount    int             `json:"member_count"`
	Settings       *GroupSettings  `json:"settings,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// GroupSettings represents group settings in DTO
type GroupSettings struct {
	WhoCanMessage    string `json:"who_can_message"`
	WhoCanAddMembers string `json:"who_can_add_members"`
	WhoCanEditInfo   string `json:"who_can_edit_info"`
}

// MemberDTO is the member data transfer object
type MemberDTO struct {
	UserID      string             `json:"user_id"`
	User        *UserDTO           `json:"user,omitempty"`
	Role        string             `json:"role"`
	Permissions *MemberPermissions `json:"permissions,omitempty"`
	JoinedAt    time.Time          `json:"joined_at"`
}

// MemberPermissions represents member permissions in DTO
type MemberPermissions struct {
	CanSendMessages  bool `json:"can_send_messages"`
	CanAddMembers    bool `json:"can_add_members"`
	CanRemoveMembers bool `json:"can_remove_members"`
	CanEditInfo      bool `json:"can_edit_info"`
}
