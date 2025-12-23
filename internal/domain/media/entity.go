package media

import (
	"time"

	"github.com/google/uuid"
)

// Type represents the type of media
type Type string

const (
	TypeImage    Type = "image"
	TypeVideo    Type = "video"
	TypeAudio    Type = "audio"
	TypeDocument Type = "document"
)

// Media represents a media file entity
type Media struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	MessageID     *uuid.UUID
	Type          Type
	FileName      string
	FileSize      int64
	MimeType      string
	URL           string
	ThumbnailURL  string
	Metadata      *Metadata
	CreatedAt     time.Time
}

// Metadata represents media metadata
type Metadata struct {
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Duration   int    `json:"duration,omitempty"` // seconds for audio/video
	Waveform   []int  `json:"waveform,omitempty"` // for audio
	Format     string `json:"format,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// NewMedia creates a new media entity
func NewMedia(userID uuid.UUID, mediaType Type, fileName, mimeType, url string, fileSize int64) *Media {
	return &Media{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      mediaType,
		FileName:  fileName,
		FileSize:  fileSize,
		MimeType:  mimeType,
		URL:       url,
		Metadata:  &Metadata{},
		CreatedAt: time.Now(),
	}
}

// SetThumbnail sets the thumbnail URL
func (m *Media) SetThumbnail(url string) {
	m.ThumbnailURL = url
}

// SetMetadata sets media metadata
func (m *Media) SetMetadata(metadata *Metadata) {
	if metadata != nil {
		m.Metadata = metadata
	}
}

// AttachToMessage attaches media to a message
func (m *Media) AttachToMessage(messageID uuid.UUID) {
	m.MessageID = &messageID
}

// IsImage checks if media is an image
func (m *Media) IsImage() bool {
	return m.Type == TypeImage
}

// IsVideo checks if media is a video
func (m *Media) IsVideo() bool {
	return m.Type == TypeVideo
}

// IsAudio checks if media is audio
func (m *Media) IsAudio() bool {
	return m.Type == TypeAudio
}

// IsDocument checks if media is a document
func (m *Media) IsDocument() bool {
	return m.Type == TypeDocument
}

// GetFileSizeInMB returns file size in MB
func (m *Media) GetFileSizeInMB() float64 {
	return float64(m.FileSize) / (1024 * 1024)
}
