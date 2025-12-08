package mocks

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockSMBServer provides a mock SMB server for testing
type MockSMBServer struct {
	logger   *zap.Logger
	listener net.Listener
	port     int
	shares   map[string]*MockSMBShare
	users    map[string]string // username -> password
	running  bool
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// MockSMBShare represents a mock SMB share
type MockSMBShare struct {
	Name        string
	Description string
	Files       map[string]*MockSMBFile
	Writable    bool
}

// MockSMBFile represents a mock file or directory
type MockSMBFile struct {
	Name        string
	IsDirectory bool
	Size        int64
	ModTime     time.Time
	Content     []byte
}

// NewMockSMBServer creates a new mock SMB server
func NewMockSMBServer(logger *zap.Logger) *MockSMBServer {
	server := &MockSMBServer{
		logger: logger,
		shares: make(map[string]*MockSMBShare),
		users:  make(map[string]string),
	}

	// Add default shares and users
	server.setupDefaultData()

	return server
}

// setupDefaultData adds default shares, users, and files for testing
func (s *MockSMBServer) setupDefaultData() {
	// Add default users
	s.users["guest"] = ""
	s.users["testuser"] = "testpass"
	s.users["admin"] = "adminpass"

	// Add default shares
	s.AddShare("shared", "Shared folder", true)
	s.AddShare("public", "Public folder", false)
	s.AddShare("media", "Media files", true)
	s.AddShare("backup", "Backup files", false)

	// Add sample files to shares
	s.AddFile("shared", "", "documents", true, 0, []byte{})
	s.AddFile("shared", "", "readme.txt", false, 1024, []byte("Welcome to the shared folder!"))
	s.AddFile("shared", "documents", "report.doc", false, 2048, []byte("Sample document content"))

	s.AddFile("media", "", "videos", true, 0, []byte{})
	s.AddFile("media", "", "music", true, 0, []byte{})
	s.AddFile("media", "videos", "sample.mp4", false, 1048576, []byte("Mock video content"))
	s.AddFile("media", "music", "song.mp3", false, 524288, []byte("Mock audio content"))

	s.AddFile("public", "", "info.txt", false, 512, []byte("Public information"))
	s.AddFile("public", "", "downloads", true, 0, []byte{})

	s.AddFile("backup", "", "backup_2024.zip", false, 10485760, []byte("Mock backup data"))
}

// AddShare adds a share to the mock server
func (s *MockSMBServer) AddShare(name, description string, writable bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shares[name] = &MockSMBShare{
		Name:        name,
		Description: description,
		Files:       make(map[string]*MockSMBFile),
		Writable:    writable,
	}
}

// AddFile adds a file or directory to a share
func (s *MockSMBServer) AddFile(shareName, parentPath, fileName string, isDirectory bool, size int64, content []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	share, exists := s.shares[shareName]
	if !exists {
		return
	}

	path := fileName
	if parentPath != "" {
		path = parentPath + "/" + fileName
	}

	share.Files[path] = &MockSMBFile{
		Name:        fileName,
		IsDirectory: isDirectory,
		Size:        size,
		ModTime:     time.Now(),
		Content:     content,
	}
}

// AddUser adds a user with password
func (s *MockSMBServer) AddUser(username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[username] = password
}

// Start starts the mock SMB server
func (s *MockSMBServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	// Listen on any available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port
	s.running = true

	s.logger.Info("Mock SMB server started", zap.Int("port", s.port))

	// Start accepting connections
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// Stop stops the mock SMB server
func (s *MockSMBServer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}

	s.wg.Wait()
	s.logger.Info("Mock SMB server stopped")
	return nil
}

// GetPort returns the port the server is listening on
func (s *MockSMBServer) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// GetAddress returns the server address
func (s *MockSMBServer) GetAddress() string {
	return fmt.Sprintf("localhost:%d", s.GetPort())
}

// acceptConnections handles incoming connections
func (s *MockSMBServer) acceptConnections() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.RLock()
			running := s.running
			s.mu.RUnlock()

			if !running {
				return // Server stopped
			}
			s.logger.Error("Failed to accept connection", zap.Error(err))
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single SMB connection
func (s *MockSMBServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	s.logger.Debug("New SMB connection", zap.String("remote", conn.RemoteAddr().String()))

	// This is a simplified mock implementation
	// In a real scenario, you'd implement the full SMB protocol

	// For testing purposes, we'll simulate basic SMB responses
	// This allows the go-smb2 client to connect and get mock data

	// Read initial request
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		s.logger.Debug("Failed to read from connection", zap.Error(err))
		return
	}

	request := buffer[:n]
	s.logger.Debug("Received SMB request", zap.Int("bytes", n))

	// Send mock response (simplified)
	response := s.generateMockResponse(request)
	_, err = conn.Write(response)
	if err != nil {
		s.logger.Debug("Failed to write response", zap.Error(err))
		return
	}

	s.logger.Debug("Sent SMB response", zap.Int("bytes", len(response)))
}

// generateMockResponse generates a mock SMB response
func (s *MockSMBServer) generateMockResponse(request []byte) []byte {
	// This is a mock implementation that handles basic SMB2 negotiation
	// We need to respond properly to SMB2 negotiate and session setup commands

	// Check if this is a negotiate request
	if len(request) >= 64 {
		// Check for SMB2 negotiate command (0x0000)
		if request[16] == 0x00 && request[17] == 0x00 {
			return s.generateNegotiateResponse()
		}
		// Check for session setup command (0x0001)
		if request[16] == 0x00 && request[17] == 0x01 {
			return s.generateSessionSetupResponse()
		}
		// Check for tree connect command (0x0003)
		if request[16] == 0x00 && request[17] == 0x03 {
			return s.generateTreeConnectResponse()
		}
		// Check for create command (0x0005)
		if request[16] == 0x00 && request[17] == 0x05 {
			return s.generateCreateResponse()
		}
		// Check for query directory command (0x0006)
		if request[16] == 0x00 && request[17] == 0x06 {
			return s.generateQueryDirectoryResponse()
		}
		// Check for tree disconnect command (0x0004)
		if request[16] == 0x00 && request[17] == 0x04 {
			return s.generateTreeDisconnectResponse()
		}
		// Check for logoff command (0x0002)
		if request[16] == 0x00 && request[17] == 0x02 {
			return s.generateLogoffResponse()
		}
	}

	// Default response for unknown commands
	response := make([]byte, 64)
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	return response
}

// generateNegotiateResponse generates a mock SMB2 negotiate response
func (s *MockSMBServer) generateNegotiateResponse() []byte {
	response := make([]byte, 64)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Add more response fields as needed
	return response
}

// generateSessionSetupResponse generates a mock session setup response
func (s *MockSMBServer) generateSessionSetupResponse() []byte {
	response := make([]byte, 64)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Add more response fields as needed
	return response
}

// generateTreeConnectResponse generates a mock tree connect response
func (s *MockSMBServer) generateTreeConnectResponse() []byte {
	response := make([]byte, 64)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Add more response fields as needed
	return response
}

// generateCreateResponse generates a mock create response
func (s *MockSMBServer) generateCreateResponse() []byte {
	response := make([]byte, 64)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Add more response fields as needed
	return response
}

// generateQueryDirectoryResponse generates a mock query directory response
func (s *MockSMBServer) generateQueryDirectoryResponse() []byte {
	response := make([]byte, 256)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Mock directory entries would go here
	// For now, just return a successful response
	return response
}

// generateTreeDisconnectResponse generates a mock tree disconnect response
func (s *MockSMBServer) generateTreeDisconnectResponse() []byte {
	response := make([]byte, 64)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Add more response fields as needed
	return response
}

// generateLogoffResponse generates a mock logoff response
func (s *MockSMBServer) generateLogoffResponse() []byte {
	response := make([]byte, 64)
	// SMB2 Header
	copy(response[0:4], []byte{0xFE, 0x53, 0x4D, 0x42}) // SMB2 signature
	response[16] = 0x00                                 // Credit charge
	response[17] = 0x00                                 // Status: SUCCESS
	// Add more response fields as needed
	return response
}

// AuthenticateUser checks if user credentials are valid
func (s *MockSMBServer) AuthenticateUser(username, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expectedPassword, exists := s.users[username]
	if !exists {
		return false
	}

	return expectedPassword == password
}

// ListShares returns available shares
func (s *MockSMBServer) ListShares() []MockSMBShare {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var shares []MockSMBShare
	for _, share := range s.shares {
		shares = append(shares, *share)
	}
	return shares
}

// ListFiles returns files in a share path
func (s *MockSMBServer) ListFiles(shareName, path string) ([]*MockSMBFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	share, exists := s.shares[shareName]
	if !exists {
		return nil, fmt.Errorf("share not found: %s", shareName)
	}

	var files []*MockSMBFile

	// List files in the specified path
	for filePath, file := range share.Files {
		// Check if file is in the requested path
		if path == "" || path == "." {
			// Root path - include files with no parent path
			if !strings.Contains(filePath, "/") {
				files = append(files, file)
			}
		} else {
			// Specific path - include files that start with the path
			if strings.HasPrefix(filePath, path+"/") {
				relativePath := strings.TrimPrefix(filePath, path+"/")
				if !strings.Contains(relativePath, "/") {
					// Direct child of the path
					files = append(files, file)
				}
			}
		}
	}

	return files, nil
}

// GetFile returns file content
func (s *MockSMBServer) GetFile(shareName, filePath string) (*MockSMBFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	share, exists := s.shares[shareName]
	if !exists {
		return nil, fmt.Errorf("share not found: %s", shareName)
	}

	file, exists := share.Files[filePath]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	return file, nil
}

// WriteFile writes content to a file (if share is writable)
func (s *MockSMBServer) WriteFile(shareName, filePath string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	share, exists := s.shares[shareName]
	if !exists {
		return fmt.Errorf("share not found: %s", shareName)
	}

	if !share.Writable {
		return fmt.Errorf("share is read-only: %s", shareName)
	}

	// Extract filename from path
	fileName := filePath
	if strings.Contains(filePath, "/") {
		parts := strings.Split(filePath, "/")
		fileName = parts[len(parts)-1]
	}

	share.Files[filePath] = &MockSMBFile{
		Name:        fileName,
		IsDirectory: false,
		Size:        int64(len(content)),
		ModTime:     time.Now(),
		Content:     content,
	}

	return nil
}

// DeleteFile deletes a file (if share is writable)
func (s *MockSMBServer) DeleteFile(shareName, filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	share, exists := s.shares[shareName]
	if !exists {
		return fmt.Errorf("share not found: %s", shareName)
	}

	if !share.Writable {
		return fmt.Errorf("share is read-only: %s", shareName)
	}

	delete(share.Files, filePath)
	return nil
}

// GetShareNames returns list of share names
func (s *MockSMBServer) GetShareNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var names []string
	for name := range s.shares {
		names = append(names, name)
	}
	return names
}

// IsRunning returns true if the server is running
func (s *MockSMBServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
