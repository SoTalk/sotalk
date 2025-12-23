package entity

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// PasskeyCredential represents a WebAuthn passkey credential
type PasskeyCredential struct {
	ID              uuid.UUID                            `json:"id"`
	UserID          uuid.UUID                            `json:"user_id"`
	CredentialID    string                               `json:"credential_id"` // Base64 encoded
	PublicKey       []byte                               `json:"-"`             // COSE format
	AttestationType string                               `json:"attestation_type"`
	AAGUID          []byte                               `json:"-"`
	SignCount       uint32                               `json:"sign_count"`
	Transports      []protocol.AuthenticatorTransport    `json:"transports,omitempty"`
	BackupEligible  bool                                 `json:"backup_eligible"`
	BackupState     bool                                 `json:"backup_state"`
	CreatedAt       time.Time                            `json:"created_at"`
	LastUsedAt      *time.Time                           `json:"last_used_at,omitempty"`
}

// ToWebAuthnCredential converts to webauthn.Credential
func (pc *PasskeyCredential) ToWebAuthnCredential() webauthn.Credential {
	return webauthn.Credential{
		ID:              []byte(pc.CredentialID),
		PublicKey:       pc.PublicKey,
		AttestationType: pc.AttestationType,
		Transport:       pc.Transports,
		Flags: webauthn.CredentialFlags{
			BackupEligible: pc.BackupEligible,
			BackupState:    pc.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    pc.AAGUID,
			SignCount: pc.SignCount,
		},
	}
}

// UpdateFromAssertion updates credential after successful authentication
func (pc *PasskeyCredential) UpdateFromAssertion(authenticator *webauthn.Authenticator) {
	pc.SignCount = authenticator.SignCount
	now := time.Now()
	pc.LastUsedAt = &now
}
