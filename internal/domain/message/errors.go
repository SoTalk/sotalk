package message

import "errors"

var (
	// ErrMessageNotFound is returned when a message is not found
	ErrMessageNotFound = errors.New("message not found")

	// ErrInvalidContent is returned when message content is invalid
	ErrInvalidContent = errors.New("invalid message content")

	// ErrInvalidSender is returned when sender is invalid
	ErrInvalidSender = errors.New("invalid sender")

	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized to perform this action")
)
