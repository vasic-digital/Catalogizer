package mocks

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockWebDAVServer provides a mock WebDAV server for testing
type MockWebDAVServer struct {
	logger   *zap.Logger
	server   *http.Server
	port     int
	users    map[string]string // username -> password
	files    map[string]*MockWebDAVFile
	running  bool
	mu       sync.RWMutex
	basePath string
}

// MockWebDAVFile represents a mock WebDAV file or collection
type MockWebDAVFile struct {
	Name        string
	Path        string
	IsDirectory bool
	Size        int64
	ModTime     time.Time
	Content     []byte
	ETag        string
	ContentType string
}

// NewMockWebDAVServer creates a new mock WebDAV server
func NewMockWebDAVServer(logger *zap.Logger) *MockWebDAVServer {
	server := &MockWebDAVServer{
		logger:   logger,
		users:    make(map[string]string),
		files:    make(map[string]*MockWebDAVFile),
		basePath: "/dav",
	}

	server.setupDefaultData()
	return server
}

// setupDefaultData adds default users and files for testing
func (s *MockWebDAVServer) setupDefaultData() {
	// Add default users
	s.users["webdavuser"] = "webdavpass"
	s.users["testuser"] = "testpass"
	s.users["admin"] = "adminpass"

	// Add default directory structure
	s.AddFile("/", "documents", true, 0, []byte{}, "text/html")
	s.AddFile("/", "media", true, 0, []byte{}, "text/html")
	s.AddFile("/", "public", true, 0, []byte{}, "text/html")
	s.AddFile("/", "welcome.txt", false, 1024, []byte("Welcome to WebDAV server!"), "text/plain")

	s.AddFile("/documents", "reports", true, 0, []byte{}, "text/html")
	s.AddFile("/documents", "templates", true, 0, []byte{}, "text/html")
	s.AddFile("/documents", "readme.md", false, 2048, []byte("# Document Repository\n\nThis is the document repository."), "text/markdown")

	s.AddFile("/documents/reports", "quarterly.pdf", false, 512000, []byte("Mock PDF content"), "application/pdf")
	s.AddFile("/documents/reports", "monthly.xlsx", false, 256000, []byte("Mock Excel content"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	s.AddFile("/documents/templates", "letter.docx", false, 128000, []byte("Mock Word template"), "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	s.AddFile("/documents/templates", "invoice.html", false, 8192, []byte("<html><body>Invoice Template</body></html>"), "text/html")

	s.AddFile("/media", "images", true, 0, []byte{}, "text/html")
	s.AddFile("/media", "videos", true, 0, []byte{}, "text/html")
	s.AddFile("/media", "audio", true, 0, []byte{}, "text/html")

	s.AddFile("/media/images", "logo.png", false, 32768, []byte("Mock PNG content"), "image/png")
	s.AddFile("/media/images", "banner.jpg", false, 65536, []byte("Mock JPEG content"), "image/jpeg")

	s.AddFile("/media/videos", "demo.mp4", false, 10485760, []byte("Mock video content"), "video/mp4")
	s.AddFile("/media/audio", "background.mp3", false, 5242880, []byte("Mock audio content"), "audio/mpeg")

	s.AddFile("/public", "info.html", false, 4096, []byte("<html><body><h1>Public Information</h1></body></html>"), "text/html")
	s.AddFile("/public", "downloads", true, 0, []byte{}, "text/html")
}

// AddFile adds a file or collection to the server
func (s *MockWebDAVServer) AddFile(parentPath, name string, isDirectory bool, size int64, content []byte, contentType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(parentPath, name)
	if parentPath == "/" {
		path = "/" + name
	}

	// Generate ETag
	etag := fmt.Sprintf(`"%x-%d"`, time.Now().Unix(), size)

	s.files[path] = &MockWebDAVFile{
		Name:        name,
		Path:        path,
		IsDirectory: isDirectory,
		Size:        size,
		ModTime:     time.Now(),
		Content:     content,
		ETag:        etag,
		ContentType: contentType,
	}
}

// AddUser adds a user with password
func (s *MockWebDAVServer) AddUser(username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[username] = password
}

// Start starts the mock WebDAV server
func (s *MockWebDAVServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:    ":0",
		Handler: mux,
	}

	// Start server in a goroutine to get the port
	listener, err := s.server.Listen()
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.port = listener.Addr().(*net.TCPAddr).Port
	s.running = true

	go func() {
		err := s.server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("WebDAV server error", zap.Error(err))
		}
	}()

	s.logger.Info("Mock WebDAV server started", zap.Int("port", s.port))
	return nil
}

// Stop stops the mock WebDAV server
func (s *MockWebDAVServer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	if s.server != nil {
		err := s.server.Close()
		if err != nil {
			return err
		}
	}

	s.logger.Info("Mock WebDAV server stopped")
	return nil
}

// GetPort returns the port the server is listening on
func (s *MockWebDAVServer) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

// GetAddress returns the server address
func (s *MockWebDAVServer) GetAddress() string {
	return fmt.Sprintf("http://localhost:%d%s", s.GetPort(), s.basePath)
}

// handleRequest handles HTTP requests
func (s *MockWebDAVServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("WebDAV request", zap.String("method", r.Method), zap.String("path", r.URL.Path))

	// Basic authentication
	if !s.authenticate(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="WebDAV"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Remove base path prefix
	path := strings.TrimPrefix(r.URL.Path, s.basePath)
	if path == "" {
		path = "/"
	}

	switch r.Method {
	case "OPTIONS":
		s.handleOptions(w, r, path)
	case "GET":
		s.handleGet(w, r, path)
	case "HEAD":
		s.handleHead(w, r, path)
	case "PROPFIND":
		s.handlePropfind(w, r, path)
	case "PUT":
		s.handlePut(w, r, path)
	case "DELETE":
		s.handleDelete(w, r, path)
	case "MKCOL":
		s.handleMkcol(w, r, path)
	case "COPY":
		s.handleCopy(w, r, path)
	case "MOVE":
		s.handleMove(w, r, path)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// authenticate performs basic authentication
func (s *MockWebDAVServer) authenticate(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	expectedPassword, exists := s.users[username]
	return exists && expectedPassword == password
}

// handleOptions handles OPTIONS requests
func (s *MockWebDAVServer) handleOptions(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Allow", "OPTIONS, GET, HEAD, POST, PUT, DELETE, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE")
	w.Header().Set("DAV", "1, 2")
	w.WriteHeader(http.StatusOK)
}

// handleGet handles GET requests
func (s *MockWebDAVServer) handleGet(w http.ResponseWriter, r *http.Request, path string) {
	s.mu.RLock()
	file, exists := s.files[path]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if file.IsDirectory {
		// Generate HTML listing for directories
		s.handleDirectoryListing(w, r, path)
		return
	}

	// Serve file content
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	w.Header().Set("ETag", file.ETag)
	w.Header().Set("Last-Modified", file.ModTime.Format(time.RFC1123))

	w.WriteHeader(http.StatusOK)
	w.Write(file.Content)
}

// handleHead handles HEAD requests
func (s *MockWebDAVServer) handleHead(w http.ResponseWriter, r *http.Request, path string) {
	s.mu.RLock()
	file, exists := s.files[path]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	w.Header().Set("ETag", file.ETag)
	w.Header().Set("Last-Modified", file.ModTime.Format(time.RFC1123))

	w.WriteHeader(http.StatusOK)
}

// handleDirectoryListing generates HTML directory listing
func (s *MockWebDAVServer) handleDirectoryListing(w http.ResponseWriter, r *http.Request, path string) {
	files := s.listFiles(path)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "<html><head><title>Directory: %s</title></head><body>", path)
	fmt.Fprintf(w, "<h1>Directory: %s</h1><ul>", path)

	if path != "/" {
		fmt.Fprintf(w, `<li><a href="../">../</a></li>`)
	}

	for _, file := range files {
		name := file.Name
		if file.IsDirectory {
			name += "/"
		}
		fmt.Fprintf(w, `<li><a href="%s">%s</a> (%d bytes)</li>`, name, name, file.Size)
	}

	fmt.Fprintf(w, "</ul></body></html>")
}

// handlePropfind handles PROPFIND requests
func (s *MockWebDAVServer) handlePropfind(w http.ResponseWriter, r *http.Request, path string) {
	depth := r.Header.Get("Depth")
	if depth == "" {
		depth = "1"
	}

	s.mu.RLock()
	file, exists := s.files[path]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	// Generate WebDAV XML response
	fmt.Fprintf(w, `<?xml version="1.0" encoding="utf-8"?>`)
	fmt.Fprintf(w, `<D:multistatus xmlns:D="DAV:">`)

	// Add current resource
	s.writePropfindResponse(w, file)

	// Add children if depth > 0 and it's a directory
	if depth != "0" && file.IsDirectory {
		children := s.listFiles(path)
		for _, child := range children {
			s.writePropfindResponse(w, child)
		}
	}

	fmt.Fprintf(w, `</D:multistatus>`)
}

// writePropfindResponse writes a single resource response
func (s *MockWebDAVServer) writePropfindResponse(w http.ResponseWriter, file *MockWebDAVFile) {
	href := s.basePath + file.Path
	if file.IsDirectory && !strings.HasSuffix(href, "/") {
		href += "/"
	}

	resourceType := ""
	if file.IsDirectory {
		resourceType = "<D:collection/>"
	}

	fmt.Fprintf(w, `<D:response>`)
	fmt.Fprintf(w, `<D:href>%s</D:href>`, href)
	fmt.Fprintf(w, `<D:propstat>`)
	fmt.Fprintf(w, `<D:prop>`)
	fmt.Fprintf(w, `<D:displayname>%s</D:displayname>`, file.Name)
	fmt.Fprintf(w, `<D:getcontentlength>%d</D:getcontentlength>`, file.Size)
	fmt.Fprintf(w, `<D:getcontenttype>%s</D:getcontenttype>`, file.ContentType)
	fmt.Fprintf(w, `<D:getetag>%s</D:getetag>`, file.ETag)
	fmt.Fprintf(w, `<D:getlastmodified>%s</D:getlastmodified>`, file.ModTime.Format(time.RFC1123))
	fmt.Fprintf(w, `<D:resourcetype>%s</D:resourcetype>`, resourceType)
	fmt.Fprintf(w, `</D:prop>`)
	fmt.Fprintf(w, `<D:status>HTTP/1.1 200 OK</D:status>`)
	fmt.Fprintf(w, `</D:propstat>`)
	fmt.Fprintf(w, `</D:response>`)
}

// handlePut handles PUT requests
func (s *MockWebDAVServer) handlePut(w http.ResponseWriter, r *http.Request, path string) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Extract filename from path
	fileName := filepath.Base(path)

	s.files[path] = &MockWebDAVFile{
		Name:        fileName,
		Path:        path,
		IsDirectory: false,
		Size:        int64(len(content)),
		ModTime:     time.Now(),
		Content:     content,
		ETag:        fmt.Sprintf(`"%x-%d"`, time.Now().Unix(), len(content)),
		ContentType: contentType,
	}

	w.WriteHeader(http.StatusCreated)
}

// handleDelete handles DELETE requests
func (s *MockWebDAVServer) handleDelete(w http.ResponseWriter, r *http.Request, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.files[path]; !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	delete(s.files, path)
	w.WriteHeader(http.StatusNoContent)
}

// handleMkcol handles MKCOL requests
func (s *MockWebDAVServer) handleMkcol(w http.ResponseWriter, r *http.Request, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.files[path]; exists {
		http.Error(w, "Collection already exists", http.StatusMethodNotAllowed)
		return
	}

	fileName := filepath.Base(path)

	s.files[path] = &MockWebDAVFile{
		Name:        fileName,
		Path:        path,
		IsDirectory: true,
		Size:        0,
		ModTime:     time.Now(),
		Content:     []byte{},
		ETag:        fmt.Sprintf(`"%x-0"`, time.Now().Unix()),
		ContentType: "text/html",
	}

	w.WriteHeader(http.StatusCreated)
}

// handleCopy handles COPY requests
func (s *MockWebDAVServer) handleCopy(w http.ResponseWriter, r *http.Request, path string) {
	destination := r.Header.Get("Destination")
	if destination == "" {
		http.Error(w, "Destination header required", http.StatusBadRequest)
		return
	}

	// Parse destination URL to get path
	destPath := strings.TrimPrefix(destination, s.GetAddress())

	s.mu.Lock()
	defer s.mu.Unlock()

	sourceFile, exists := s.files[path]
	if !exists {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	// Create copy
	s.files[destPath] = &MockWebDAVFile{
		Name:        filepath.Base(destPath),
		Path:        destPath,
		IsDirectory: sourceFile.IsDirectory,
		Size:        sourceFile.Size,
		ModTime:     time.Now(),
		Content:     append([]byte(nil), sourceFile.Content...),
		ETag:        fmt.Sprintf(`"%x-%d"`, time.Now().Unix(), sourceFile.Size),
		ContentType: sourceFile.ContentType,
	}

	w.WriteHeader(http.StatusCreated)
}

// handleMove handles MOVE requests
func (s *MockWebDAVServer) handleMove(w http.ResponseWriter, r *http.Request, path string) {
	destination := r.Header.Get("Destination")
	if destination == "" {
		http.Error(w, "Destination header required", http.StatusBadRequest)
		return
	}

	// Parse destination URL to get path
	destPath := strings.TrimPrefix(destination, s.GetAddress())

	s.mu.Lock()
	defer s.mu.Unlock()

	sourceFile, exists := s.files[path]
	if !exists {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	// Move file
	sourceFile.Path = destPath
	sourceFile.Name = filepath.Base(destPath)
	s.files[destPath] = sourceFile
	delete(s.files, path)

	w.WriteHeader(http.StatusCreated)
}

// listFiles returns files in a directory
func (s *MockWebDAVServer) listFiles(path string) []*MockWebDAVFile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var files []*MockWebDAVFile

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

// IsRunning returns true if the server is running
func (s *MockWebDAVServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetFileCount returns the number of files/directories
func (s *MockWebDAVServer) GetFileCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.files)
}

// GetUserCount returns the number of users
func (s *MockWebDAVServer) GetUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}