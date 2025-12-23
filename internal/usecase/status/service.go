package status

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/contact"
	domainStatus "github.com/yourusername/sotalk/internal/domain/status"
	"github.com/yourusername/sotalk/internal/domain/user"
	"github.com/yourusername/sotalk/internal/usecase/dto"
	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

type service struct {
	statusRepo  domainStatus.Repository
	contactRepo contact.Repository
	userRepo    user.Repository
}

// NewService creates a new status service
func NewService(statusRepo domainStatus.Repository, contactRepo contact.Repository, userRepo user.Repository) Service {
	return &service{
		statusRepo:  statusRepo,
		contactRepo: contactRepo,
		userRepo:    userRepo,
	}
}

func (s *service) CreateStatus(ctx context.Context, userID uuid.UUID, req *dto.CreateStatusRequest) (*dto.StatusResponse, error) {
	mediaID, err := uuid.Parse(req.MediaID)
	if err != nil {
		return nil, fmt.Errorf("invalid media ID: %w", err)
	}

	privacy := domainStatus.StatusPrivacyContacts
	if req.Privacy != nil {
		privacy = domainStatus.StatusPrivacy(*req.Privacy)
	}

	caption := ""
	if req.Caption != nil {
		caption = *req.Caption
	}

	status := &domainStatus.Status{
		UserID:    userID,
		MediaID:   mediaID,
		Caption:   caption,
		Privacy:   privacy,
		ViewCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry
	}

	if err := s.statusRepo.CreateStatus(ctx, status); err != nil {
		return nil, err
	}

	logger.Info("Status created",
		zap.String("status_id", status.ID.String()),
		zap.String("user_id", userID.String()),
	)

	return &dto.StatusResponse{
		ID:        status.ID.String(),
		UserID:    status.UserID.String(),
		MediaID:   status.MediaID.String(),
		Caption:   status.Caption,
		Privacy:   string(status.Privacy),
		ViewCount: status.ViewCount,
		HasViewed: false,
		ExpiresAt: status.ExpiresAt,
		CreatedAt: status.CreatedAt,
	}, nil
}

func (s *service) GetUserStatuses(ctx context.Context, viewerID, targetUserID uuid.UUID) ([]*dto.StatusResponse, error) {
	statuses, err := s.statusRepo.GetUserStatuses(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// Filter based on privacy settings and contact relationship
	result := make([]*dto.StatusResponse, 0, len(statuses))
	for _, st := range statuses {
		// Check if viewer can see this status
		canView, err := s.canViewStatus(ctx, viewerID, targetUserID, st.Privacy)
		if err != nil {
			logger.Error("Failed to check status visibility",
				zap.String("status_id", st.ID.String()),
				zap.String("viewer_id", viewerID.String()),
				zap.Error(err),
			)
			continue // Skip on error
		}

		if !canView {
			continue // Skip if viewer cannot see this status
		}

		// Check if viewer has viewed
		hasViewed, _ := s.statusRepo.HasViewed(ctx, st.ID, viewerID)

		result = append(result, &dto.StatusResponse{
			ID:        st.ID.String(),
			UserID:    st.UserID.String(),
			MediaID:   st.MediaID.String(),
			Caption:   st.Caption,
			Privacy:   string(st.Privacy),
			ViewCount: st.ViewCount,
			HasViewed: hasViewed,
			ExpiresAt: st.ExpiresAt,
			CreatedAt: st.CreatedAt,
		})
	}

	return result, nil
}

// canViewStatus checks if a viewer can see a status based on privacy settings
func (s *service) canViewStatus(ctx context.Context, viewerID, statusOwnerID uuid.UUID, privacy domainStatus.StatusPrivacy) (bool, error) {
	// Owner can always see their own status
	if viewerID == statusOwnerID {
		return true, nil
	}

	switch privacy {
	case domainStatus.StatusPrivacyEveryone:
		// Anyone can view
		return true, nil

	case domainStatus.StatusPrivacyContacts:
		// Only contacts can view - check if they are mutual contacts
		isContact1, err := s.contactRepo.IsContact(ctx, statusOwnerID, viewerID)
		if err != nil {
			return false, fmt.Errorf("failed to check contact relationship: %w", err)
		}
		if !isContact1 {
			return false, nil
		}

		// Check if it's mutual (viewer also has owner as contact)
		isContact2, err := s.contactRepo.IsContact(ctx, viewerID, statusOwnerID)
		if err != nil {
			return false, fmt.Errorf("failed to check mutual contact relationship: %w", err)
		}
		return isContact2, nil

	case domainStatus.StatusPrivacyPrivate:
		// Only owner can view (already checked above)
		return false, nil

	default:
		// Unknown privacy setting, deny by default
		return false, nil
	}
}

func (s *service) GetStatusFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.StatusResponse, error) {
	if limit == 0 {
		limit = 20
	}

	statuses, err := s.statusRepo.GetStatusFeed(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.StatusResponse, len(statuses))
	for i, st := range statuses {
		hasViewed, _ := s.statusRepo.HasViewed(ctx, st.ID, userID)

		result[i] = &dto.StatusResponse{
			ID:        st.ID.String(),
			UserID:    st.UserID.String(),
			MediaID:   st.MediaID.String(),
			Caption:   st.Caption,
			Privacy:   string(st.Privacy),
			ViewCount: st.ViewCount,
			HasViewed: hasViewed,
			ExpiresAt: st.ExpiresAt,
			CreatedAt: st.CreatedAt,
		}
	}

	return result, nil
}

func (s *service) ViewStatus(ctx context.Context, statusID, viewerID uuid.UUID) error {
	// Get status
	status, err := s.statusRepo.GetStatus(ctx, statusID)
	if err != nil {
		return err
	}

	if status.IsExpired() {
		return domainStatus.ErrStatusExpired
	}

	// Add view
	view := &domainStatus.StatusView{
		StatusID: statusID,
		ViewerID: viewerID,
	}

	if err := s.statusRepo.AddView(ctx, view); err != nil {
		if err != domainStatus.ErrAlreadyViewed {
			return err
		}
		// Already viewed, that's fine
		return nil
	}

	// Increment view count
	if err := s.statusRepo.IncrementViewCount(ctx, statusID); err != nil {
		logger.Warn("Failed to increment view count", zap.Error(err))
	}

	logger.Info("Status viewed",
		zap.String("status_id", statusID.String()),
		zap.String("viewer_id", viewerID.String()),
	)

	return nil
}

func (s *service) GetStatusViews(ctx context.Context, userID, statusID uuid.UUID) ([]*dto.StatusViewResponse, error) {
	// Verify status belongs to user
	status, err := s.statusRepo.GetStatus(ctx, statusID)
	if err != nil {
		return nil, err
	}

	if status.UserID != userID {
		return nil, fmt.Errorf("unauthorized: status does not belong to user")
	}

	views, err := s.statusRepo.GetStatusViews(ctx, statusID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.StatusViewResponse, len(views))
	for i, v := range views {
		result[i] = &dto.StatusViewResponse{
			ViewerID: v.ViewerID.String(),
			ViewedAt: v.ViewedAt,
		}
	}

	return result, nil
}

func (s *service) DeleteStatus(ctx context.Context, userID, statusID uuid.UUID) error {
	// Verify status belongs to user
	status, err := s.statusRepo.GetStatus(ctx, statusID)
	if err != nil {
		return err
	}

	if status.UserID != userID {
		return fmt.Errorf("unauthorized: status does not belong to user")
	}

	if err := s.statusRepo.DeleteStatus(ctx, statusID); err != nil {
		return err
	}

	logger.Info("Status deleted",
		zap.String("status_id", statusID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}
