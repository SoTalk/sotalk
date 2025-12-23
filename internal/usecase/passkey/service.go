package passkey

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/entity"
	"github.com/yourusername/sotalk/internal/domain/repository"
	domainUser "github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

type service struct {
	webAuthn     *webauthn.WebAuthn
	userRepo     domainUser.Repository
	passkeyRepo  repository.PasskeyRepository
	sessionStore map[string]*webauthn.SessionData // In production, use Redis or similar
}

// NewService creates a new passkey service
func NewService(
	rpID string,
	rpOrigin string,
	userRepo domainUser.Repository,
	passkeyRepo repository.PasskeyRepository,
) (Service, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: "SoTalk",
		RPID:          rpID,
		RPOrigins:     []string{rpOrigin},
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &service{
		webAuthn:     webAuthn,
		userRepo:     userRepo,
		passkeyRepo:  passkeyRepo,
		sessionStore: make(map[string]*webauthn.SessionData),
	}, nil
}

// webAuthnUser adapts domain User to webauthn.User interface
type webAuthnUser struct {
	user        *domainUser.User
	credentials []webauthn.Credential
}

func (u *webAuthnUser) WebAuthnID() []byte {
	return []byte(u.user.ID.String())
}

func (u *webAuthnUser) WebAuthnName() string {
	return u.user.Username
}

func (u *webAuthnUser) WebAuthnDisplayName() string {
	return u.user.Username
}

func (u *webAuthnUser) WebAuthnIcon() string {
	if u.user.Avatar != nil {
		return *u.user.Avatar
	}
	return ""
}

func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// BeginRegistration starts the passkey registration process
func (s *service) BeginRegistration(ctx context.Context, req *dto.RegisterPasskeyBeginRequest) (*dto.RegisterPasskeyBeginResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to find user", zap.Error(err), zap.String("user_id", req.UserID.String()))
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get existing credentials
	existingCreds, err := s.passkeyRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user credentials", zap.Error(err))
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Convert to webauthn credentials
	credentials := make([]webauthn.Credential, len(existingCreds))
	for i, cred := range existingCreds {
		credentials[i] = cred.ToWebAuthnCredential()
	}

	webUser := &webAuthnUser{
		user:        user,
		credentials: credentials,
	}

	// Generate registration options
	options, sessionData, err := s.webAuthn.BeginRegistration(webUser)
	if err != nil {
		logger.Error("Failed to begin registration", zap.Error(err))
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Store session data (in production, use Redis with expiration)
	sessionKey := fmt.Sprintf("reg_%s", user.ID.String())
	s.sessionStore[sessionKey] = sessionData

	return &dto.RegisterPasskeyBeginResponse{
		Options: options,
	}, nil
}

// FinishRegistration completes the passkey registration process
func (s *service) FinishRegistration(ctx context.Context, req *dto.RegisterPasskeyFinishRequest) (*dto.RegisterPasskeyFinishResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get session data
	sessionKey := fmt.Sprintf("reg_%s", user.ID.String())
	sessionData, ok := s.sessionStore[sessionKey]
	if !ok {
		logger.Error("Session not found", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("session not found")
	}
	defer delete(s.sessionStore, sessionKey)

	// Get existing credentials
	existingCreds, err := s.passkeyRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user credentials", zap.Error(err))
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Convert to webauthn credentials
	credentials := make([]webauthn.Credential, len(existingCreds))
	for i, cred := range existingCreds {
		credentials[i] = cred.ToWebAuthnCredential()
	}

	webUser := &webAuthnUser{
		user:        user,
		credentials: credentials,
	}

	// Verify and create credential
	credential, err := s.webAuthn.CreateCredential(webUser, *sessionData, req.Credential)
	if err != nil {
		logger.Error("Failed to create credential", zap.Error(err))
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Store credential in database
	passkeyCredential := &entity.PasskeyCredential{
		ID:              uuid.New(),
		UserID:          user.ID,
		CredentialID:    base64.RawURLEncoding.EncodeToString(credential.ID),
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		Transports:      credential.Transport,
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		CreatedAt:       time.Now(),
	}

	if err := s.passkeyRepo.Create(ctx, passkeyCredential); err != nil {
		logger.Error("Failed to store credential", zap.Error(err))
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	logger.Info("Passkey registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("credential_id", passkeyCredential.CredentialID),
	)

	return &dto.RegisterPasskeyFinishResponse{
		CredentialID: passkeyCredential.CredentialID,
		CreatedAt:    passkeyCredential.CreatedAt,
	}, nil
}

// BeginAuthentication starts the passkey authentication process
func (s *service) BeginAuthentication(ctx context.Context, req *dto.AuthenticatePasskeyBeginRequest) (*dto.AuthenticatePasskeyBeginResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get user credentials
	existingCreds, err := s.passkeyRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user credentials", zap.Error(err))
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	if len(existingCreds) == 0 {
		return nil, fmt.Errorf("no passkeys registered for user")
	}

	// Convert to webauthn credentials
	credentials := make([]webauthn.Credential, len(existingCreds))
	for i, cred := range existingCreds {
		credentials[i] = cred.ToWebAuthnCredential()
	}

	webUser := &webAuthnUser{
		user:        user,
		credentials: credentials,
	}

	// Generate authentication options
	options, sessionData, err := s.webAuthn.BeginLogin(webUser)
	if err != nil {
		logger.Error("Failed to begin authentication", zap.Error(err))
		return nil, fmt.Errorf("failed to begin authentication: %w", err)
	}

	// Store session data
	sessionKey := fmt.Sprintf("auth_%s", user.ID.String())
	s.sessionStore[sessionKey] = sessionData

	return &dto.AuthenticatePasskeyBeginResponse{
		Options: options,
	}, nil
}

// FinishAuthentication completes the passkey authentication process
func (s *service) FinishAuthentication(ctx context.Context, req *dto.AuthenticatePasskeyFinishRequest) (*dto.AuthenticatePasskeyFinishResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get session data
	sessionKey := fmt.Sprintf("auth_%s", user.ID.String())
	sessionData, ok := s.sessionStore[sessionKey]
	if !ok {
		logger.Error("Session not found", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("session not found")
	}
	defer delete(s.sessionStore, sessionKey)

	// Get user credentials
	existingCreds, err := s.passkeyRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user credentials", zap.Error(err))
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Convert to webauthn credentials
	credentials := make([]webauthn.Credential, len(existingCreds))
	for i, cred := range existingCreds {
		credentials[i] = cred.ToWebAuthnCredential()
	}

	webUser := &webAuthnUser{
		user:        user,
		credentials: credentials,
	}

	// Verify authentication
	credential, err := s.webAuthn.ValidateLogin(webUser, *sessionData, req.Credential)
	if err != nil {
		logger.Error("Failed to validate authentication", zap.Error(err))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Update credential usage
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	passkeyCredential, err := s.passkeyRepo.GetByCredentialID(ctx, credentialID)
	if err != nil {
		logger.Error("Failed to get credential", zap.Error(err))
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Update sign count and last used
	passkeyCredential.UpdateFromAssertion(&credential.Authenticator)
	if err := s.passkeyRepo.Update(ctx, passkeyCredential); err != nil {
		logger.Error("Failed to update credential", zap.Error(err))
		// Don't fail authentication if update fails
	}

	logger.Info("Passkey authentication successful",
		zap.String("user_id", user.ID.String()),
		zap.String("credential_id", credentialID),
	)

	return &dto.AuthenticatePasskeyFinishResponse{
		Success: true,
		Message: "Authentication successful",
		UsedAt:  time.Now(),
	}, nil
}

// GetUserPasskeys retrieves all passkeys for a user
func (s *service) GetUserPasskeys(ctx context.Context, req *dto.GetPasskeysRequest) (*dto.GetPasskeysResponse, error) {
	credentials, err := s.passkeyRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user passkeys", zap.Error(err))
		return nil, fmt.Errorf("failed to get passkeys: %w", err)
	}

	passkeys := make([]dto.PasskeyDTO, len(credentials))
	for i, cred := range credentials {
		passkeys[i] = dto.PasskeyDTO{
			ID:              cred.ID.String(),
			UserID:          cred.UserID.String(),
			CredentialID:    cred.CredentialID,
			AttestationType: cred.AttestationType,
			Transports:      cred.Transports,
			BackupEligible:  cred.BackupEligible,
			BackupState:     cred.BackupState,
			CreatedAt:       cred.CreatedAt,
			LastUsedAt:      cred.LastUsedAt,
		}
	}

	return &dto.GetPasskeysResponse{
		Passkeys: passkeys,
	}, nil
}

// DeletePasskey removes a passkey credential
func (s *service) DeletePasskey(ctx context.Context, req *dto.DeletePasskeyRequest) (*dto.DeletePasskeyResponse, error) {
	// Get credential
	credential, err := s.passkeyRepo.GetByCredentialID(ctx, req.CredentialID)
	if err != nil {
		logger.Error("Failed to get credential", zap.Error(err))
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	// Verify ownership
	if credential.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: credential does not belong to user")
	}

	// Delete credential
	if err := s.passkeyRepo.Delete(ctx, credential.ID); err != nil {
		logger.Error("Failed to delete credential", zap.Error(err))
		return nil, fmt.Errorf("failed to delete credential: %w", err)
	}

	logger.Info("Passkey deleted successfully",
		zap.String("user_id", req.UserID.String()),
		zap.String("credential_id", req.CredentialID),
	)

	return &dto.DeletePasskeyResponse{
		Success: true,
		Message: "Passkey deleted successfully",
	}, nil
}
