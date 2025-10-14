package mocks

import (
	"bufio"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockFTPServer provides a mock FTP server for testing
type MockFTPServer struct {
	logger      *zap.Logger
	listener    net.Listener
	port        int
	users       map[string]string // username -> password
	files       map[string]*MockFTPFile
	running     bool
	wg          sync.WaitGroup
	mu          sync.RWMutex
	currentDir  string
	passiveMode bool
	dataPort    int
}

// MockFTPFile represents a mock FTP file or directory
type MockFTPFile struct {
	Name        string
	Path        string
	IsDirectory bool
	Size        int64
	ModTime     time.Time
	Content     []byte
	Permissions string
}

// NewMockFTPServer creates a new mock FTP server
func NewMockFTPServer(logger *zap.Logger) *MockFTPServer {
	server := &MockFTPServer{
		logger:     logger,
		users:      make(map[string]string),
		files:      make(map[string]*MockFTPFile),
		currentDir: "/",
	}

	server.setupDefaultData()
	return server
}

// setupDefaultData adds default users and files for testing
func (s *MockFTPServer) setupDefaultData() {
	// Add default users
	s.users["anonymous"] = ""
	s.users["testuser"] = "testpass"
	s.users["ftpuser"] = "ftppass"

	// Add default directory structure
	s.AddFile("/", "public", true, 0, []byte{}, "drwxr-xr-x")
	s.AddFile("/", "uploads", true, 0, []byte{}, "drwxrwxrwx")
	s.AddFile("/", "readme.txt", false, 512, []byte("Welcome to the FTP server!"), "-rw-r--r--")

	s.AddFile("/public", "documents", true, 0, []byte{}, "drwxr-xr-x")
	s.AddFile("/public", "software", true, 0, []byte{}, "drwxr-xr-x")
	s.AddFile("/public", "info.txt", false, 1024, []byte("Public information file"), "-rw-r--r--")

	s.AddFile("/public/documents", "manual.pdf", false, 204800, []byte("Mock PDF content"), "-rw-r--r--")
	s.AddFile("/public/documents", "guide.doc", false, 102400, []byte("Mock document content"), "-rw-r--r--")

	s.AddFile("/public/software", "installer.exe", false, 5242880, []byte("Mock installer content"), "-rw-r--r--")
	s.AddFile("/public/software", "update.zip", false, 1048576, []byte("Mock update content"), "-rw-r--r--")

	s.AddFile("/uploads", "temp", true, 0, []byte{}, "drwxrwxrwx")
}

// AddFile adds a file or directory to the server
func (s *MockFTPServer) AddFile(parentPath, name string, isDirectory bool, size int64, content []byte, permissions string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(parentPath, name)
	if parentPath == "/" {
		path = "/" + name
	}

	s.files[path] = &MockFTPFile{
		Name:        name,
		Path:        path,
		IsDirectory: isDirectory,
		Size:        size,
		ModTime:     time.Now(),
		Content:     content,
		Permissions: permissions,
	}
}

// AddUser adds a user with password
func (s *MockFTPServer) AddUser(username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[username] = password
}

// Start starts the mock FTP server
func (s *MockFTPServer) Start() error {
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

	s.logger.Info("Mock FTP server started", zap.Int("port", s.port))

	// Start accepting connections
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// Stop stops the mock FTP server
func (s *MockFTPServer) Stop() error {
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
	s.logger.Info("Mock FTP server stopped")
	return nil
}

// GetPort returns the port the server is listening on
func (s *MockFTPServer) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// GetAddress returns the server address
func (s *MockFTPServer) GetAddress() string {
	return fmt.Sprintf("localhost:%d", s.GetPort())
}

// acceptConnections handles incoming connections
func (s *MockFTPServer) acceptConnections() {
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
			s.logger.Error("Failed to accept FTP connection", zap.Error(err))
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single FTP connection
func (s *MockFTPServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	s.logger.Debug("New FTP connection", zap.String("remote", conn.RemoteAddr().String()))

	// Send welcome message
	s.sendResponse(conn, "220 Mock FTP Server Ready")

	scanner := bufio.NewScanner(conn)
	authenticated := false
	username := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		s.logger.Debug("FTP command received", zap.String("command", line))

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch command {
		case "USER":
			username = args
			if s.userExists(username) {
				s.sendResponse(conn, "331 Password required for "+username)
			} else {
				s.sendResponse(conn, "530 User not found")
			}

		case "PASS":
			if s.authenticateUser(username, args) {
				authenticated = true
				s.sendResponse(conn, "230 User logged in")
			} else {
				s.sendResponse(conn, "530 Authentication failed")
			}

		case "SYST":
			s.sendResponse(conn, "215 UNIX Type: L8")

		case "PWD":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.sendResponse(conn, fmt.Sprintf("257 \"%s\" is current directory", s.currentDir))

		case "CWD":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			if s.changeDirectory(args) {
				s.sendResponse(conn, "250 Directory changed")
			} else {
				s.sendResponse(conn, "550 Directory not found")
			}

		case "LIST":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleList(conn, args)

		case "NLST":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleNlst(conn, args)

		case "SIZE":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleSize(conn, args)

		case "MDTM":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleMdtm(conn, args)

		case "TYPE":
			s.sendResponse(conn, "200 Type set to "+args)

		case "PASV":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handlePasv(conn)

		case "RETR":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleRetr(conn, args)

		case "STOR":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleStor(conn, args)

		case "DELE":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleDele(conn, args)

		case "MKD", "XMKD":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleMkd(conn, args)

		case "RMD", "XRMD":
			if !authenticated {
				s.sendResponse(conn, "530 Not logged in")
				continue
			}
			s.handleRmd(conn, args)

		case "QUIT":
			s.sendResponse(conn, "221 Goodbye")
			return

		case "NOOP":
			s.sendResponse(conn, "200 OK")

		default:
			s.sendResponse(conn, "502 Command not implemented")
		}
	}
}

// sendResponse sends an FTP response
func (s *MockFTPServer) sendResponse(conn net.Conn, response string) {
	_, err := conn.Write([]byte(response + "\r\n"))
	if err != nil {
		s.logger.Debug("Failed to send FTP response", zap.Error(err))
	}
	s.logger.Debug("FTP response sent", zap.String("response", response))
}

// userExists checks if a user exists
func (s *MockFTPServer) userExists(username string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.users[username]
	return exists
}

// authenticateUser checks if user credentials are valid
func (s *MockFTPServer) authenticateUser(username, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expectedPassword, exists := s.users[username]
	if !exists {
		return false
	}

	return expectedPassword == password
}

// changeDirectory changes the current directory
func (s *MockFTPServer) changeDirectory(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetPath := path
	if !strings.HasPrefix(path, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + path
		} else {
			targetPath = s.currentDir + "/" + path
		}
	}

	// Check if directory exists
	if file, exists := s.files[targetPath]; exists && file.IsDirectory {
		s.currentDir = targetPath
		return true
	}

	return false
}

// handleList handles the LIST command
func (s *MockFTPServer) handleList(conn net.Conn, path string) {
	targetPath := s.currentDir
	if path != "" {
		if strings.HasPrefix(path, "/") {
			targetPath = path
		} else {
			if s.currentDir == "/" {
				targetPath = "/" + path
			} else {
				targetPath = s.currentDir + "/" + path
			}
		}
	}

	// For simplicity, send list directly on command channel
	// In real FTP, this would use a data channel
	s.sendResponse(conn, "150 Opening data connection")

	files := s.listFiles(targetPath)
	for _, file := range files {
		listing := s.formatFileListing(file)
		s.sendResponse(conn, listing)
	}

	s.sendResponse(conn, "226 Transfer complete")
}

// handleNlst handles the NLST command (name list)
func (s *MockFTPServer) handleNlst(conn net.Conn, path string) {
	targetPath := s.currentDir
	if path != "" {
		if strings.HasPrefix(path, "/") {
			targetPath = path
		} else {
			if s.currentDir == "/" {
				targetPath = "/" + path
			} else {
				targetPath = s.currentDir + "/" + path
			}
		}
	}

	s.sendResponse(conn, "150 Opening data connection")

	files := s.listFiles(targetPath)
	for _, file := range files {
		s.sendResponse(conn, file.Name)
	}

	s.sendResponse(conn, "226 Transfer complete")
}

// handleSize handles the SIZE command
func (s *MockFTPServer) handleSize(conn net.Conn, path string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	targetPath := path
	if !strings.HasPrefix(path, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + path
		} else {
			targetPath = s.currentDir + "/" + path
		}
	}

	if file, exists := s.files[targetPath]; exists && !file.IsDirectory {
		s.sendResponse(conn, fmt.Sprintf("213 %d", file.Size))
	} else {
		s.sendResponse(conn, "550 File not found")
	}
}

// handleMdtm handles the MDTM command (modification time)
func (s *MockFTPServer) handleMdtm(conn net.Conn, path string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	targetPath := path
	if !strings.HasPrefix(path, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + path
		} else {
			targetPath = s.currentDir + "/" + path
		}
	}

	if file, exists := s.files[targetPath]; exists {
		mdtm := file.ModTime.Format("20060102150405")
		s.sendResponse(conn, fmt.Sprintf("213 %s", mdtm))
	} else {
		s.sendResponse(conn, "550 File not found")
	}
}

// handlePasv handles passive mode
func (s *MockFTPServer) handlePasv(conn net.Conn) {
	// For simplicity, just acknowledge passive mode
	s.passiveMode = true
	s.dataPort = s.port + 1

	// Format: 227 Entering Passive Mode (h1,h2,h3,h4,p1,p2)
	// where IP is h1.h2.h3.h4 and port is p1*256+p2
	p1 := s.dataPort / 256
	p2 := s.dataPort % 256

	s.sendResponse(conn, fmt.Sprintf("227 Entering Passive Mode (127,0,0,1,%d,%d)", p1, p2))
}

// handleRetr handles file retrieval
func (s *MockFTPServer) handleRetr(conn net.Conn, filename string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	targetPath := filename
	if !strings.HasPrefix(filename, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + filename
		} else {
			targetPath = s.currentDir + "/" + filename
		}
	}

	if file, exists := s.files[targetPath]; exists && !file.IsDirectory {
		s.sendResponse(conn, "150 Opening data connection for file transfer")
		// In real FTP, content would be sent over data channel
		s.sendResponse(conn, "226 Transfer complete")
	} else {
		s.sendResponse(conn, "550 File not found")
	}
}

// handleStor handles file storage
func (s *MockFTPServer) handleStor(conn net.Conn, filename string) {
	targetPath := filename
	if !strings.HasPrefix(filename, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + filename
		} else {
			targetPath = s.currentDir + "/" + filename
		}
	}

	s.sendResponse(conn, "150 Opening data connection for file upload")

	// Simulate file upload
	s.AddFile(filepath.Dir(targetPath), filepath.Base(targetPath), false, 1024, []byte("Uploaded content"), "-rw-r--r--")

	s.sendResponse(conn, "226 Transfer complete")
}

// handleDele handles file deletion
func (s *MockFTPServer) handleDele(conn net.Conn, filename string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetPath := filename
	if !strings.HasPrefix(filename, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + filename
		} else {
			targetPath = s.currentDir + "/" + filename
		}
	}

	if _, exists := s.files[targetPath]; exists {
		delete(s.files, targetPath)
		s.sendResponse(conn, "250 File deleted")
	} else {
		s.sendResponse(conn, "550 File not found")
	}
}

// handleMkd handles directory creation
func (s *MockFTPServer) handleMkd(conn net.Conn, dirname string) {
	targetPath := dirname
	if !strings.HasPrefix(dirname, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + dirname
		} else {
			targetPath = s.currentDir + "/" + dirname
		}
	}

	s.AddFile(filepath.Dir(targetPath), filepath.Base(targetPath), true, 0, []byte{}, "drwxr-xr-x")
	s.sendResponse(conn, fmt.Sprintf("257 \"%s\" directory created", targetPath))
}

// handleRmd handles directory removal
func (s *MockFTPServer) handleRmd(conn net.Conn, dirname string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetPath := dirname
	if !strings.HasPrefix(dirname, "/") {
		if s.currentDir == "/" {
			targetPath = "/" + dirname
		} else {
			targetPath = s.currentDir + "/" + dirname
		}
	}

	if file, exists := s.files[targetPath]; exists && file.IsDirectory {
		delete(s.files, targetPath)
		s.sendResponse(conn, "250 Directory removed")
	} else {
		s.sendResponse(conn, "550 Directory not found")
	}
}

// listFiles lists files in a directory
func (s *MockFTPServer) listFiles(path string) []*MockFTPFile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var files []*MockFTPFile

	for filePath, file := range s.files {
		if filePath == path {
			continue // Skip the directory itself
		}

		// Check if file is a direct child of the path
		expectedPrefix := path
		if path == "/" {
			expectedPrefix = "/"
		} else {
			expectedPrefix = path + "/"
		}

		if strings.HasPrefix(filePath, expectedPrefix) {
			relativePath := strings.TrimPrefix(filePath, expectedPrefix)
			if !strings.Contains(relativePath, "/") {
				// Direct child
				files = append(files, file)
			}
		}
	}

	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files
}

// formatFileListing formats a file for LIST command output
func (s *MockFTPServer) formatFileListing(file *MockFTPFile) string {
	// Format: permissions links owner group size month day time filename
	// Example: -rw-r--r--   1 user  group     1024 Jan 01 12:00 filename.txt
	modTime := file.ModTime.Format("Jan 02 15:04")
	return fmt.Sprintf("%s   1 user  group  %8d %s %s",
		file.Permissions, file.Size, modTime, file.Name)
}

// IsRunning returns true if the server is running
func (s *MockFTPServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetFileCount returns the number of files/directories
func (s *MockFTPServer) GetFileCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.files)
}

// GetUserCount returns the number of users
func (s *MockFTPServer) GetUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}
