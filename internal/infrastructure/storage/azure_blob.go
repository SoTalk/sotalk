package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// AzureBlobService implements storage.Service using Azure Blob Storage
type AzureBlobService struct {
	client        *azblob.Client
	containerName string
	accountName   string
}

// AzureBlobConfig holds configuration for Azure Blob Storage
type AzureBlobConfig struct {
	AccountName   string
	AccountKey    string
	ContainerName string
}

// NewAzureBlobService creates a new Azure Blob storage service
func NewAzureBlobService(config AzureBlobConfig) (Service, error) {
	// Create credential from account key
	credential, err := azblob.NewSharedKeyCredential(config.AccountName, config.AccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create service URL
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", config.AccountName)

	// Create client
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	logger.Info("Azure Blob Storage initialized",
		zap.String("account", config.AccountName),
		zap.String("container", config.ContainerName),
	)

	return &AzureBlobService{
		client:        client,
		containerName: config.ContainerName,
		accountName:   config.AccountName,
	}, nil
}

// Upload uploads a file to Azure Blob Storage
func (s *AzureBlobService) Upload(ctx context.Context, fileName string, contentType string, data io.Reader) (string, error) {
	// Clean the file name to use as blob name
	blobName := cleanBlobName(fileName)

	// Upload to blob
	_, err := s.client.UploadStream(ctx, s.containerName, blobName, data, &azblob.UploadStreamOptions{
		Metadata: map[string]*string{
			"content-type": &contentType,
		},
	})
	if err != nil {
		logger.Error("Failed to upload to Azure Blob",
			zap.Error(err),
			zap.String("blob_name", blobName),
		)
		return "", fmt.Errorf("failed to upload blob: %w", err)
	}

	// Return the blob URL
	blobURL := s.GetURL(blobName)

	logger.Info("File uploaded to Azure Blob",
		zap.String("blob_name", blobName),
		zap.String("url", blobURL),
	)

	return blobURL, nil
}

// Delete deletes a file from Azure Blob Storage
func (s *AzureBlobService) Delete(ctx context.Context, filePath string) error {
	// Extract blob name from URL or path
	blobName := extractBlobName(filePath)

	// Delete the blob
	_, err := s.client.DeleteBlob(ctx, s.containerName, blobName, nil)
	if err != nil {
		logger.Error("Failed to delete blob",
			zap.Error(err),
			zap.String("blob_name", blobName),
		)
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	logger.Info("Blob deleted",
		zap.String("blob_name", blobName),
	)

	return nil
}

// GetURL returns the public URL for a blob
func (s *AzureBlobService) GetURL(blobName string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.accountName,
		s.containerName,
		blobName,
	)
}

// Helper functions

// cleanBlobName ensures the blob name is valid
func cleanBlobName(fileName string) string {
	// Remove leading slashes
	fileName = strings.TrimPrefix(fileName, "/")
	// Replace backslashes with forward slashes
	fileName = strings.ReplaceAll(fileName, "\\", "/")
	return fileName
}

// extractBlobName extracts the blob name from a full URL or path
func extractBlobName(urlOrPath string) string {
	// If it's a full URL, extract the blob name
	if strings.Contains(urlOrPath, "blob.core.windows.net") {
		parts := strings.Split(urlOrPath, "/")
		if len(parts) >= 2 {
			// Return everything after the container name
			return strings.Join(parts[len(parts)-1:], "/")
		}
	}
	// If it's already a blob name or path
	return cleanBlobName(urlOrPath)
}
