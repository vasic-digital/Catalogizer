package smb

import (
	"os"
	"time"
)

// FileInfo represents file information from SMB
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Mode    os.FileMode
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

// SmbConnectionPool manages multiple SMB connections
type SmbConnectionPool struct {
	connections    map[string]*SmbClient
	maxConnections int
}

// NewSmbConnectionPool creates a new connection pool
func NewSmbConnectionPool(maxConnections int) *SmbConnectionPool {
	return &SmbConnectionPool{
		connections:    make(map[string]*SmbClient),
		maxConnections: maxConnections,
	}
}

// GetConnection gets or creates an SMB connection
func (p *SmbConnectionPool) GetConnection(key string, config *SmbConfig) (*SmbClient, error) {
	if client, exists := p.connections[key]; exists {
		// Test the existing connection
		if err := client.TestConnection(); err == nil {
			return client, nil
		}
		// Connection is stale, remove it
		client.Close()
		delete(p.connections, key)
	}

	// Create new connection
	client, err := NewSmbClient(config)
	if err != nil {
		return nil, err
	}

	// Store in pool if there's space
	if len(p.connections) < p.maxConnections {
		p.connections[key] = client
	}

	return client, nil
}

// CloseAll closes all connections in the pool
func (p *SmbConnectionPool) CloseAll() {
	for key, client := range p.connections {
		client.Close()
		delete(p.connections, key)
	}
}
