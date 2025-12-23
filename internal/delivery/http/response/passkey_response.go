package response

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
)

// RegisterPasskeyBeginResponse is the HTTP response containing registration options
type RegisterPasskeyBeginResponse struct {
	Options *protocol.CredentialCreation `json:"options"`
}

// RegisterPasskeyFinishResponse is the HTTP response after successful registration
type RegisterPasskeyFinishResponse struct {
	CredentialID string    `json:"credential_id"`
	CreatedAt    time.Time `json:"created_at"`
	Message      string    `json:"message"`
}

// AuthenticatePasskeyBeginResponse is the HTTP response containing authentication challenge
type AuthenticatePasskeyBeginResponse struct {
	Options *protocol.CredentialAssertion `json:"options"`
}

// AuthenticatePasskeyFinishResponse is the HTTP response after successful authentication
type AuthenticatePasskeyFinishResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	UsedAt  time.Time `json:"used_at"`
}

// PasskeyDTO is the HTTP response for a passkey credential
type PasskeyDTO struct {
	ID              string                            `json:"id"`
	CredentialID    string                            `json:"credential_id"`
	AttestationType string                            `json:"attestation_type"`
	Transports      []string                          `json:"transports,omitempty"`
	BackupEligible  bool                              `json:"backup_eligible"`
	BackupState     bool                              `json:"backup_state"`
	CreatedAt       time.Time                         `json:"created_at"`
	LastUsedAt      *time.Time                        `json:"last_used_at,omitempty"`
}

// GetPasskeysResponse is the HTTP response containing user's passkeys
type GetPasskeysResponse struct {
	Passkeys []PasskeyDTO `json:"passkeys"`
	Count    int          `json:"count"`
}

// DeletePasskeyResponse is the HTTP response after deleting a passkey
type DeletePasskeyResponse struct {
	Message string `json:"message"`
}
