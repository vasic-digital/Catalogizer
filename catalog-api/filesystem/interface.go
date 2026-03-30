// Package filesystem provides unified filesystem client types for multi-protocol
// storage access (SMB, FTP, NFS, WebDAV, Local).
//
// The core types (FileInfo, FileSystemClient, StorageConfig, ClientFactory,
// CopyOperation, CopyResult, ConnectionPool) are type aliases to
// digital.vasic.filesystem/pkg/client, ensuring compatibility with the
// reusable vasic-digital module ecosystem.
//
// DirectoryTreeInfo is a Catalogizer-specific extension not present in
// the base module.
package filesystem

import (
	"digital.vasic.filesystem/pkg/client"
)

// FileInfo represents file information from any filesystem.
// Type alias to digital.vasic.filesystem/pkg/client.FileInfo.
type FileInfo = client.FileInfo

// FileSystemClient defines the interface for filesystem operations.
// Supports multiple protocols: SMB, FTP, NFS, WebDAV, Local.
//
// Type alias to digital.vasic.filesystem/pkg/client.Client.
type FileSystemClient = client.Client

// StorageConfig represents the configuration for a storage backend.
// Type alias to digital.vasic.filesystem/pkg/client.StorageConfig.
type StorageConfig = client.StorageConfig

// ClientFactory creates filesystem clients based on protocol.
// Type alias to digital.vasic.filesystem/pkg/client.Factory.
type ClientFactory = client.Factory

// CopyOperation represents a file copy operation.
// Type alias to digital.vasic.filesystem/pkg/client.CopyOperation.
type CopyOperation = client.CopyOperation

// CopyResult represents the result of a copy operation.
// Type alias to digital.vasic.filesystem/pkg/client.CopyResult.
type CopyResult = client.CopyResult

// ConnectionPool manages multiple connections for a protocol.
// Type alias to digital.vasic.filesystem/pkg/client.ConnectionPool.
type ConnectionPool = client.ConnectionPool

// ReadSeekCloser combines io.Reader, io.Seeker, and io.Closer.
// Type alias to digital.vasic.filesystem/pkg/client.ReadSeekCloser.
type ReadSeekCloser = client.ReadSeekCloser

// SeekableClient is an optional extension of FileSystemClient for protocols
// that support random-access reads (SMB, local). When a FileSystemClient also
// implements SeekableClient, the stream handler can use OpenSeekable to get a
// reader with Seek support, enabling HTTP Range requests for video seeking.
// Type alias to digital.vasic.filesystem/pkg/client.SeekableClient.
type SeekableClient = client.SeekableClient

// DirectoryTreeInfo represents directory tree information.
// This is a Catalogizer-specific extension not present in the base module.
type DirectoryTreeInfo struct {
	Path       string
	TotalFiles int
	TotalDirs  int
	TotalSize  int64
	MaxDepth   int
	Files      []*FileInfo
	Subdirs    []*DirectoryTreeInfo
}
