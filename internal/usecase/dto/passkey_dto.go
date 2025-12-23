package dto

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

// RegisterPasskeyBeginRequest is the request for beginning passkey registration
type RegisterPasskeyBeginRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// RegisterPasskeyBeginResponse is the response containing registration options
type RegisterPasskeyBeginResponse struct {
	Options *protocol.CredentialCreation `json:"options"`
}

// RegisterPasskeyFinishRequest is the request for finishing passkey registration
type RegisterPasskeyFinishRequest struct {
	UserID     uuid.UUID                                  `json:"user_id" validate:"required"`
	Credential *protocol.ParsedCredentialCreationData     `json:"credential" validate:"required"`
}

// RegisterPasskeyFinishResponse is the response after successful registration
type RegisterPasskeyFinishResponse struct {
	CredentialID string    `json:"credential_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthenticatePasskeyBeginRequest is the request for beginning passkey authentication
type AuthenticatePasskeyBeginRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// AuthenticatePasskeyBeginResponse is the response containing authentication challenge
type AuthenticatePasskeyBeginResponse struct {
	Options *protocol.CredentialAssertion `json:"options"`
}

// AuthenticatePasskeyFinishRequest is the request for finishing passkey authentication
type AuthenticatePasskeyFinishRequest struct {
	UserID     uuid.UUID                                 `json:"user_id" validate:"required"`
	Credential *protocol.ParsedCredentialAssertionData   `json:"credential" validate:"required"`
}

// AuthenticatePasskeyFinishResponse is the response after successful authentication
type AuthenticatePasskeyFinishResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	UsedAt  time.Time `json:"used_at"`
}

// GetPasskeysRequest is the request for getting user's passkeys
type GetPasskeysRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// PasskeyDTO is the data transfer object for passkey credential
type PasskeyDTO struct {
	ID              string                               `json:"id"`
	UserID          string                               `json:"user_id"`
	CredentialID    string                               `json:"credential_id"`
	AttestationType string                               `json:"attestation_type"`
	Transports      []protocol.AuthenticatorTransport    `json:"transports,omitempty"`
	BackupEligible  bool                                 `json:"backup_eligible"`
	BackupState     bool                                 `json:"backup_state"`
	CreatedAt       time.Time                            `json:"created_at"`
	LastUsedAt      *time.Time                           `json:"last_used_at,omitempty"`
}

// GetPasskeysResponse is the response containing user's passkeys
type GetPasskeysResponse struct {
	Passkeys []PasskeyDTO `json:"passkeys"`
}

// DeletePasskeyRequest is the request for deleting a passkey
type DeletePasskeyRequest struct {
	UserID       uuid.UUID `json:"user_id" validate:"required"`
	CredentialID string    `json:"credential_id" validate:"required"`
}

// DeletePasskeyResponse is the response after deleting a passkey
type DeletePasskeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
