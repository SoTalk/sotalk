package contact

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the contact management use case interface
type Service interface {
	// AddContact adds a new contact
	AddContact(ctx context.Context, userID uuid.UUID, req *dto.AddContactRequest) (*dto.ContactResponse, error)

	// RemoveContact removes a contact
	RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error

	// GetContacts gets user's contacts
	GetContacts(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.ContactResponse, error)

	// GetFavoriteContacts gets user's favorite contacts
	GetFavoriteContacts(ctx context.Context, userID uuid.UUID) ([]*dto.ContactResponse, error)

	// SetFavorite sets or unsets a contact as favorite
	SetFavorite(ctx context.Context, userID, contactID uuid.UUID, favorite bool) error

	// SendInvite sends a contact invitation
	SendInvite(ctx context.Context, senderID uuid.UUID, req *dto.SendInviteRequest) (*dto.InviteResponse, error)

	// GetPendingInvites gets pending invitations
	GetPendingInvites(ctx context.Context, userID uuid.UUID) ([]*dto.InviteResponse, error)

	// AcceptInvite accepts a contact invitation
	AcceptInvite(ctx context.Context, recipientID, inviteID uuid.UUID) error

	// RejectInvite rejects a contact invitation
	RejectInvite(ctx context.Context, recipientID, inviteID uuid.UUID) error

	// SearchContacts searches user's contacts
	SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*dto.ContactResponse, error)
}
