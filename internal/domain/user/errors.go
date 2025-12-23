package user

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when a user with the same wallet address already exists
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrUsernameAlreadyTaken is returned when a username is already taken
	ErrUsernameAlreadyTaken = errors.New("username already taken")

	// ErrInvalidWalletAddress is returned when the wallet address is invalid
	ErrInvalidWalletAddress = errors.New("invalid wallet address")

	// ErrInvalidPublicKey is returned when the public key is invalid
	ErrInvalidPublicKey = errors.New("invalid public key")

	// ErrInvalidUsername is returned when the username is invalid
	ErrInvalidUsername = errors.New("invalid username")
)
