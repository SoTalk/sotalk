package conversation

import "errors"

var (
	// ErrConversationNotFound is returned when a conversation is not found
	ErrConversationNotFound = errors.New("conversation not found")

	// ErrNotParticipant is returned when user is not a participant
	ErrNotParticipant = errors.New("user is not a participant of this conversation")

	// ErrAlreadyParticipant is returned when user is already a participant
	ErrAlreadyParticipant = errors.New("user is already a participant")

	// ErrInvalidConversationType is returned when conversation type is invalid
	ErrInvalidConversationType = errors.New("invalid conversation type")
)
