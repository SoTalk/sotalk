package status

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/usecase/dto"
)

// Service defines the status/stories use case interface
type Service interface {
	// CreateStatus creates a new status
	CreateStatus(ctx context.Context, userID uuid.UUID, req *dto.CreateStatusRequest) (*dto.StatusResponse, error)

	// GetUserStatuses gets a user's active statuses
	GetUserStatuses(ctx context.Context, viewerID, targetUserID uuid.UUID) ([]*dto.StatusResponse, error)

	// GetStatusFeed gets status feed for user
	GetStatusFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.StatusResponse, error)

	// ViewStatus marks a status as viewed
	ViewStatus(ctx context.Context, statusID, viewerID uuid.UUID) error

	// GetStatusViews gets viewers of a status
	GetStatusViews(ctx context.Context, userID, statusID uuid.UUID) ([]*dto.StatusViewResponse, error)

	// DeleteStatus deletes a status
	DeleteStatus(ctx context.Context, userID, statusID uuid.UUID) error
}
