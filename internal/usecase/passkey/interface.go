package passkey

import (
	"context"

	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the passkey use case interface
type Service interface {
	// BeginRegistration starts the passkey registration process
	BeginRegistration(ctx context.Context, req *dto.RegisterPasskeyBeginRequest) (*dto.RegisterPasskeyBeginResponse, error)

	// FinishRegistration completes the passkey registration process
	FinishRegistration(ctx context.Context, req *dto.RegisterPasskeyFinishRequest) (*dto.RegisterPasskeyFinishResponse, error)

	// BeginAuthentication starts the passkey authentication process
	BeginAuthentication(ctx context.Context, req *dto.AuthenticatePasskeyBeginRequest) (*dto.AuthenticatePasskeyBeginResponse, error)

	// FinishAuthentication completes the passkey authentication process
	FinishAuthentication(ctx context.Context, req *dto.AuthenticatePasskeyFinishRequest) (*dto.AuthenticatePasskeyFinishResponse, error)

	// GetUserPasskeys retrieves all passkeys for a user
	GetUserPasskeys(ctx context.Context, req *dto.GetPasskeysRequest) (*dto.GetPasskeysResponse, error)

	// DeletePasskey removes a passkey credential
	DeletePasskey(ctx context.Context, req *dto.DeletePasskeyRequest) (*dto.DeletePasskeyResponse, error)
}
