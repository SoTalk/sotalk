package group

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the group repository interface
type Repository interface {
	// Group operations
	Create(ctx context.Context, group *Group) error
	FindByID(ctx context.Context, id uuid.UUID) (*Group, error)
	FindByConversationID(ctx context.Context, conversationID uuid.UUID) (*Group, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Member operations
	AddMember(ctx context.Context, member *Member) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	FindMember(ctx context.Context, groupID, userID uuid.UUID) (*Member, error)
	FindMembers(ctx context.Context, groupID uuid.UUID) ([]*Member, error)
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role Role) error
	CountMembers(ctx context.Context, groupID uuid.UUID) (int, error)
}
