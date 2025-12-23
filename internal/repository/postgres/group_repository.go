package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/group"
	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *gorm.DB) group.Repository {
	return &groupRepository{
		db: db,
	}
}

// Create creates a new group
func (r *groupRepository) Create(ctx context.Context, grp *group.Group) error {
	dbGroup := toDBGroup(grp)
	result := r.db.WithContext(ctx).Create(dbGroup)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID finds a group by ID
func (r *groupRepository) FindByID(ctx context.Context, id uuid.UUID) (*group.Group, error) {
	var dbGroup Group
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbGroup)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, group.ErrGroupNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainGroup(&dbGroup), nil
}

// FindByConversationID finds a group by conversation ID
func (r *groupRepository) FindByConversationID(ctx context.Context, conversationID uuid.UUID) (*group.Group, error) {
	var dbGroup Group
	result := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).First(&dbGroup)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, group.ErrGroupNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainGroup(&dbGroup), nil
}

// FindByUserID finds all groups where user is a member
func (r *groupRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*group.Group, error) {
	var dbGroups []Group

	result := r.db.WithContext(ctx).
		Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Order("groups.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbGroups)

	if result.Error != nil {
		return nil, result.Error
	}

	groups := make([]*group.Group, len(dbGroups))
	for i, dbGroup := range dbGroups {
		groups[i] = toDomainGroup(&dbGroup)
	}

	return groups, nil
}

// Update updates a group
func (r *groupRepository) Update(ctx context.Context, grp *group.Group) error {
	dbGroup := toDBGroup(grp)
	result := r.db.WithContext(ctx).
		Model(&Group{}).
		Where("id = ?", grp.ID).
		Updates(dbGroup)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return group.ErrGroupNotFound
	}
	return nil
}

// Delete deletes a group
func (r *groupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Group{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return group.ErrGroupNotFound
	}
	return nil
}

// AddMember adds a member to a group
func (r *groupRepository) AddMember(ctx context.Context, member *group.Member) error {
	dbMember := toDBGroupMember(member)
	result := r.db.WithContext(ctx).Create(dbMember)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// RemoveMember removes a member from a group
func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&GroupMember{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return group.ErrMemberNotFound
	}
	return nil
}

// FindMember finds a specific member in a group
func (r *groupRepository) FindMember(ctx context.Context, groupID, userID uuid.UUID) (*group.Member, error) {
	var dbMember GroupMember
	result := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&dbMember)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, group.ErrMemberNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainGroupMember(&dbMember), nil
}

// FindMembers finds all members of a group
func (r *groupRepository) FindMembers(ctx context.Context, groupID uuid.UUID) ([]*group.Member, error) {
	var dbMembers []GroupMember
	result := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("joined_at ASC").
		Find(&dbMembers)

	if result.Error != nil {
		return nil, result.Error
	}

	members := make([]*group.Member, len(dbMembers))
	for i, dbMember := range dbMembers {
		members[i] = toDomainGroupMember(&dbMember)
	}

	return members, nil
}

// IsMember checks if a user is a member of a group
func (r *groupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// UpdateMemberRole updates a member's role
func (r *groupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role group.Role) error {
	result := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", string(role))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return group.ErrMemberNotFound
	}
	return nil
}

// CountMembers counts the number of members in a group
func (r *groupRepository) CountMembers(ctx context.Context, groupID uuid.UUID) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// Mapper functions

func toDBGroup(grp *group.Group) *Group {
	var settingsJSON string
	if grp.Settings != nil {
		data, _ := json.Marshal(grp.Settings)
		settingsJSON = string(data)
	}

	return &Group{
		ID:             grp.ID,
		ConversationID: grp.ConversationID,
		Name:           grp.Name,
		Description:    grp.Description,
		Avatar:         grp.Avatar,
		CreatorID:      grp.CreatorID,
		MaxMembers:     grp.MaxMembers,
		Settings:       settingsJSON,
		CreatedAt:      grp.CreatedAt,
		UpdatedAt:      grp.UpdatedAt,
	}
}

func toDomainGroup(dbGroup *Group) *group.Group {
	var settings group.Settings
	if dbGroup.Settings != "" {
		json.Unmarshal([]byte(dbGroup.Settings), &settings)
	}

	return &group.Group{
		ID:             dbGroup.ID,
		ConversationID: dbGroup.ConversationID,
		Name:           dbGroup.Name,
		Description:    dbGroup.Description,
		Avatar:         dbGroup.Avatar,
		CreatorID:      dbGroup.CreatorID,
		MaxMembers:     dbGroup.MaxMembers,
		Settings:       &settings,
		CreatedAt:      dbGroup.CreatedAt,
		UpdatedAt:      dbGroup.UpdatedAt,
	}
}

func toDBGroupMember(member *group.Member) *GroupMember {
	var permissionsJSON string
	if member.Permissions != nil {
		data, _ := json.Marshal(member.Permissions)
		permissionsJSON = string(data)
	}

	return &GroupMember{
		GroupID:     member.GroupID,
		UserID:      member.UserID,
		Role:        string(member.Role),
		Permissions: permissionsJSON,
		JoinedAt:    member.JoinedAt,
	}
}

func toDomainGroupMember(dbMember *GroupMember) *group.Member {
	var permissions group.Permissions
	if dbMember.Permissions != "" {
		json.Unmarshal([]byte(dbMember.Permissions), &permissions)
	}

	return &group.Member{
		GroupID:     dbMember.GroupID,
		UserID:      dbMember.UserID,
		Role:        group.Role(dbMember.Role),
		Permissions: &permissions,
		JoinedAt:    dbMember.JoinedAt,
	}
}
