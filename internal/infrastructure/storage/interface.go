package storage

import (
	"context"
	"io"
)

// Service defines the interface for file storage operations
type Service interface {
	// Upload uploads a file and returns the URL
	Upload(ctx context.Context, fileName string, contentType string, data io.Reader) (string, error)

	// Delete deletes a file by its path
	Delete(ctx context.Context, filePath string) error

	// GetURL returns the public URL for a file
	GetURL(filePath string) string
}
