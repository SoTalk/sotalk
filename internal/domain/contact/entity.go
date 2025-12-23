package contact

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Contact represents a user's contact
type Contact struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`      // Owner of the contact list
	ContactID   uuid.UUID `json:"contact_id"`   // The contact user
	DisplayName string    `json:"display_name"` // Custom name for this contact
	IsFavorite  bool      `json:"is_favorite"`
	IsBlocked   bool      `json:"is_blocked"` // Cached from privacy settings
	AddedAt     time.Time `json:"added_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactInvite represents an invitation to connect
type ContactInvite struct {
	ID           uuid.UUID    `json:"id"`
	SenderID     uuid.UUID    `json:"sender_id"`
	RecipientID  uuid.UUID    `json:"recipient_id"`
	Message      string       `json:"message"`
	Status       InviteStatus `json:"status"`
	ExpiresAt    time.Time    `json:"expires_at"`
	CreatedAt    time.Time    `json:"created_at"`
	RespondedAt  *time.Time   `json:"responded_at,omitempty"`
}

// InviteStatus represents the status of an invitation
type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRejected InviteStatus = "rejected"
	InviteStatusExpired  InviteStatus = "expired"
)

// IsExpired checks if the invite has expired
func (i *ContactInvite) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending checks if invite is still pending
func (i *ContactInvite) IsPending() bool {
	return i.Status == InviteStatusPending && !i.IsExpired()
}

// Repository defines the interface for contact operations
type Repository interface {
	// Contact Management
	AddContact(ctx context.Context, contact *Contact) error
	RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error
	GetContact(ctx context.Context, userID, contactID uuid.UUID) (*Contact, error)
	GetUserContacts(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Contact, error)
	UpdateContact(ctx context.Context, contact *Contact) error
	IsContact(ctx context.Context, userID, targetID uuid.UUID) (bool, error)

	// Favorites
	SetFavorite(ctx context.Context, userID, contactID uuid.UUID, favorite bool) error
	GetFavoriteContacts(ctx context.Context, userID uuid.UUID) ([]*Contact, error)

	// Contact Invitations
	CreateInvite(ctx context.Context, invite *ContactInvite) error
	GetInvite(ctx context.Context, inviteID uuid.UUID) (*ContactInvite, error)
	GetPendingInvites(ctx context.Context, recipientID uuid.UUID) ([]*ContactInvite, error)
	GetSentInvites(ctx context.Context, senderID uuid.UUID) ([]*ContactInvite, error)
	UpdateInviteStatus(ctx context.Context, inviteID uuid.UUID, status InviteStatus) error
	DeleteInvite(ctx context.Context, inviteID uuid.UUID) error
	HasPendingInvite(ctx context.Context, senderID, recipientID uuid.UUID) (bool, error)

	// Search
	SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*Contact, error)
}

// Errors
var (
	ErrContactNotFound       = errors.New("contact not found")
	ErrContactAlreadyExists  = errors.New("contact already exists")
	ErrCannotAddSelf         = errors.New("cannot add yourself as contact")
	ErrInviteNotFound        = errors.New("invite not found")
	ErrInviteExpired         = errors.New("invite has expired")
	ErrInviteAlreadyExists   = errors.New("pending invite already exists")
	ErrInviteAlreadyAnswered = errors.New("invite already answered")
	ErrUnauthorizedInvite    = errors.New("not authorized to respond to this invite")
)
