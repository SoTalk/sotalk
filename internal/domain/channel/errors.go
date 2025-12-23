package channel

import "errors"

var (
	// ErrChannelNotFound is returned when a channel is not found
	ErrChannelNotFound = errors.New("channel not found")

	// ErrUsernameAlreadyExists is returned when channel username already exists
	ErrUsernameAlreadyExists = errors.New("channel username already exists")

	// ErrSubscriberNotFound is returned when a subscriber is not found
	ErrSubscriberNotFound = errors.New("subscriber not found")

	// ErrNotSubscribed is returned when user is not subscribed to the channel
	ErrNotSubscribed = errors.New("user is not subscribed to this channel")

	// ErrAlreadySubscribed is returned when user is already subscribed
	ErrAlreadySubscribed = errors.New("user is already subscribed to this channel")

	// ErrAdminNotFound is returned when an admin is not found
	ErrAdminNotFound = errors.New("admin not found")

	// ErrNotAdmin is returned when user is not an admin
	ErrNotAdmin = errors.New("user is not an admin of this channel")

	// ErrNotOwner is returned when user is not the owner
	ErrNotOwner = errors.New("user is not the owner of this channel")

	// ErrNotAuthorized is returned when user is not authorized for an action
	ErrNotAuthorized = errors.New("user is not authorized for this action")

	// ErrCannotRemoveOwner is returned when trying to remove channel owner
	ErrCannotRemoveOwner = errors.New("cannot remove channel owner")
)
