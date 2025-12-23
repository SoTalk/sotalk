package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// LocalStorage implements storage.Service using local filesystem
type LocalStorage struct {
	basePath string
	baseURL  string
}

// LocalStorageConfig holds configuration for local storage
type LocalStorageConfig struct {
	BasePath string
	BaseURL  string
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(config LocalStorageConfig) (Service, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	logger.Info("Local storage initialized",
		zap.String("base_path", config.BasePath),
		zap.String("base_url", config.BaseURL),
	)

	return &LocalStorage{
		basePath: config.BasePath,
		baseURL:  config.BaseURL,
	}, nil
}

// Upload uploads a file to local filesystem
func (s *LocalStorage) Upload(ctx context.Context, fileName string, contentType string, data io.Reader) (string, error) {
	// Create full file path
	fullPath := filepath.Join(s.basePath, fileName)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy data to file
	if _, err := io.Copy(dst, data); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return the URL
	fileURL := s.GetURL(fileName)

	logger.Info("File uploaded to local storage",
		zap.String("file_name", fileName),
		zap.String("url", fileURL),
	)

	return fileURL, nil
}

// Delete deletes a file from local filesystem
func (s *LocalStorage) Delete(ctx context.Context, filePath string) error {
	// If filePath is a URL, extract the file path
	if filepath.IsAbs(filePath) || filepath.VolumeName(filePath) != "" {
		// It's already a path
	} else {
		// It might be a relative path or URL
		filePath = filepath.Join(s.basePath, filePath)
	}

	// Remove the file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to delete file",
			zap.Error(err),
			zap.String("file_path", filePath),
		)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted from local storage",
		zap.String("file_path", filePath),
	)

	return nil
}

// GetURL returns the public URL for a file
func (s *LocalStorage) GetURL(fileName string) string {
	// For local storage, return base URL + file path
	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", s.baseURL, fileName)
	}
	// Default to relative path
	return fmt.Sprintf("/%s/%s", s.basePath, fileName)
}
