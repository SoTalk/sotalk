package group

import "errors"

var (
	// ErrGroupNotFound is returned when a group is not found
	ErrGroupNotFound = errors.New("group not found")

	// ErrMemberNotFound is returned when a member is not found
	ErrMemberNotFound = errors.New("member not found")

	// ErrNotMember is returned when user is not a member of the group
	ErrNotMember = errors.New("user is not a member of this group")

	// ErrNotAdmin is returned when user is not an admin
	ErrNotAdmin = errors.New("user is not an admin of this group")

	// ErrNotAuthorized is returned when user is not authorized for an action
	ErrNotAuthorized = errors.New("user is not authorized for this action")

	// ErrGroupFull is returned when group has reached max members
	ErrGroupFull = errors.New("group has reached maximum members")

	// ErrAlreadyMember is returned when user is already a member
	ErrAlreadyMember = errors.New("user is already a member of this group")

	// ErrCannotRemoveCreator is returned when trying to remove group creator
	ErrCannotRemoveCreator = errors.New("cannot remove group creator")
)
