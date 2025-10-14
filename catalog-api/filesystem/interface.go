package filesystem

import (
	"context"
	"io"
	"os"
	"time"
)

// FileInfo represents file information from any filesystem
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Mode    os.FileMode
	Path    string
}

// FileSystemClient defines the interface for filesystem operations
// This abstraction allows supporting multiple protocols (SMB, FTP, NFS, WebDAV, Local)
type FileSystemClient interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	TestConnection(ctx context.Context) error

	// File operations
	ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
	WriteFile(ctx context.Context, path string, data io.Reader) error
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
	FileExists(ctx context.Context, path string) (bool, error)
	DeleteFile(ctx context.Context, path string) error
	CopyFile(ctx context.Context, srcPath, dstPath string) error

	// Directory operations
	ListDirectory(ctx context.Context, path string) ([]*FileInfo, error)
	CreateDirectory(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, path string) error

	// Utility methods
	GetProtocol() string
	GetConfig() interface{}
}

// StorageConfig represents the configuration for a storage backend
type StorageConfig struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Protocol  string                 `json:"protocol"` // "smb", "ftp", "nfs", "webdav", "local"
	Enabled   bool                   `json:"enabled"`
	MaxDepth  int                    `json:"max_depth"`
	Settings  map[string]interface{} `json:"settings"` // Protocol-specific settings
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ClientFactory creates filesystem clients based on protocol
type ClientFactory interface {
	CreateClient(config *StorageConfig) (FileSystemClient, error)
	SupportedProtocols() []string
}

// CopyOperation represents a file copy operation
type CopyOperation struct {
	SourcePath        string
	DestinationPath   string
	OverwriteExisting bool
}

// CopyResult represents the result of a copy operation
type CopyResult struct {
	Success     bool
	BytesCopied int64
	Error       error
	TimeTaken   time.Duration
}

// DirectoryTreeInfo represents directory tree information
type DirectoryTreeInfo struct {
	Path       string
	TotalFiles int
	TotalDirs  int
	TotalSize  int64
	MaxDepth   int
	Files      []*FileInfo
	Subdirs    []*DirectoryTreeInfo
}

// ConnectionPool manages multiple connections for a protocol
type ConnectionPool interface {
	GetClient(config *StorageConfig) (FileSystemClient, error)
	ReturnClient(client FileSystemClient) error
	CloseAll() error
}
