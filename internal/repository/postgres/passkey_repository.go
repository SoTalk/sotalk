package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/entity"
	"github.com/yourusername/sotalk/internal/domain/repository"
	"gorm.io/gorm"
)

// passkeyRepository implements repository.PasskeyRepository interface
type passkeyRepository struct {
	db *gorm.DB
}

// NewPasskeyRepository creates a new passkey repository
func NewPasskeyRepository(db *gorm.DB) repository.PasskeyRepository {
	return &passkeyRepository{db: db}
}

// Create stores a new passkey credential
func (r *passkeyRepository) Create(ctx context.Context, credential *entity.PasskeyCredential) error {
	dbPasskey := toPasskeyModel(credential)
	result := r.db.WithContext(ctx).Create(dbPasskey)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return errors.New("passkey credential already exists")
		}
		return result.Error
	}

	// Update domain entity with generated fields
	credential.ID = dbPasskey.ID
	credential.CreatedAt = dbPasskey.CreatedAt

	return nil
}

// GetByUserID retrieves all credentials for a user
func (r *passkeyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PasskeyCredential, error) {
	var dbPasskeys []PasskeyCredential
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dbPasskeys)

	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to domain entities
	credentials := make([]*entity.PasskeyCredential, len(dbPasskeys))
	for i, dbPasskey := range dbPasskeys {
		credentials[i] = toDomainPasskey(&dbPasskey)
	}

	return credentials, nil
}

// GetByCredentialID retrieves a credential by its ID
func (r *passkeyRepository) GetByCredentialID(ctx context.Context, credentialID string) (*entity.PasskeyCredential, error) {
	var dbPasskey PasskeyCredential
	result := r.db.WithContext(ctx).
		Where("credential_id = ?", credentialID).
		First(&dbPasskey)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, result.Error
	}

	return toDomainPasskey(&dbPasskey), nil
}

// Update updates a credential
func (r *passkeyRepository) Update(ctx context.Context, credential *entity.PasskeyCredential) error {
	updates := map[string]interface{}{
		"sign_count":   credential.SignCount,
		"last_used_at": credential.LastUsedAt,
		"backup_state": credential.BackupState,
	}

	result := r.db.WithContext(ctx).
		Model(&PasskeyCredential{}).
		Where("id = ?", credential.ID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete removes a credential
func (r *passkeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&PasskeyCredential{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// DeleteByUserID removes all credentials for a user
func (r *passkeyRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&PasskeyCredential{})

	return result.Error
}

// toPasskeyModel converts domain entity to GORM model
func toPasskeyModel(credential *entity.PasskeyCredential) *PasskeyCredential {
	// Convert transports array to PostgreSQL text array format
	transports := make([]string, len(credential.Transports))
	for i, t := range credential.Transports {
		transports[i] = string(t)
	}
	transportsStr := "{" + strings.Join(transports, ",") + "}"

	return &PasskeyCredential{
		ID:              credential.ID,
		UserID:          credential.UserID,
		CredentialID:    credential.CredentialID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.AAGUID,
		SignCount:       credential.SignCount,
		Transports:      transportsStr,
		BackupEligible:  credential.BackupEligible,
		BackupState:     credential.BackupState,
		CreatedAt:       credential.CreatedAt,
		LastUsedAt:      credential.LastUsedAt,
	}
}

// toDomainPasskey converts GORM model to domain entity
func toDomainPasskey(dbPasskey *PasskeyCredential) *entity.PasskeyCredential {
	// Parse PostgreSQL text array format
	transportsStr := strings.Trim(dbPasskey.Transports, "{}")
	var transports []protocol.AuthenticatorTransport
	if transportsStr != "" {
		transportsList := strings.Split(transportsStr, ",")
		transports = make([]protocol.AuthenticatorTransport, len(transportsList))
		for i, t := range transportsList {
			transports[i] = protocol.AuthenticatorTransport(strings.TrimSpace(t))
		}
	}

	return &entity.PasskeyCredential{
		ID:              dbPasskey.ID,
		UserID:          dbPasskey.UserID,
		CredentialID:    dbPasskey.CredentialID,
		PublicKey:       dbPasskey.PublicKey,
		AttestationType: dbPasskey.AttestationType,
		AAGUID:          dbPasskey.AAGUID,
		SignCount:       dbPasskey.SignCount,
		Transports:      transports,
		BackupEligible:  dbPasskey.BackupEligible,
		BackupState:     dbPasskey.BackupState,
		CreatedAt:       dbPasskey.CreatedAt,
		LastUsedAt:      dbPasskey.LastUsedAt,
	}
}
