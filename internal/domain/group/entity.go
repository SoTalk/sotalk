package group

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a member's role in the group
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
)

// Group represents a group chat entity
type Group struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	Name           string
	Description    string
	Avatar         string
	CreatorID      uuid.UUID
	MaxMembers     int
	Settings       *Settings
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Settings represents group settings
type Settings struct {
	WhoCanMessage   string // all, admins_only
	WhoCanAddMembers string // all, admins_only
	WhoCanEditInfo  string // all, admins_only
}

// Member represents a group member
type Member struct {
	GroupID     uuid.UUID
	UserID      uuid.UUID
	Role        Role
	Permissions *Permissions
	JoinedAt    time.Time
}

// Permissions represents member permissions
type Permissions struct {
	CanSendMessages  bool
	CanAddMembers    bool
	CanRemoveMembers bool
	CanEditInfo      bool
}

// NewGroup creates a new group
func NewGroup(name, description string, creatorID uuid.UUID, conversationID uuid.UUID) *Group {
	return &Group{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Name:           name,
		Description:    description,
		CreatorID:      creatorID,
		MaxMembers:     256,
		Settings: &Settings{
			WhoCanMessage:    "all",
			WhoCanAddMembers: "admins_only",
			WhoCanEditInfo:   "admins_only",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewMember creates a new group member
func NewMember(groupID, userID uuid.UUID, role Role) *Member {
	permissions := &Permissions{
		CanSendMessages:  true,
		CanAddMembers:    role == RoleAdmin || role == RoleModerator,
		CanRemoveMembers: role == RoleAdmin || role == RoleModerator,
		CanEditInfo:      role == RoleAdmin,
	}

	return &Member{
		GroupID:     groupID,
		UserID:      userID,
		Role:        role,
		Permissions: permissions,
		JoinedAt:    time.Now(),
	}
}

// UpdateInfo updates group information
func (g *Group) UpdateInfo(name, description, avatar string) {
	if name != "" {
		g.Name = name
	}
	if description != "" {
		g.Description = description
	}
	if avatar != "" {
		g.Avatar = avatar
	}
	g.UpdatedAt = time.Now()
}

// UpdateSettings updates group settings
func (g *Group) UpdateSettings(settings *Settings) {
	if settings != nil {
		g.Settings = settings
		g.UpdatedAt = time.Now()
	}
}

// PromoteToAdmin promotes a member to admin
func (m *Member) PromoteToAdmin() {
	m.Role = RoleAdmin
	m.Permissions.CanSendMessages = true
	m.Permissions.CanAddMembers = true
	m.Permissions.CanRemoveMembers = true
	m.Permissions.CanEditInfo = true
}

// PromoteToModerator promotes a member to moderator
func (m *Member) PromoteToModerator() {
	m.Role = RoleModerator
	m.Permissions.CanSendMessages = true
	m.Permissions.CanAddMembers = true
	m.Permissions.CanRemoveMembers = true
	m.Permissions.CanEditInfo = false
}

// DemoteToMember demotes a member to regular member
func (m *Member) DemoteToMember() {
	m.Role = RoleMember
	m.Permissions.CanSendMessages = true
	m.Permissions.CanAddMembers = false
	m.Permissions.CanRemoveMembers = false
	m.Permissions.CanEditInfo = false
}

// IsAdmin checks if member is an admin
func (m *Member) IsAdmin() bool {
	return m.Role == RoleAdmin
}

// IsModerator checks if member is a moderator
func (m *Member) IsModerator() bool {
	return m.Role == RoleModerator
}

// CanManageMembers checks if member can manage other members
func (m *Member) CanManageMembers() bool {
	return m.Role == RoleAdmin || m.Role == RoleModerator
}
