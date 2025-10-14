package models

import (
	"time"
)

// Media types
const (
	MediaTypeVideo = "video"
	MediaTypeAudio = "audio"
	MediaTypeImage = "image"
	MediaTypeText  = "text"
	MediaTypeBook  = "book"
	MediaTypeGame  = "game"
	MediaTypeOther = "other"
)

// FileItem represents a simplified file/directory item for API responses
type FileItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}

// SearchResult represents a search result item
type SearchResult struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
}

// DirectoryInfo represents directory information with statistics
type DirectoryInfo struct {
	Path      string `json:"path"`
	TotalSize int64  `json:"total_size"`
	FileCount int64  `json:"file_count"`
}

// SMBPath represents a parsed SMB path
type SMBPath struct {
	Server string `json:"server"`
	Share  string `json:"share"`
	Path   string `json:"path"`
	Valid  bool   `json:"valid"`
}

// FileInfo represents a file or directory in the catalog
type FileInfo struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Path         string    `json:"path" db:"path"`
	IsDirectory  bool      `json:"is_directory" db:"is_directory"`
	Type         string    `json:"type" db:"type"`
	Size         int64     `json:"size" db:"size"`
	LastModified time.Time `json:"last_modified" db:"last_modified"`
	Hash         *string   `json:"hash,omitempty" db:"hash"`
	Extension    *string   `json:"extension,omitempty" db:"extension"`
	MimeType     *string   `json:"mime_type,omitempty" db:"mime_type"`
	MediaType    *string   `json:"media_type,omitempty" db:"media_type"`
	ParentID     *int64    `json:"parent_id,omitempty" db:"parent_id"`
	SmbRoot      string    `json:"smb_root" db:"smb_root"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// DirectoryStats represents statistics for a directory
type DirectoryStats struct {
	Path           string `json:"path"`
	TotalSize      int64  `json:"total_size"`
	FileCount      int64  `json:"file_count"`
	DirectoryCount int64  `json:"directory_count"`
	DuplicateCount int64  `json:"duplicate_count"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash      string     `json:"hash"`
	Size      int64      `json:"size"`
	Count     int        `json:"count"`
	Files     []FileInfo `json:"files"`
	TotalSize int64      `json:"total_size"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query       string   `json:"query" form:"query"`
	Path        string   `json:"path" form:"path"`
	Extension   string   `json:"extension" form:"extension"`
	MimeType    string   `json:"mime_type" form:"mime_type"`
	MinSize     *int64   `json:"min_size" form:"min_size"`
	MaxSize     *int64   `json:"max_size" form:"max_size"`
	SmbRoots    []string `json:"smb_roots" form:"smb_roots"`
	IsDirectory *bool    `json:"is_directory" form:"is_directory"`
	Limit       int      `json:"limit" form:"limit"`
	Offset      int      `json:"offset" form:"offset"`
	SortBy      string   `json:"sort_by" form:"sort_by"`
	SortOrder   string   `json:"sort_order" form:"sort_order"`
}

// CopyRequest represents a copy operation request
type CopyRequest struct {
	SourcePath      string `json:"source_path" binding:"required"`
	DestinationPath string `json:"destination_path" binding:"required"`
	SmbRoot         string `json:"smb_root"`
	Overwrite       bool   `json:"overwrite"`
}

// DownloadRequest represents a download request
type DownloadRequest struct {
	Paths   []string `json:"paths" binding:"required"`
	Format  string   `json:"format"` // zip, tar, tar.gz
	SmbRoot string   `json:"smb_root"`
}

// Statistics models
type OverallStats struct {
	TotalFiles         int64 `json:"total_files"`
	TotalDirectories   int64 `json:"total_directories"`
	TotalSize          int64 `json:"total_size"`
	TotalDuplicates    int64 `json:"total_duplicates"`
	DuplicateGroups    int64 `json:"duplicate_groups"`
	StorageRootsCount  int64 `json:"storage_roots_count"`
	ActiveStorageRoots int64 `json:"active_storage_roots"`
	LastScanTime       int64 `json:"last_scan_time"`
}

type FileTypeStats struct {
	FileType    string  `json:"file_type"`
	Extension   string  `json:"extension"`
	Count       int64   `json:"count"`
	TotalSize   int64   `json:"total_size"`
	AverageSize float64 `json:"average_size"`
}

type SizeDistribution struct {
	Tiny    int64 `json:"tiny"`    // < 1KB
	Small   int64 `json:"small"`   // 1KB - 1MB
	Medium  int64 `json:"medium"`  // 1MB - 10MB
	Large   int64 `json:"large"`   // 10MB - 100MB
	Huge    int64 `json:"huge"`    // 100MB - 1GB
	Massive int64 `json:"massive"` // > 1GB
}

// StorageRootStats represents statistics for a specific storage root
type StorageRootStats struct {
	Name             string `json:"name"`
	TotalFiles       int64  `json:"total_files"`
	TotalDirectories int64  `json:"total_directories"`
	TotalSize        int64  `json:"total_size"`
	DuplicateFiles   int64  `json:"duplicate_files"`
	DuplicateGroups  int64  `json:"duplicate_groups"`
	LastScanTime     int64  `json:"last_scan_time"`
	IsOnline         bool   `json:"is_online"`
}

// DuplicateStats represents duplicate file statistics
type DuplicateStats struct {
	TotalDuplicates       int64   `json:"total_duplicates"`
	DuplicateGroups       int64   `json:"duplicate_groups"`
	WastedSpace           int64   `json:"wasted_space"`
	LargestDuplicateGroup int64   `json:"largest_duplicate_group"`
	AverageGroupSize      float64 `json:"average_group_size"`
}

// DuplicateGroupStats represents statistics for a duplicate group
type DuplicateGroupStats struct {
	GroupID     int64  `json:"group_id"`
	FileCount   int64  `json:"file_count"`
	TotalSize   int64  `json:"total_size"`
	WastedSpace int64  `json:"wasted_space"`
	SamplePath  string `json:"sample_path"`
}

// AccessPatterns represents file access patterns
type AccessPatterns struct {
	RecentlyAccessed   int64    `json:"recently_accessed"`
	NeverAccessed      int64    `json:"never_accessed"`
	AccessFrequency    []int64  `json:"access_frequency"`
	PopularExtensions  []string `json:"popular_extensions"`
	PopularDirectories []string `json:"popular_directories"`
}

// GrowthTrends represents storage growth trends
type GrowthTrends struct {
	MonthlyGrowth   []MonthlyGrowth `json:"monthly_growth"`
	TotalGrowthRate float64         `json:"total_growth_rate"`
	FileGrowthRate  float64         `json:"file_growth_rate"`
	SizeGrowthRate  float64         `json:"size_growth_rate"`
}

// MonthlyGrowth represents growth data for a specific month
type MonthlyGrowth struct {
	Month      string `json:"month"`
	FilesAdded int64  `json:"files_added"`
	SizeAdded  int64  `json:"size_added"`
}

// ScanHistoryItem represents a scan operation in history
type ScanHistoryItem struct {
	ID             int64      `json:"id"`
	SmbRootName    string     `json:"smb_root_name"`
	ScanType       string     `json:"scan_type"`
	Status         string     `json:"status"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	FilesProcessed int64      `json:"files_processed"`
	FilesAdded     int64      `json:"files_added"`
	FilesUpdated   int64      `json:"files_updated"`
	FilesDeleted   int64      `json:"files_deleted"`
	ErrorCount     int64      `json:"error_count"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
}

// MediaMetadata represents media metadata information
type MediaMetadata struct {
	ID          int64                  `json:"id" db:"id"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description,omitempty" db:"description"`
	Genre       string                 `json:"genre,omitempty" db:"genre"`
	Year        *int                   `json:"year,omitempty" db:"year"`
	Rating      *float64               `json:"rating,omitempty" db:"rating"`
	Duration    *int                   `json:"duration,omitempty" db:"duration"`
	Language    string                 `json:"language,omitempty" db:"language"`
	Country     string                 `json:"country,omitempty" db:"country"`
	Director    string                 `json:"director,omitempty" db:"director"`
	Producer    string                 `json:"producer,omitempty" db:"producer"`
	Cast        []string               `json:"cast,omitempty" db:"cast"`
	MediaType   string                 `json:"media_type,omitempty" db:"media_type"`
	Resolution  string                 `json:"resolution,omitempty" db:"resolution"`
	FileSize    *int64                 `json:"file_size,omitempty" db:"file_size"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	ExternalIDs map[string]string      `json:"external_ids,omitempty" db:"external_ids"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// CoverArtResult represents cover art search results
type CoverArtResult struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Quality   string `json:"quality,omitempty"`
	Source    string `json:"source,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"`
}
