package models

import (
	"time"
)

// File represents a file record from the catalog database
type File struct {
	ID              int64     `json:"id" db:"id"`
	StorageRootID   int64     `json:"storage_root_id" db:"storage_root_id"`
	StorageRootName string    `json:"storage_root_name" db:"storage_root_name"`
	Path            string    `json:"path" db:"path"`
	Name            string    `json:"name" db:"name"`
	Extension       *string   `json:"extension" db:"extension"`
	MimeType        *string   `json:"mime_type" db:"mime_type"`
	FileType        *string   `json:"file_type" db:"file_type"`
	Size            int64     `json:"size" db:"size"`
	IsDirectory     bool      `json:"is_directory" db:"is_directory"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	ModifiedAt      time.Time `json:"modified_at" db:"modified_at"`
	AccessedAt      *time.Time `json:"accessed_at" db:"accessed_at"`
	Deleted         bool      `json:"deleted" db:"deleted"`
	DeletedAt       *time.Time `json:"deleted_at" db:"deleted_at"`
	LastScanAt      time.Time `json:"last_scan_at" db:"last_scan_at"`
	LastVerifiedAt  *time.Time `json:"last_verified_at" db:"last_verified_at"`
	MD5             *string   `json:"md5" db:"md5"`
	SHA256          *string   `json:"sha256" db:"sha256"`
	SHA1            *string   `json:"sha1" db:"sha1"`
	BLAKE3          *string   `json:"blake3" db:"blake3"`
	QuickHash       *string   `json:"quick_hash" db:"quick_hash"`
	IsDuplicate     bool      `json:"is_duplicate" db:"is_duplicate"`
	DuplicateGroupID *int64   `json:"duplicate_group_id" db:"duplicate_group_id"`
	ParentID        *int64    `json:"parent_id" db:"parent_id"`
}

// StorageRoot represents a storage root configuration for any protocol
type StorageRoot struct {
	ID                      int64     `json:"id" db:"id"`
	Name                    string    `json:"name" db:"name"`
	Protocol                string    `json:"protocol" db:"protocol"` // smb, ftp, nfs, webdav, local
	Host                    *string   `json:"host,omitempty" db:"host"`
	Port                    *int      `json:"port,omitempty" db:"port"`
	Path                    *string   `json:"path,omitempty" db:"path"` // share for SMB, path for FTP/NFS/WebDAV, base_path for local
	Username                *string   `json:"username,omitempty" db:"username"`
	Password                *string   `json:"password,omitempty" db:"password"`
	Domain                  *string   `json:"domain,omitempty" db:"domain"` // SMB specific
	MountPoint              *string   `json:"mount_point,omitempty" db:"mount_point"` // NFS specific
	Options                 *string   `json:"options,omitempty" db:"options"` // NFS/WebDAV specific
	URL                     *string   `json:"url,omitempty" db:"url"` // WebDAV specific
	Enabled                 bool      `json:"enabled" db:"enabled"`
	MaxDepth                int       `json:"max_depth" db:"max_depth"`
	EnableDuplicateDetection bool     `json:"enable_duplicate_detection" db:"enable_duplicate_detection"`
	EnableMetadataExtraction bool     `json:"enable_metadata_extraction" db:"enable_metadata_extraction"`
	IncludePatterns         *string   `json:"include_patterns" db:"include_patterns"`
	ExcludePatterns         *string   `json:"exclude_patterns" db:"exclude_patterns"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
	LastScanAt              *time.Time `json:"last_scan_at" db:"last_scan_at"`
}

// SmbRoot represents an SMB root configuration (deprecated - use StorageRoot)
type SmbRoot struct {
	ID                      int64     `json:"id" db:"id"`
	Name                    string    `json:"name" db:"name"`
	Host                    string    `json:"host" db:"host"`
	Port                    int       `json:"port" db:"port"`
	Share                   string    `json:"share" db:"share"`
	Username                string    `json:"username" db:"username"`
	Domain                  *string   `json:"domain" db:"domain"`
	Enabled                 bool      `json:"enabled" db:"enabled"`
	MaxDepth                int       `json:"max_depth" db:"max_depth"`
	EnableDuplicateDetection bool     `json:"enable_duplicate_detection" db:"enable_duplicate_detection"`
	EnableMetadataExtraction bool     `json:"enable_metadata_extraction" db:"enable_metadata_extraction"`
	IncludePatterns         *string   `json:"include_patterns" db:"include_patterns"`
	ExcludePatterns         *string   `json:"exclude_patterns" db:"exclude_patterns"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
	LastScanAt              *time.Time `json:"last_scan_at" db:"last_scan_at"`
}

// FileMetadata represents file metadata
type FileMetadata struct {
	ID       int64  `json:"id" db:"id"`
	FileID   int64  `json:"file_id" db:"file_id"`
	Key      string `json:"key" db:"key"`
	Value    string `json:"value" db:"value"`
	DataType string `json:"data_type" db:"data_type"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	ID        int64     `json:"id" db:"id"`
	FileCount int       `json:"file_count" db:"file_count"`
	TotalSize int64     `json:"total_size" db:"total_size"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// VirtualPath represents a virtual file system path
type VirtualPath struct {
	ID         int64     `json:"id" db:"id"`
	Path       string    `json:"path" db:"path"`
	TargetType string    `json:"target_type" db:"target_type"`
	TargetID   int64     `json:"target_id" db:"target_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ScanHistory represents scan operation history
type ScanHistory struct {
	ID              int64      `json:"id" db:"id"`
	StorageRootID   int64      `json:"storage_root_id" db:"storage_root_id"`
	ScanType        string     `json:"scan_type" db:"scan_type"`
	Status          string     `json:"status" db:"status"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time" db:"end_time"`
	FilesProcessed  int        `json:"files_processed" db:"files_processed"`
	FilesAdded      int        `json:"files_added" db:"files_added"`
	FilesUpdated    int        `json:"files_updated" db:"files_updated"`
	FilesDeleted    int        `json:"files_deleted" db:"files_deleted"`
	ErrorCount      int        `json:"error_count" db:"error_count"`
	ErrorMessage    *string    `json:"error_message" db:"error_message"`
}

// FileWithMetadata represents a file with its metadata
type FileWithMetadata struct {
	File
	Metadata []FileMetadata `json:"metadata,omitempty"`
}

// DirectoryInfo represents directory information with statistics
type DirectoryInfo struct {
	Path            string `json:"path" db:"path"`
	Name            string `json:"name" db:"name"`
	StorageRootName string `json:"storage_root_name" db:"storage_root_name"`
	FileCount       int    `json:"file_count" db:"file_count"`
	DirectoryCount  int    `json:"directory_count" db:"directory_count"`
	TotalSize       int64  `json:"total_size" db:"total_size"`
	DuplicateCount  int    `json:"duplicate_count" db:"duplicate_count"`
	ModifiedAt      time.Time `json:"modified_at" db:"modified_at"`
}

// SearchFilter represents search filter criteria
type SearchFilter struct {
	Query           string    `json:"query,omitempty"`
	Path            string    `json:"path,omitempty"`
	Name            string    `json:"name,omitempty"`
	Extension       string    `json:"extension,omitempty"`
	FileType        string    `json:"file_type,omitempty"`
	MimeType        string    `json:"mime_type,omitempty"`
	StorageRoots    []string  `json:"storage_roots,omitempty"`
	MinSize         *int64    `json:"min_size,omitempty"`
	MaxSize         *int64    `json:"max_size,omitempty"`
	ModifiedAfter   *time.Time `json:"modified_after,omitempty"`
	ModifiedBefore  *time.Time `json:"modified_before,omitempty"`
	IncludeDeleted  bool      `json:"include_deleted"`
	OnlyDuplicates  bool      `json:"only_duplicates"`
	ExcludeDuplicates bool    `json:"exclude_duplicates"`
	IncludeDirectories bool   `json:"include_directories"`
}

// SortOptions represents sorting options
type SortOptions struct {
	Field string `json:"field"`  // name, size, modified_at, created_at, path, extension
	Order string `json:"order"`  // asc, desc
}

// PaginationOptions represents pagination options
type PaginationOptions struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// SearchResult represents search results with pagination
type SearchResult struct {
	Files      []FileWithMetadata `json:"files"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}