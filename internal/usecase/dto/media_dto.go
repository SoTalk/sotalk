package dto

import "time"

// UploadMediaResponse is the response DTO for uploading media
type UploadMediaResponse struct {
	Media MediaDTO `json:"media"`
}

// GetMediaResponse is the response DTO for getting a media
type GetMediaResponse struct {
	Media MediaDTO `json:"media"`
}

// GetMediaListResponse is the response DTO for getting media list
type GetMediaListResponse struct {
	Media []MediaDTO `json:"media"`
	Total int        `json:"total"`
}

// MediaDTO is the media data transfer object
type MediaDTO struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	MessageID     *string       `json:"message_id,omitempty"`
	Type          string        `json:"type"`
	FileName      string        `json:"file_name"`
	FileSize      int64         `json:"file_size"`
	FileSizeMB    float64       `json:"file_size_mb"`
	MimeType      string        `json:"mime_type"`
	URL           string        `json:"url"`
	ThumbnailURL  string        `json:"thumbnail_url,omitempty"`
	Metadata      *MediaMetadata `json:"metadata,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// MediaMetadata represents media metadata in DTO
type MediaMetadata struct {
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Duration   int    `json:"duration,omitempty"` // seconds
	Waveform   []int  `json:"waveform,omitempty"`
	Format     string `json:"format,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}
