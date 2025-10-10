package services

import (
	"catalog-api/filesystem"
	"context"
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// LocalProtocolHandler handles rename detection for local filesystem
type LocalProtocolHandler struct {
	logger *zap.Logger
}

func NewLocalProtocolHandler(logger *zap.Logger) *LocalProtocolHandler {
	return &LocalProtocolHandler{logger: logger}
}

func (h *LocalProtocolHandler) GetFileIdentifier(ctx context.Context, path string, size int64, isDir bool) (string, error) {
	// For local filesystem, use inode information if available
	// Fallback to path + size + modification time
	return fmt.Sprintf("local:%s:%d:%t", path, size, isDir), nil
}

func (h *LocalProtocolHandler) PerformMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string, isDir bool) error {
	// Local filesystem moves are handled by the OS file system watcher
	// No explicit move operation needed here
	return nil
}

func (h *LocalProtocolHandler) ValidateMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	// Basic validation - ensure paths are different and target doesn't exist
	if oldPath == newPath {
		return fmt.Errorf("source and destination paths are the same")
	}

	exists, err := client.FileExists(ctx, newPath)
	if err != nil {
		return fmt.Errorf("failed to check if destination exists: %w", err)
	}

	if exists {
		return fmt.Errorf("destination path already exists: %s", newPath)
	}

	return nil
}

func (h *LocalProtocolHandler) GetMoveWindow() time.Duration {
	// Local filesystem operations are very fast
	return 2 * time.Second
}

func (h *LocalProtocolHandler) SupportsRealTimeNotification() bool {
	return true // Local filesystem supports inotify/fsnotify
}

// SMBProtocolHandler handles rename detection for SMB protocol
type SMBProtocolHandler struct {
	logger *zap.Logger
}

func NewSMBProtocolHandler(logger *zap.Logger) *SMBProtocolHandler {
	return &SMBProtocolHandler{logger: logger}
}

func (h *SMBProtocolHandler) GetFileIdentifier(ctx context.Context, path string, size int64, isDir bool) (string, error) {
	// For SMB, use file path, size, and directory flag
	// SMB doesn't have reliable inode-like identifiers across all implementations
	return fmt.Sprintf("smb:%s:%d:%t", path, size, isDir), nil
}

func (h *SMBProtocolHandler) PerformMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string, isDir bool) error {
	// For SMB, perform copy + delete since not all SMB servers support native move
	if isDir {
		return h.moveDirectory(ctx, client, oldPath, newPath)
	}
	return h.moveFile(ctx, client, oldPath, newPath)
}

func (h *SMBProtocolHandler) moveFile(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	// Copy file content
	if err := client.CopyFile(ctx, oldPath, newPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Delete original file
	if err := client.DeleteFile(ctx, oldPath); err != nil {
		// Attempt to clean up the copy if deletion fails
		client.DeleteFile(ctx, newPath)
		return fmt.Errorf("failed to delete original file: %w", err)
	}

	return nil
}

func (h *SMBProtocolHandler) moveDirectory(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	// List directory contents
	files, err := client.ListDirectory(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("failed to list directory contents: %w", err)
	}

	// Create destination directory
	if err := client.CreateDirectory(ctx, newPath); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move each file/subdirectory
	for _, file := range files {
		oldItemPath := filepath.Join(oldPath, file.Name)
		newItemPath := filepath.Join(newPath, file.Name)

		if file.IsDir {
			if err := h.moveDirectory(ctx, client, oldItemPath, newItemPath); err != nil {
				return fmt.Errorf("failed to move subdirectory %s: %w", file.Name, err)
			}
		} else {
			if err := h.moveFile(ctx, client, oldItemPath, newItemPath); err != nil {
				return fmt.Errorf("failed to move file %s: %w", file.Name, err)
			}
		}
	}

	// Delete original directory
	if err := client.DeleteDirectory(ctx, oldPath); err != nil {
		return fmt.Errorf("failed to delete original directory: %w", err)
	}

	return nil
}

func (h *SMBProtocolHandler) ValidateMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if oldPath == newPath {
		return fmt.Errorf("source and destination paths are the same")
	}

	exists, err := client.FileExists(ctx, newPath)
	if err != nil {
		return fmt.Errorf("failed to check if destination exists: %w", err)
	}

	if exists {
		return fmt.Errorf("destination path already exists: %s", newPath)
	}

	return nil
}

func (h *SMBProtocolHandler) GetMoveWindow() time.Duration {
	// SMB operations can be slower, especially over network
	return 10 * time.Second
}

func (h *SMBProtocolHandler) SupportsRealTimeNotification() bool {
	return false // SMB typically uses polling for change detection
}

// FTPProtocolHandler handles rename detection for FTP protocol
type FTPProtocolHandler struct {
	logger *zap.Logger
}

func NewFTPProtocolHandler(logger *zap.Logger) *FTPProtocolHandler {
	return &FTPProtocolHandler{logger: logger}
}

func (h *FTPProtocolHandler) GetFileIdentifier(ctx context.Context, path string, size int64, isDir bool) (string, error) {
	// For FTP, use path + size + directory flag
	// Some FTP servers provide modification time which could be added
	return fmt.Sprintf("ftp:%s:%d:%t", path, size, isDir), nil
}

func (h *FTPProtocolHandler) PerformMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string, isDir bool) error {
	// FTP doesn't typically support atomic moves, so use copy + delete
	if isDir {
		return h.moveDirectory(ctx, client, oldPath, newPath)
	}
	return h.moveFile(ctx, client, oldPath, newPath)
}

func (h *FTPProtocolHandler) moveFile(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if err := client.CopyFile(ctx, oldPath, newPath); err != nil {
		return fmt.Errorf("failed to copy file via FTP: %w", err)
	}

	if err := client.DeleteFile(ctx, oldPath); err != nil {
		client.DeleteFile(ctx, newPath) // Cleanup on failure
		return fmt.Errorf("failed to delete original file via FTP: %w", err)
	}

	return nil
}

func (h *FTPProtocolHandler) moveDirectory(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	files, err := client.ListDirectory(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("failed to list FTP directory: %w", err)
	}

	if err := client.CreateDirectory(ctx, newPath); err != nil {
		return fmt.Errorf("failed to create FTP directory: %w", err)
	}

	for _, file := range files {
		oldItemPath := filepath.Join(oldPath, file.Name)
		newItemPath := filepath.Join(newPath, file.Name)

		if file.IsDir {
			if err := h.moveDirectory(ctx, client, oldItemPath, newItemPath); err != nil {
				return err
			}
		} else {
			if err := h.moveFile(ctx, client, oldItemPath, newItemPath); err != nil {
				return err
			}
		}
	}

	return client.DeleteDirectory(ctx, oldPath)
}

func (h *FTPProtocolHandler) ValidateMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if oldPath == newPath {
		return fmt.Errorf("source and destination paths are the same")
	}

	exists, err := client.FileExists(ctx, newPath)
	if err != nil {
		return fmt.Errorf("failed to check FTP destination: %w", err)
	}

	if exists {
		return fmt.Errorf("FTP destination already exists: %s", newPath)
	}

	return nil
}

func (h *FTPProtocolHandler) GetMoveWindow() time.Duration {
	// FTP operations can be slow, especially for large files
	return 30 * time.Second
}

func (h *FTPProtocolHandler) SupportsRealTimeNotification() bool {
	return false // FTP requires polling
}

// NFSProtocolHandler handles rename detection for NFS protocol
type NFSProtocolHandler struct {
	logger *zap.Logger
}

func NewNFSProtocolHandler(logger *zap.Logger) *NFSProtocolHandler {
	return &NFSProtocolHandler{logger: logger}
}

func (h *NFSProtocolHandler) GetFileIdentifier(ctx context.Context, path string, size int64, isDir bool) (string, error) {
	// NFS can potentially provide inode information
	// For now, use path + size + directory flag
	return fmt.Sprintf("nfs:%s:%d:%t", path, size, isDir), nil
}

func (h *NFSProtocolHandler) PerformMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string, isDir bool) error {
	// NFS supports native rename operations in most cases
	// For simplicity, use copy + delete approach
	if isDir {
		return h.moveDirectory(ctx, client, oldPath, newPath)
	}
	return h.moveFile(ctx, client, oldPath, newPath)
}

func (h *NFSProtocolHandler) moveFile(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if err := client.CopyFile(ctx, oldPath, newPath); err != nil {
		return fmt.Errorf("failed to copy file via NFS: %w", err)
	}

	if err := client.DeleteFile(ctx, oldPath); err != nil {
		client.DeleteFile(ctx, newPath)
		return fmt.Errorf("failed to delete original file via NFS: %w", err)
	}

	return nil
}

func (h *NFSProtocolHandler) moveDirectory(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	files, err := client.ListDirectory(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("failed to list NFS directory: %w", err)
	}

	if err := client.CreateDirectory(ctx, newPath); err != nil {
		return fmt.Errorf("failed to create NFS directory: %w", err)
	}

	for _, file := range files {
		oldItemPath := filepath.Join(oldPath, file.Name)
		newItemPath := filepath.Join(newPath, file.Name)

		if file.IsDir {
			if err := h.moveDirectory(ctx, client, oldItemPath, newItemPath); err != nil {
				return err
			}
		} else {
			if err := h.moveFile(ctx, client, oldItemPath, newItemPath); err != nil {
				return err
			}
		}
	}

	return client.DeleteDirectory(ctx, oldPath)
}

func (h *NFSProtocolHandler) ValidateMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if oldPath == newPath {
		return fmt.Errorf("source and destination paths are the same")
	}

	exists, err := client.FileExists(ctx, newPath)
	if err != nil {
		return fmt.Errorf("failed to check NFS destination: %w", err)
	}

	if exists {
		return fmt.Errorf("NFS destination already exists: %s", newPath)
	}

	return nil
}

func (h *NFSProtocolHandler) GetMoveWindow() time.Duration {
	// NFS operations are generally fast
	return 5 * time.Second
}

func (h *NFSProtocolHandler) SupportsRealTimeNotification() bool {
	// NFS can support inotify if mounted locally, but generally uses polling
	return false
}

// WebDAVProtocolHandler handles rename detection for WebDAV protocol
type WebDAVProtocolHandler struct {
	logger *zap.Logger
}

func NewWebDAVProtocolHandler(logger *zap.Logger) *WebDAVProtocolHandler {
	return &WebDAVProtocolHandler{logger: logger}
}

func (h *WebDAVProtocolHandler) GetFileIdentifier(ctx context.Context, path string, size int64, isDir bool) (string, error) {
	// WebDAV can provide ETags for some servers
	// For now, use path + size + directory flag
	pathHash := h.hashString(path)
	return fmt.Sprintf("webdav:%s:%d:%t", pathHash, size, isDir), nil
}

func (h *WebDAVProtocolHandler) hashString(s string) string {
	hash := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", hash)
}

func (h *WebDAVProtocolHandler) PerformMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string, isDir bool) error {
	// WebDAV supports MOVE method, but not all implementations support it
	// Use copy + delete as fallback
	if isDir {
		return h.moveDirectory(ctx, client, oldPath, newPath)
	}
	return h.moveFile(ctx, client, oldPath, newPath)
}

func (h *WebDAVProtocolHandler) moveFile(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if err := client.CopyFile(ctx, oldPath, newPath); err != nil {
		return fmt.Errorf("failed to copy file via WebDAV: %w", err)
	}

	if err := client.DeleteFile(ctx, oldPath); err != nil {
		client.DeleteFile(ctx, newPath)
		return fmt.Errorf("failed to delete original file via WebDAV: %w", err)
	}

	return nil
}

func (h *WebDAVProtocolHandler) moveDirectory(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	files, err := client.ListDirectory(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("failed to list WebDAV directory: %w", err)
	}

	if err := client.CreateDirectory(ctx, newPath); err != nil {
		return fmt.Errorf("failed to create WebDAV directory: %w", err)
	}

	for _, file := range files {
		oldItemPath := filepath.Join(oldPath, file.Name)
		newItemPath := filepath.Join(newPath, file.Name)

		if file.IsDir {
			if err := h.moveDirectory(ctx, client, oldItemPath, newItemPath); err != nil {
				return err
			}
		} else {
			if err := h.moveFile(ctx, client, oldItemPath, newItemPath); err != nil {
				return err
			}
		}
	}

	return client.DeleteDirectory(ctx, oldPath)
}

func (h *WebDAVProtocolHandler) ValidateMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error {
	if oldPath == newPath {
		return fmt.Errorf("source and destination paths are the same")
	}

	exists, err := client.FileExists(ctx, newPath)
	if err != nil {
		return fmt.Errorf("failed to check WebDAV destination: %w", err)
	}

	if exists {
		return fmt.Errorf("WebDAV destination already exists: %s", newPath)
	}

	return nil
}

func (h *WebDAVProtocolHandler) GetMoveWindow() time.Duration {
	// WebDAV operations depend on network latency
	return 15 * time.Second
}

func (h *WebDAVProtocolHandler) SupportsRealTimeNotification() bool {
	return false // WebDAV requires polling
}

// ProtocolHandlerFactory creates protocol handlers based on configuration
type ProtocolHandlerFactory struct {
	logger *zap.Logger
}

func NewProtocolHandlerFactory(logger *zap.Logger) *ProtocolHandlerFactory {
	return &ProtocolHandlerFactory{logger: logger}
}

func (f *ProtocolHandlerFactory) CreateHandler(protocol string) (ProtocolHandler, error) {
	switch strings.ToLower(protocol) {
	case "local":
		return NewLocalProtocolHandler(f.logger), nil
	case "smb":
		return NewSMBProtocolHandler(f.logger), nil
	case "ftp":
		return NewFTPProtocolHandler(f.logger), nil
	case "nfs":
		return NewNFSProtocolHandler(f.logger), nil
	case "webdav":
		return NewWebDAVProtocolHandler(f.logger), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func (f *ProtocolHandlerFactory) GetSupportedProtocols() []string {
	return []string{"local", "smb", "ftp", "nfs", "webdav"}
}

// ProtocolCapabilities provides information about protocol capabilities
type ProtocolCapabilities struct {
	Protocol                   string        `json:"protocol"`
	SupportsRealTimeNotification bool          `json:"supports_realtime_notification"`
	MoveWindow                 time.Duration `json:"move_window"`
	SupportsAtomicMove         bool          `json:"supports_atomic_move"`
	RequiresPolling           bool          `json:"requires_polling"`
}

func GetProtocolCapabilities(protocol string, logger *zap.Logger) (*ProtocolCapabilities, error) {
	factory := NewProtocolHandlerFactory(logger)
	handler, err := factory.CreateHandler(protocol)
	if err != nil {
		return nil, err
	}

	return &ProtocolCapabilities{
		Protocol:                   protocol,
		SupportsRealTimeNotification: handler.SupportsRealTimeNotification(),
		MoveWindow:                 handler.GetMoveWindow(),
		SupportsAtomicMove:         protocol == "local" || protocol == "nfs",
		RequiresPolling:           !handler.SupportsRealTimeNotification(),
	}, nil
}