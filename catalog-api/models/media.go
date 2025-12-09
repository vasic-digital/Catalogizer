package models

import (
	"database/sql/driver"
	"encoding/json"
)

// MediaCatalogItem represents a media item in catalog for API
type MediaCatalogItem struct {
	ID               int64                  `json:"id" db:"id"`
	Title            string                 `json:"title" db:"title"`
	MediaType        string                 `json:"media_type" db:"media_type"`
	Year             *int                   `json:"year" db:"year"`
	Description      *string                `json:"description" db:"description"`
	CoverImage       *string                `json:"cover_image" db:"cover_image"`
	Rating           *float64               `json:"rating" db:"rating"`
	Quality          *string                `json:"quality" db:"quality"`
	FileSize         *int64                 `json:"file_size" db:"file_size"`
	Duration         *int64                 `json:"duration" db:"duration"`
	DirectoryPath    string                 `json:"directory_path" db:"directory_path"`
	SMBPath          *string                `json:"smb_path" db:"smb_path"`
	CreatedAt        string                 `json:"created_at" db:"created_at"`
	UpdatedAt        string                 `json:"updated_at" db:"updated_at"`
	ExternalMetadata []ExternalMetadata     `json:"external_metadata" db:"external_metadata"`
	Versions         []MediaVersion         `json:"versions" db:"versions"`
	IsFavorite       bool                   `json:"is_favorite" db:"is_favorite"`
	WatchProgress    float64                `json:"watch_progress" db:"watch_progress"`
	LastWatched      *string                `json:"last_watched" db:"last_watched"`
	IsDownloaded     bool                   `json:"is_downloaded" db:"is_downloaded"`
}

// ExternalMetadata represents metadata from external providers
type ExternalMetadata struct {
	ID           int64             `json:"id" db:"id"`
	MediaID      int64             `json:"media_id" db:"media_id"`
	Provider     string            `json:"provider" db:"provider"`
	ExternalID   string            `json:"external_id" db:"external_id"`
	Title        string            `json:"title" db:"title"`
	Description  *string           `json:"description" db:"description"`
	Year         *int              `json:"year" db:"year"`
	Rating       *float64          `json:"rating" db:"rating"`
	PosterURL    *string           `json:"poster_url" db:"poster_url"`
	BackdropURL  *string           `json:"backdrop_url" db:"backdrop_url"`
	Genres       []string          `json:"genres" db:"genres"`
	Cast         []string          `json:"cast" db:"cast"`
	Crew         []string          `json:"crew" db:"crew"`
	Metadata     map[string]string `json:"metadata" db:"metadata"`
	LastUpdated   string            `json:"last_updated" db:"last_updated"`
}

// MediaVersion represents different versions of a media item
type MediaVersion struct {
	ID            int64     `json:"id" db:"id"`
	MediaID       int64     `json:"media_id" db:"media_id"`
	Version       string    `json:"version" db:"version"`
	Quality       string    `json:"quality" db:"quality"`
	FilePath      string    `json:"file_path" db:"file_path"`
	FileSize      int64     `json:"file_size" db:"file_size"`
	Codec         *string   `json:"codec" db:"codec"`
	Resolution    *string   `json:"resolution" db:"resolution"`
	Bitrate       *int64    `json:"bitrate" db:"bitrate"`
	Language      *string   `json:"language" db:"language"`
	FrameRate     *float64  `json:"frame_rate" db:"frame_rate"`
	AudioChannels *int      `json:"audio_channels" db:"audio_channels"`
	SampleRate    *int      `json:"sample_rate" db:"sample_rate"`
}

// MediaStats represents statistics about the media catalog
type MediaStats struct {
	TotalItems      int64            `json:"total_items"`
	ByType          map[string]int64 `json:"by_type"`
	ByQuality       map[string]int64 `json:"by_quality"`
	TotalSize       int64            `json:"total_size"`
	RecentAdditions int              `json:"recent_additions"`
}

// ExternalMetadataList implements driver.Valuer interface for ExternalMetadata slice
func (em ExternalMetadataList) Value() (driver.Value, error) {
	return json.Marshal(em)
}

// ExternalMetadataList implements sql.Scanner interface for ExternalMetadata slice
func (em *ExternalMetadataList) Scan(value interface{}) error {
	if value == nil {
		*em = ExternalMetadataList{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, em)
	case string:
		return json.Unmarshal([]byte(v), em)
	}
	return nil
}

// MediaVersionsJSON implements driver.Valuer interface for MediaVersion slice
func (mv MediaVersionsJSON) Value() (driver.Value, error) {
	return json.Marshal(mv)
}

// MediaVersionsJSON implements sql.Scanner interface for MediaVersion slice
func (mv *MediaVersionsJSON) Scan(value interface{}) error {
	if value == nil {
		*mv = MediaVersionsJSON{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, mv)
	case string:
		return json.Unmarshal([]byte(v), mv)
	}
	return nil
}

// Type aliases for the JSON serialization
type ExternalMetadataList []ExternalMetadata
type MediaVersionsJSON []MediaVersion