package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	domainContact "github.com/yourusername/sotalk/internal/domain/contact"
	"gorm.io/gorm"
)

type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) domainContact.Repository {
	return &contactRepository{db: db}
}

// Contact Management

func (r *contactRepository) AddContact(ctx context.Context, contact *domainContact.Contact) error {
	// Check if contact already exists
	var existing Contact
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", contact.UserID, contact.ContactID).
		First(&existing).Error

	if err == nil {
		return domainContact.ErrContactAlreadyExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := &Contact{
		UserID:      contact.UserID,
		ContactID:   contact.ContactID,
		DisplayName: contact.DisplayName,
		IsFavorite:  contact.IsFavorite,
		IsBlocked:   contact.IsBlocked,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	contact.ID = model.ID
	contact.AddedAt = model.AddedAt
	contact.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *contactRepository) RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", userID, contactID).
		Delete(&Contact{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainContact.ErrContactNotFound
	}

	return nil
}

func (r *contactRepository) GetContact(ctx context.Context, userID, contactID uuid.UUID) (*domainContact.Contact, error) {
	var model Contact
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", userID, contactID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainContact.ErrContactNotFound
		}
		return nil, err
	}

	return &domainContact.Contact{
		ID:          model.ID,
		UserID:      model.UserID,
		ContactID:   model.ContactID,
		DisplayName: model.DisplayName,
		IsFavorite:  model.IsFavorite,
		IsBlocked:   model.IsBlocked,
		AddedAt:     model.AddedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (r *contactRepository) GetUserContacts(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domainContact.Contact, error) {
	var models []Contact
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("added_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainContact.Contact, len(models))
	for i, m := range models {
		result[i] = &domainContact.Contact{
			ID:          m.ID,
			UserID:      m.UserID,
			ContactID:   m.ContactID,
			DisplayName: m.DisplayName,
			IsFavorite:  m.IsFavorite,
			IsBlocked:   m.IsBlocked,
			AddedAt:     m.AddedAt,
			UpdatedAt:   m.UpdatedAt,
		}
	}

	return result, nil
}

func (r *contactRepository) UpdateContact(ctx context.Context, contact *domainContact.Contact) error {
	updates := map[string]interface{}{
		"display_name": contact.DisplayName,
		"is_favorite":  contact.IsFavorite,
		"is_blocked":   contact.IsBlocked,
	}

	result := r.db.WithContext(ctx).
		Model(&Contact{}).
		Where("user_id = ? AND contact_id = ?", contact.UserID, contact.ContactID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainContact.ErrContactNotFound
	}

	return nil
}

func (r *contactRepository) IsContact(ctx context.Context, userID, targetID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Contact{}).
		Where("user_id = ? AND contact_id = ?", userID, targetID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Favorites

func (r *contactRepository) SetFavorite(ctx context.Context, userID, contactID uuid.UUID, favorite bool) error {
	result := r.db.WithContext(ctx).
		Model(&Contact{}).
		Where("user_id = ? AND contact_id = ?", userID, contactID).
		Update("is_favorite", favorite)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainContact.ErrContactNotFound
	}

	return nil
}

func (r *contactRepository) GetFavoriteContacts(ctx context.Context, userID uuid.UUID) ([]*domainContact.Contact, error) {
	var models []Contact
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_favorite = ?", userID, true).
		Order("added_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainContact.Contact, len(models))
	for i, m := range models {
		result[i] = &domainContact.Contact{
			ID:          m.ID,
			UserID:      m.UserID,
			ContactID:   m.ContactID,
			DisplayName: m.DisplayName,
			IsFavorite:  m.IsFavorite,
			IsBlocked:   m.IsBlocked,
			AddedAt:     m.AddedAt,
			UpdatedAt:   m.UpdatedAt,
		}
	}

	return result, nil
}

// Contact Invitations

func (r *contactRepository) CreateInvite(ctx context.Context, invite *domainContact.ContactInvite) error {
	// Check for pending invite
	hasPending, err := r.HasPendingInvite(ctx, invite.SenderID, invite.RecipientID)
	if err != nil {
		return err
	}

	if hasPending {
		return domainContact.ErrInviteAlreadyExists
	}

	model := &ContactInvite{
		SenderID:    invite.SenderID,
		RecipientID: invite.RecipientID,
		Message:     invite.Message,
		Status:      string(invite.Status),
		ExpiresAt:   invite.ExpiresAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	invite.ID = model.ID
	invite.CreatedAt = model.CreatedAt
	return nil
}

func (r *contactRepository) GetInvite(ctx context.Context, inviteID uuid.UUID) (*domainContact.ContactInvite, error) {
	var model ContactInvite
	if err := r.db.WithContext(ctx).Where("id = ?", inviteID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainContact.ErrInviteNotFound
		}
		return nil, err
	}

	return &domainContact.ContactInvite{
		ID:          model.ID,
		SenderID:    model.SenderID,
		RecipientID: model.RecipientID,
		Message:     model.Message,
		Status:      domainContact.InviteStatus(model.Status),
		ExpiresAt:   model.ExpiresAt,
		CreatedAt:   model.CreatedAt,
		RespondedAt: model.RespondedAt,
	}, nil
}

func (r *contactRepository) GetPendingInvites(ctx context.Context, recipientID uuid.UUID) ([]*domainContact.ContactInvite, error) {
	var models []ContactInvite
	if err := r.db.WithContext(ctx).
		Where("recipient_id = ? AND status = ? AND expires_at > ?", recipientID, string(domainContact.InviteStatusPending), time.Now()).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainContact.ContactInvite, len(models))
	for i, m := range models {
		result[i] = &domainContact.ContactInvite{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			Message:     m.Message,
			Status:      domainContact.InviteStatus(m.Status),
			ExpiresAt:   m.ExpiresAt,
			CreatedAt:   m.CreatedAt,
			RespondedAt: m.RespondedAt,
		}
	}

	return result, nil
}

func (r *contactRepository) GetSentInvites(ctx context.Context, senderID uuid.UUID) ([]*domainContact.ContactInvite, error) {
	var models []ContactInvite
	if err := r.db.WithContext(ctx).
		Where("sender_id = ?", senderID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainContact.ContactInvite, len(models))
	for i, m := range models {
		result[i] = &domainContact.ContactInvite{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			Message:     m.Message,
			Status:      domainContact.InviteStatus(m.Status),
			ExpiresAt:   m.ExpiresAt,
			CreatedAt:   m.CreatedAt,
			RespondedAt: m.RespondedAt,
		}
	}

	return result, nil
}

func (r *contactRepository) UpdateInviteStatus(ctx context.Context, inviteID uuid.UUID, status domainContact.InviteStatus) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       string(status),
		"responded_at": &now,
	}

	result := r.db.WithContext(ctx).
		Model(&ContactInvite{}).
		Where("id = ?", inviteID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainContact.ErrInviteNotFound
	}

	return nil
}

func (r *contactRepository) DeleteInvite(ctx context.Context, inviteID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", inviteID).Delete(&ContactInvite{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainContact.ErrInviteNotFound
	}

	return nil
}

func (r *contactRepository) HasPendingInvite(ctx context.Context, senderID, recipientID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ContactInvite{}).
		Where("sender_id = ? AND recipient_id = ? AND status = ? AND expires_at > ?",
			senderID, recipientID, string(domainContact.InviteStatusPending), time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Search

func (r *contactRepository) SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int) ([]*domainContact.Contact, error) {
	var models []Contact
	searchQuery := "%" + query + "%"

	q := r.db.WithContext(ctx).
		Where("user_id = ? AND display_name ILIKE ?", userID, searchQuery).
		Order("display_name ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*domainContact.Contact, len(models))
	for i, m := range models {
		result[i] = &domainContact.Contact{
			ID:          m.ID,
			UserID:      m.UserID,
			ContactID:   m.ContactID,
			DisplayName: m.DisplayName,
			IsFavorite:  m.IsFavorite,
			IsBlocked:   m.IsBlocked,
			AddedAt:     m.AddedAt,
			UpdatedAt:   m.UpdatedAt,
		}
	}

	return result, nil
}
