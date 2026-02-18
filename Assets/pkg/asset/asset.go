package asset

import (
	"time"

	"github.com/google/uuid"
)

// ID is a unique identifier for an asset.
type ID string

// NewID generates a new unique asset ID.
func NewID() ID {
	return ID(uuid.New().String())
}

// String returns the string representation of the ID.
func (id ID) String() string {
	return string(id)
}

// Type represents the kind of asset.
type Type string

const (
	TypeImage             Type = "image"
	TypeVideoThumbnail    Type = "video_thumbnail"
	TypeAudioCover        Type = "audio_cover"
	TypeDocumentThumbnail Type = "document_thumbnail"
)

// Status represents the lifecycle state of an asset.
type Status string

const (
	StatusPending   Status = "pending"
	StatusResolving Status = "resolving"
	StatusReady     Status = "ready"
	StatusFailed    Status = "failed"
	StatusExpired   Status = "expired"
)

// Asset represents a lazily-resolved media asset.
type Asset struct {
	ID          ID
	Type        Type
	Status      Status
	ContentType string
	Size        int64
	SourceHint  string
	EntityType  string
	EntityID    string
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ResolvedAt  *time.Time
	ExpiresAt   *time.Time
}

// New creates a new Asset in pending status.
func New(assetType Type, entityType, entityID string) *Asset {
	now := time.Now()
	return &Asset{
		ID:         NewID(),
		Type:       assetType,
		Status:     StatusPending,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsTerminal returns true if the asset is in a terminal state.
func (a *Asset) IsTerminal() bool {
	return a.Status == StatusReady || a.Status == StatusFailed || a.Status == StatusExpired
}

// MarkResolving transitions the asset to resolving status.
func (a *Asset) MarkResolving() {
	a.Status = StatusResolving
	a.UpdatedAt = time.Now()
}

// MarkReady transitions the asset to ready status.
func (a *Asset) MarkReady(contentType string, size int64) {
	now := time.Now()
	a.Status = StatusReady
	a.ContentType = contentType
	a.Size = size
	a.ResolvedAt = &now
	a.UpdatedAt = now
}

// MarkFailed transitions the asset to failed status.
func (a *Asset) MarkFailed() {
	a.Status = StatusFailed
	a.UpdatedAt = time.Now()
}

// IsExpired returns true if the asset has passed its expiration time.
func (a *Asset) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}
