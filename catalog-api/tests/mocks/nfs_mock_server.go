package mocks

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockNFSServer provides a mock NFS server for testing
// Note: This is a simplified mock for testing purposes
// Real NFS protocol implementation would be much more complex
type MockNFSServer struct {
	logger    *zap.Logger
	exports   map[string]*MockNFSExport
	files     map[string]*MockNFSFile
	running   bool
	mu        sync.RWMutex
	port      int
	mountPath string
}

// MockNFSExport represents an NFS export
type MockNFSExport struct {
	Path        string
	Description string
	Options     string
	Clients     []string // Allowed client IPs/hostnames
}

// MockNFSFile represents a mock NFS file or directory
type MockNFSFile struct {
	Name        string
	Path        string
	IsDirectory bool
	Size        int64
	ModTime     time.Time
	Content     []byte
	Mode        uint32
	UID         uint32
	GID         uint32
	Inode       uint64
}

// NewMockNFSServer creates a new mock NFS server
func NewMockNFSServer(logger *zap.Logger, mountPath string) *MockNFSServer {
	server := &MockNFSServer{
		logger:    logger,
		exports:   make(map[string]*MockNFSExport),
		files:     make(map[string]*MockNFSFile),
		mountPath: mountPath,
		port:      2049, // Standard NFS port
	}

	server.setupDefaultData()
	return server
}

// setupDefaultData adds default exports and files for testing
func (s *MockNFSServer) setupDefaultData() {
	// Add default exports
	s.AddExport("/export/media", "Media files export", "rw,sync,no_subtree_check", []string{"*"})
	s.AddExport("/export/backup", "Backup files export", "ro,sync,no_subtree_check", []string{"192.168.1.0/24"})
	s.AddExport("/export/shared", "Shared files export", "rw,async,no_root_squash", []string{"localhost", "127.0.0.1"})

	// Add default directory structure for /export/media
	s.AddFile("/export/media", "", "movies", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "", "music", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "", "photos", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "", "readme.txt", false, 512, []byte("Media files repository"), 0644, 1000, 1000)

	s.AddFile("/export/media", "movies", "action", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "movies", "comedy", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "movies/action", "movie1.mp4", false, 1073741824, []byte("Mock movie content"), 0644, 1000, 1000)
	s.AddFile("/export/media", "movies/comedy", "funny.mkv", false, 536870912, []byte("Mock comedy content"), 0644, 1000, 1000)

	s.AddFile("/export/media", "music", "rock", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "music", "jazz", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "music/rock", "song1.mp3", false, 5242880, []byte("Mock rock song"), 0644, 1000, 1000)
	s.AddFile("/export/media", "music/jazz", "smooth.flac", false, 41943040, []byte("Mock jazz song"), 0644, 1000, 1000)

	s.AddFile("/export/media", "photos", "2024", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/media", "photos/2024", "vacation.jpg", false, 2097152, []byte("Mock JPEG content"), 0644, 1000, 1000)
	s.AddFile("/export/media", "photos/2024", "family.png", false, 1048576, []byte("Mock PNG content"), 0644, 1000, 1000)

	// Add files for /export/backup
	s.AddFile("/export/backup", "", "daily", true, 0, []byte{}, 0755, 0, 0)
	s.AddFile("/export/backup", "", "weekly", true, 0, []byte{}, 0755, 0, 0)
	s.AddFile("/export/backup", "", "monthly", true, 0, []byte{}, 0755, 0, 0)

	s.AddFile("/export/backup", "daily", "backup_2024-01-01.tar.gz", false, 104857600, []byte("Mock backup data"), 0644, 0, 0)
	s.AddFile("/export/backup", "daily", "backup_2024-01-02.tar.gz", false, 98765432, []byte("Mock backup data"), 0644, 0, 0)

	s.AddFile("/export/backup", "weekly", "backup_week_01.tar.gz", false, 1073741824, []byte("Mock weekly backup"), 0644, 0, 0)
	s.AddFile("/export/backup", "monthly", "backup_202401.tar.gz", false, 5368709120, []byte("Mock monthly backup"), 0644, 0, 0)

	// Add files for /export/shared
	s.AddFile("/export/shared", "", "documents", true, 0, []byte{}, 0755, 1000, 1000)
	s.AddFile("/export/shared", "", "tmp", true, 0, []byte{}, 0777, 1000, 1000)
	s.AddFile("/export/shared", "", "info.txt", false, 1024, []byte("Shared folder information"), 0644, 1000, 1000)

	s.AddFile("/export/shared", "documents", "manual.pdf", false, 1048576, []byte("Mock PDF manual"), 0644, 1000, 1000)
	s.AddFile("/export/shared", "documents", "config.xml", false, 4096, []byte("<config></config>"), 0644, 1000, 1000)
}

// AddExport adds an NFS export
func (s *MockNFSServer) AddExport(path, description, options string, clients []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.exports[path] = &MockNFSExport{
		Path:        path,
		Description: description,
		Options:     options,
		Clients:     clients,
	}
}

// AddFile adds a file or directory to an export
func (s *MockNFSServer) AddFile(exportPath, parentPath, name string, isDirectory bool, size int64, content []byte, mode, uid, gid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := name
	if parentPath != "" {
		path = parentPath + "/" + name
	}

	fullPath := exportPath + "/" + path
	if parentPath == "" && name != "" {
		fullPath = exportPath + "/" + name
	}

	// Generate inode number
	inode := uint64(len(s.files) + 1)

	s.files[fullPath] = &MockNFSFile{
		Name:        name,
		Path:        fullPath,
		IsDirectory: isDirectory,
		Size:        size,
		ModTime:     time.Now(),
		Content:     content,
		Mode:        mode,
		UID:         uid,
		GID:         gid,
		Inode:       inode,
	}
}

// Start starts the mock NFS server
func (s *MockNFSServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	s.running = true
	s.logger.Info("Mock NFS server started", zap.String("mount_path", s.mountPath))

	// In a real NFS implementation, you would:
	// 1. Start the portmapper service
	// 2. Register NFS services (MOUNT, NFS, etc.)
	// 3. Listen for RPC calls
	// 4. Handle mount requests and file operations

	// For this mock, we just simulate the server being started
	return nil
}

// Stop stops the mock NFS server
func (s *MockNFSServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	s.logger.Info("Mock NFS server stopped")
	return nil
}

// GetPort returns the NFS port (2049)
func (s *MockNFSServer) GetPort() int {
	return s.port
}

// GetMountPath returns the mount path
func (s *MockNFSServer) GetMountPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mountPath
}

// IsRunning returns true if the server is running
func (s *MockNFSServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// ListExports returns all available exports
func (s *MockNFSServer) ListExports() []*MockNFSExport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var exports []*MockNFSExport
	for _, export := range s.exports {
		exports = append(exports, export)
	}

	// Sort by path
	sort.Slice(exports, func(i, j int) bool {
		return exports[i].Path < exports[j].Path
	})

	return exports
}

// Mount simulates mounting an NFS export
func (s *MockNFSServer) Mount(exportPath, clientIP string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	export, exists := s.exports[exportPath]
	if !exists {
		return fmt.Errorf("export not found: %s", exportPath)
	}

	// Check if client is allowed
	allowed := false
	for _, allowedClient := range export.Clients {
		if allowedClient == "*" || allowedClient == clientIP || allowedClient == "localhost" {
			allowed = true
			break
		}
		// Simple subnet check for CIDR notation
		if strings.Contains(allowedClient, "/") && strings.HasPrefix(clientIP, strings.Split(allowedClient, "/")[0][:3]) {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("client %s not allowed to mount %s", clientIP, exportPath)
	}

	s.logger.Info("NFS mount successful", zap.String("export", exportPath), zap.String("client", clientIP))
	return nil
}

// Unmount simulates unmounting an NFS export
func (s *MockNFSServer) Unmount(exportPath, clientIP string) error {
	s.logger.Info("NFS unmount", zap.String("export", exportPath), zap.String("client", clientIP))
	return nil
}

// ListFiles lists files in an export path
func (s *MockNFSServer) ListFiles(exportPath, dirPath string) ([]*MockNFSFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if export exists
	if _, exists := s.exports[exportPath]; !exists {
		return nil, fmt.Errorf("export not found: %s", exportPath)
	}

	fullPath := exportPath
	if dirPath != "" && dirPath != "." {
		fullPath = exportPath + "/" + dirPath
	}

	var files []*MockNFSFile

	for filePath, file := range s.files {
		// Check if file is in the requested directory
		if dirPath == "" || dirPath == "." {
			// List files directly in the export
			expectedPrefix := exportPath + "/"
			if strings.HasPrefix(filePath, expectedPrefix) {
				relativePath := strings.TrimPrefix(filePath, expectedPrefix)
				if !strings.Contains(relativePath, "/") && relativePath != "" {
					files = append(files, file)
				}
			}
		} else {
			// List files in a specific subdirectory
			expectedPrefix := fullPath + "/"
			if strings.HasPrefix(filePath, expectedPrefix) {
				relativePath := strings.TrimPrefix(filePath, expectedPrefix)
				if !strings.Contains(relativePath, "/") && relativePath != "" {
					files = append(files, file)
				}
			}
		}
	}

	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// GetFile retrieves a specific file
func (s *MockNFSServer) GetFile(exportPath, filePath string) (*MockNFSFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if export exists
	if _, exists := s.exports[exportPath]; !exists {
		return nil, fmt.Errorf("export not found: %s", exportPath)
	}

	fullPath := exportPath + "/" + filePath
	if filePath == "" || filePath == "." {
		fullPath = exportPath
	}

	file, exists := s.files[fullPath]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", fullPath)
	}

	return file, nil
}

// WriteFile writes content to a file (if export allows writing)
func (s *MockNFSServer) WriteFile(exportPath, filePath string, content []byte, mode, uid, gid uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if export exists and is writable
	export, exists := s.exports[exportPath]
	if !exists {
		return fmt.Errorf("export not found: %s", exportPath)
	}

	if !strings.Contains(export.Options, "rw") {
		return fmt.Errorf("export is read-only: %s", exportPath)
	}

	fullPath := exportPath + "/" + filePath
	fileName := filepath.Base(filePath)

	// Generate inode number
	inode := uint64(len(s.files) + 1)

	s.files[fullPath] = &MockNFSFile{
		Name:        fileName,
		Path:        fullPath,
		IsDirectory: false,
		Size:        int64(len(content)),
		ModTime:     time.Now(),
		Content:     content,
		Mode:        mode,
		UID:         uid,
		GID:         gid,
		Inode:       inode,
	}

	return nil
}

// DeleteFile deletes a file (if export allows writing)
func (s *MockNFSServer) DeleteFile(exportPath, filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if export exists and is writable
	export, exists := s.exports[exportPath]
	if !exists {
		return fmt.Errorf("export not found: %s", exportPath)
	}

	if !strings.Contains(export.Options, "rw") {
		return fmt.Errorf("export is read-only: %s", exportPath)
	}

	fullPath := exportPath + "/" + filePath
	if _, exists := s.files[fullPath]; !exists {
		return fmt.Errorf("file not found: %s", fullPath)
	}

	delete(s.files, fullPath)
	return nil
}

// CreateDirectory creates a directory (if export allows writing)
func (s *MockNFSServer) CreateDirectory(exportPath, dirPath string, mode, uid, gid uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if export exists and is writable
	export, exists := s.exports[exportPath]
	if !exists {
		return fmt.Errorf("export not found: %s", exportPath)
	}

	if !strings.Contains(export.Options, "rw") {
		return fmt.Errorf("export is read-only: %s", exportPath)
	}

	fullPath := exportPath + "/" + dirPath
	dirName := filepath.Base(dirPath)

	// Generate inode number
	inode := uint64(len(s.files) + 1)

	s.files[fullPath] = &MockNFSFile{
		Name:        dirName,
		Path:        fullPath,
		IsDirectory: true,
		Size:        0,
		ModTime:     time.Now(),
		Content:     []byte{},
		Mode:        mode,
		UID:         uid,
		GID:         gid,
		Inode:       inode,
	}

	return nil
}

// GetFileCount returns the number of files/directories
func (s *MockNFSServer) GetFileCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.files)
}

// GetExportCount returns the number of exports
func (s *MockNFSServer) GetExportCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.exports)
}

// GetExportNames returns list of export paths
func (s *MockNFSServer) GetExportNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var names []string
	for path := range s.exports {
		names = append(names, path)
	}

	sort.Strings(names)
	return names
}

// TestConnection tests if a client can connect to an export
func (s *MockNFSServer) TestConnection(exportPath, clientIP string) error {
	return s.Mount(exportPath, clientIP)
}
