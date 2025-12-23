package request

// CreateGroupRequest is the HTTP request for creating a group
type CreateGroupRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"member_ids"`
}

// UpdateGroupRequest is the HTTP request for updating a group
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// AddMemberRequest is the HTTP request for adding a member
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"` // admin, moderator, member
}

// UpdateMemberRoleRequest is the HTTP request for updating member role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"` // admin, moderator, member
}

// UpdateGroupSettingsRequest is the HTTP request for updating group settings
type UpdateGroupSettingsRequest struct {
	WhoCanMessage    string `json:"who_can_message"`    // all, admins_only
	WhoCanAddMembers string `json:"who_can_add_members"` // all, admins_only
	WhoCanEditInfo   string `json:"who_can_edit_info"`   // all, admins_only
}

// GetGroupsRequest is the HTTP request for getting user's groups
type GetGroupsRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}
