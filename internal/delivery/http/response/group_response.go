package response

import "time"

// CreateGroupResponse is the HTTP response for creating a group
type CreateGroupResponse struct {
	Group GroupDTO `json:"group"`
}

// GetGroupResponse is the HTTP response for getting a group
type GetGroupResponse struct {
	Group   GroupDTO    `json:"group"`
	Members []MemberDTO `json:"members"`
}

// GetGroupsResponse is the HTTP response for getting user's groups
type GetGroupsResponse struct {
	Groups []GroupDTO `json:"groups"`
}

// GroupDTO is the group data in response
type GroupDTO struct {
	ID             string         `json:"id"`
	ConversationID string         `json:"conversation_id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Avatar         string         `json:"avatar,omitempty"`
	CreatorID      string         `json:"creator_id"`
	MaxMembers     int            `json:"max_members"`
	MemberCount    int            `json:"member_count"`
	Settings       *GroupSettings `json:"settings,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// GroupSettings represents group settings in response
type GroupSettings struct {
	WhoCanMessage    string `json:"who_can_message"`
	WhoCanAddMembers string `json:"who_can_add_members"`
	WhoCanEditInfo   string `json:"who_can_edit_info"`
}

// MemberDTO is the member data in response
type MemberDTO struct {
	UserID      string             `json:"user_id"`
	User        *UserDTO           `json:"user,omitempty"`
	Role        string             `json:"role"`
	Permissions *MemberPermissions `json:"permissions,omitempty"`
	JoinedAt    time.Time          `json:"joined_at"`
}

// MemberPermissions represents member permissions in response
type MemberPermissions struct {
	CanSendMessages  bool `json:"can_send_messages"`
	CanAddMembers    bool `json:"can_add_members"`
	CanRemoveMembers bool `json:"can_remove_members"`
	CanEditInfo      bool `json:"can_edit_info"`
}
