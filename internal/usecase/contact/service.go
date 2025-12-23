package contact

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainContact "github.com/yourusername/sotalk/internal/domain/contact"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

type service struct {
	contactRepo domainContact.Repository
}

// NewService creates a new contact service
func NewService(contactRepo domainContact.Repository) Service {
	return &service{
		contactRepo: contactRepo,
	}
}

func (s *service) AddContact(ctx context.Context, userID uuid.UUID, req *dto.AddContactRequest) (*dto.ContactResponse, error) {
	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %w", err)
	}

	if userID == contactID {
		return nil, domainContact.ErrCannotAddSelf
	}

	displayName := ""
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	contact := &domainContact.Contact{
		UserID:      userID,
		ContactID:   contactID,
		DisplayName: displayName,
		IsFavorite:  false,
		IsBlocked:   false,
	}

	if err := s.contactRepo.AddContact(ctx, contact); err != nil {
		return nil, err
	}

	logger.Info("Contact added",
		zap.String("user_id", userID.String()),
		zap.String("contact_id", contactID.String()),
	)

	return &dto.ContactResponse{
		ID:          contact.ID.String(),
		ContactID:   contact.ContactID.String(),
		DisplayName: contact.DisplayName,
		IsFavorite:  contact.IsFavorite,
		AddedAt:     contact.AddedAt,
	}, nil
}

func (s *service) RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error {
	if err := s.contactRepo.RemoveContact(ctx, userID, contactID); err != nil {
		return err
	}

	logger.Info("Contact removed",
		zap.String("user_id", userID.String()),
		zap.String("contact_id", contactID.String()),
	)

	return nil
}

func (s *service) GetContacts(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.ContactResponse, error) {
	contacts, err := s.contactRepo.GetUserContacts(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.ContactResponse, len(contacts))
	for i, c := range contacts {
		result[i] = &dto.ContactResponse{
			ID:          c.ID.String(),
			ContactID:   c.ContactID.String(),
			DisplayName: c.DisplayName,
			IsFavorite:  c.IsFavorite,
			AddedAt:     c.AddedAt,
		}
	}

	return result, nil
}

func (s *service) GetFavoriteContacts(ctx context.Context, userID uuid.UUID) ([]*dto.ContactResponse, error) {
	contacts, err := s.contactRepo.GetFavoriteContacts(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.ContactResponse, len(contacts))
	for i, c := range contacts {
		result[i] = &dto.ContactResponse{
			ID:          c.ID.String(),
			ContactID:   c.ContactID.String(),
			DisplayName: c.DisplayName,
			IsFavorite:  c.IsFavorite,
			AddedAt:     c.AddedAt,
		}
	}

	return result, nil
}

func (s *service) SetFavorite(ctx context.Context, userID, contactID uuid.UUID, favorite bool) error {
	if err := s.contactRepo.SetFavorite(ctx, userID, contactID, favorite); err != nil {
		return err
	}

	logger.Info("Contact favorite updated",
		zap.String("user_id", userID.String()),
		zap.String("contact_id", contactID.String()),
		zap.Bool("favorite", favorite),
	)

	return nil
}

func (s *service) SendInvite(ctx context.Context, senderID uuid.UUID, req *dto.SendInviteRequest) (*dto.InviteResponse, error) {
	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient ID: %w", err)
	}

	if senderID == recipientID {
		return nil, fmt.Errorf("cannot invite yourself")
	}

	// Check if already contacts
	isContact, err := s.contactRepo.IsContact(ctx, senderID, recipientID)
	if err != nil {
		return nil, err
	}

	if isContact {
		return nil, fmt.Errorf("already contacts")
	}

	message := ""
	if req.Message != nil {
		message = *req.Message
	}

	invite := &domainContact.ContactInvite{
		SenderID:    senderID,
		RecipientID: recipientID,
		Message:     message,
		Status:      domainContact.InviteStatusPending,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.contactRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	logger.Info("Contact invite sent",
		zap.String("sender_id", senderID.String()),
		zap.String("recipient_id", recipientID.String()),
	)

	return &dto.InviteResponse{
		ID:          invite.ID.String(),
		SenderID:    invite.SenderID.String(),
		RecipientID: invite.RecipientID.String(),
		Message:     invite.Message,
		Status:      string(invite.Status),
		ExpiresAt:   invite.ExpiresAt,
		CreatedAt:   invite.CreatedAt,
	}, nil
}

func (s *service) GetPendingInvites(ctx context.Context, userID uuid.UUID) ([]*dto.InviteResponse, error) {
	invites, err := s.contactRepo.GetPendingInvites(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.InviteResponse, len(invites))
	for i, inv := range invites {
		result[i] = &dto.InviteResponse{
			ID:          inv.ID.String(),
			SenderID:    inv.SenderID.String(),
			RecipientID: inv.RecipientID.String(),
			Message:     inv.Message,
			Status:      string(inv.Status),
			ExpiresAt:   inv.ExpiresAt,
			CreatedAt:   inv.CreatedAt,
			RespondedAt: inv.RespondedAt,
		}
	}

	return result, nil
}

func (s *service) AcceptInvite(ctx context.Context, recipientID, inviteID uuid.UUID) error {
	// Get invite
	invite, err := s.contactRepo.GetInvite(ctx, inviteID)
	if err != nil {
		return err
	}

	if invite.RecipientID != recipientID {
		return domainContact.ErrUnauthorizedInvite
	}

	if !invite.IsPending() {
		return domainContact.ErrInviteAlreadyAnswered
	}

	// Update invite status
	if err := s.contactRepo.UpdateInviteStatus(ctx, inviteID, domainContact.InviteStatusAccepted); err != nil {
		return err
	}

	// Add bidirectional contacts
	contact1 := &domainContact.Contact{
		UserID:    recipientID,
		ContactID: invite.SenderID,
	}

	contact2 := &domainContact.Contact{
		UserID:    invite.SenderID,
		ContactID: recipientID,
	}

	if err := s.contactRepo.AddContact(ctx, contact1); err != nil {
		logger.Warn("Failed to add contact for recipient", zap.Error(err))
	}

	if err := s.contactRepo.AddContact(ctx, contact2); err != nil {
		logger.Warn("Failed to add contact for sender", zap.Error(err))
	}

	logger.Info("Contact invite accepted",
		zap.String("invite_id", inviteID.String()),
		zap.String("recipient_id", recipientID.String()),
	)

	return nil
}

func (s *service) RejectInvite(ctx context.Context, recipientID, inviteID uuid.UUID) error {
	// Get invite
	invite, err := s.contactRepo.GetInvite(ctx, inviteID)
	if err != nil {
		return err
	}

	if invite.RecipientID != recipientID {
		return domainContact.ErrUnauthorizedInvite
	}

	if !invite.IsPending() {
		return domainContact.ErrInviteAlreadyAnswered
	}

	// Update invite status
	if err := s.contactRepo.UpdateInviteStatus(ctx, inviteID, domainContact.InviteStatusRejected); err != nil {
		return err
	}

	logger.Info("Contact invite rejected",
		zap.String("invite_id", inviteID.String()),
		zap.String("recipient_id", recipientID.String()),
	)

	return nil
}

func (s *service) SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*dto.ContactResponse, error) {
	if limit == 0 {
		limit = 20
	}

	contacts, err := s.contactRepo.SearchContacts(ctx, userID, query, limit)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.ContactResponse, len(contacts))
	for i, c := range contacts {
		result[i] = &dto.ContactResponse{
			ID:          c.ID.String(),
			ContactID:   c.ContactID.String(),
			DisplayName: c.DisplayName,
			IsFavorite:  c.IsFavorite,
			AddedAt:     c.AddedAt,
		}
	}

	return result, nil
}
