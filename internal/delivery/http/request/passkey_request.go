package request

import "github.com/go-webauthn/webauthn/protocol"

// RegisterPasskeyBeginRequest is the HTTP request for beginning passkey registration
type RegisterPasskeyBeginRequest struct {
	// UserID is extracted from JWT token in middleware
}

// RegisterPasskeyFinishRequest is the HTTP request for finishing passkey registration
type RegisterPasskeyFinishRequest struct {
	Credential *protocol.ParsedCredentialCreationData `json:"credential" binding:"required"`
}

// AuthenticatePasskeyBeginRequest is the HTTP request for beginning passkey authentication
type AuthenticatePasskeyBeginRequest struct {
	// UserID is extracted from JWT token in middleware
}

// AuthenticatePasskeyFinishRequest is the HTTP request for finishing passkey authentication
type AuthenticatePasskeyFinishRequest struct {
	Credential *protocol.ParsedCredentialAssertionData `json:"credential" binding:"required"`
}

// DeletePasskeyRequest is the HTTP request for deleting a passkey
type DeletePasskeyRequest struct {
	CredentialID string `json:"credential_id" binding:"required"`
}
