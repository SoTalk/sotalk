package media

import "errors"

var (
	// ErrMediaNotFound is returned when media is not found
	ErrMediaNotFound = errors.New("media not found")

	// ErrInvalidMediaType is returned when media type is invalid
	ErrInvalidMediaType = errors.New("invalid media type")

	// ErrFileTooLarge is returned when file size exceeds limit
	ErrFileTooLarge = errors.New("file size exceeds limit")

	// ErrInvalidFileFormat is returned when file format is not supported
	ErrInvalidFileFormat = errors.New("unsupported file format")

	// ErrStorageQuotaExceeded is returned when user storage quota is exceeded
	ErrStorageQuotaExceeded = errors.New("storage quota exceeded")

	// ErrUploadFailed is returned when upload fails
	ErrUploadFailed = errors.New("media upload failed")

	// ErrNotAuthorized is returned when user is not authorized to access media
	ErrNotAuthorized = errors.New("not authorized to access this media")
)
