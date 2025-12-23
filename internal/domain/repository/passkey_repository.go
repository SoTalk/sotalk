package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/entity"
)

// Common repository errors
var (
	ErrNotFound = errors.New("resource not found")
)

// PasskeyRepository handles passkey credential persistence
type PasskeyRepository interface {
	// Create stores a new passkey credential
	Create(ctx context.Context, credential *entity.PasskeyCredential) error

	// GetByUserID retrieves all credentials for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PasskeyCredential, error)

	// GetByCredentialID retrieves a credential by its ID
	GetByCredentialID(ctx context.Context, credentialID string) (*entity.PasskeyCredential, error)

	// Update updates a credential (e.g., sign count after authentication)
	Update(ctx context.Context, credential *entity.PasskeyCredential) error

	// Delete removes a credential
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserID removes all credentials for a user
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
