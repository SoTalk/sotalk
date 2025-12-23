package request

// CreateChannelRequest is the HTTP request for creating a channel
type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required"`
	Username    string `json:"username" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateChannelRequest is the HTTP request for updating a channel
type UpdateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// AddChannelAdminRequest is the HTTP request for adding a channel admin
type AddChannelAdminRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// UpdateChannelAdminPermissionsRequest is the HTTP request for updating admin permissions
type UpdateChannelAdminPermissionsRequest struct {
	CanPostMessages   bool `json:"can_post_messages"`
	CanEditMessages   bool `json:"can_edit_messages"`
	CanDeleteMessages bool `json:"can_delete_messages"`
	CanManageAdmins   bool `json:"can_manage_admins"`
}

// GetChannelsRequest is the HTTP request for getting channels
type GetChannelsRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}
