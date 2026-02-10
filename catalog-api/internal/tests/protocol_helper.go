package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// WebDAVMockServer provides a mock WebDAV server for testing
type WebDAVMockServer struct {
	server *httptest.Server
	files  map[string][]byte
	dirs   map[string]bool
	t      *testing.T
}

// NewWebDAVMockServer creates a new mock WebDAV server
func NewWebDAVMockServer(t *testing.T) *WebDAVMockServer {
	mock := &WebDAVMockServer{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
		t:     t,
	}

	// Create test server
	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock
}

// URL returns the mock server URL
func (m *WebDAVMockServer) URL() string {
	return m.server.URL
}

// AddFile adds a file to the mock server
func (m *WebDAVMockServer) AddFile(path string, content []byte) {
	m.files[path] = content
}

// AddDir adds a directory to the mock server
func (m *WebDAVMockServer) AddDir(path string) {
	m.dirs[path] = true
}

// handleRequest handles HTTP requests for the mock server
func (m *WebDAVMockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch r.Method {
	case "PROPFIND":
		m.handlePropFind(w, r, path)
	case "GET":
		m.handleGet(w, r, path)
	case "PUT":
		m.handlePut(w, r, path)
	case "DELETE":
		m.handleDelete(w, r, path)
	case "MKCOL":
		m.handleMkCol(w, r, path)
	case "OPTIONS":
		m.handleOptions(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *WebDAVMockServer) handlePropFind(w http.ResponseWriter, r *http.Request, path string) {
	// Return multistatus response
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusMultiStatus)

	// Simple XML response for directory listing
	fmt.Fprintf(w, `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
  <D:response>
    <D:href>%s</D:href>
    <D:propstat>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`, path)
}

func (m *WebDAVMockServer) handleGet(w http.ResponseWriter, r *http.Request, path string) {
	if content, exists := m.files[path]; exists {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *WebDAVMockServer) handlePut(w http.ResponseWriter, r *http.Request, path string) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	m.files[path] = content
	w.WriteHeader(http.StatusCreated)
}

func (m *WebDAVMockServer) handleDelete(w http.ResponseWriter, r *http.Request, path string) {
	if _, exists := m.files[path]; exists {
		delete(m.files, path)
		w.WriteHeader(http.StatusNoContent)
	} else if _, exists := m.dirs[path]; exists {
		delete(m.dirs, path)
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *WebDAVMockServer) handleMkCol(w http.ResponseWriter, r *http.Request, path string) {
	m.dirs[path] = true
	w.WriteHeader(http.StatusCreated)
}

func (m *WebDAVMockServer) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("DAV", "1, 2")
	w.Header().Set("Allow", "GET, PUT, DELETE, PROPFIND, MKCOL, OPTIONS")
	w.WriteHeader(http.StatusOK)
}

// Close closes the mock server
func (m *WebDAVMockServer) Close() {
	m.server.Close()
}

// FTPMockConfig provides configuration for FTP testing
type FTPMockConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	BasePath string
}

// GetFTPMockConfig returns a mock FTP configuration for testing
// Note: This doesn't start an actual FTP server, just provides config
// Real FTP testing would require a containerized FTP server
func GetFTPMockConfig(t *testing.T) FTPMockConfig {
	return FTPMockConfig{
		Host:     "localhost",
		Port:     21,
		Username: "testuser",
		Password: "testpass",
		BasePath: "/test",
	}
}

// NFSMockConfig provides configuration for NFS testing
type NFSMockConfig struct {
	Host       string
	ExportPath string
	MountPoint string
}

// GetNFSMockConfig returns a mock NFS configuration for testing
// Note: This doesn't mount an actual NFS share, just provides config
// Real NFS testing would require proper mount permissions
func GetNFSMockConfig(t *testing.T) NFSMockConfig {
	return NFSMockConfig{
		Host:       "localhost",
		ExportPath: "/export/test",
		MountPoint: "/tmp/nfs-test",
	}
}

// ConcurrentTestHelper provides utilities for concurrent testing
type ConcurrentTestHelper struct {
	t *testing.T
}

// NewConcurrentTestHelper creates a new concurrent test helper
func NewConcurrentTestHelper(t *testing.T) *ConcurrentTestHelper {
	return &ConcurrentTestHelper{t: t}
}

// RunConcurrent runs n goroutines executing fn concurrently
func (h *ConcurrentTestHelper) RunConcurrent(n int, fn func(id int)) {
	done := make(chan bool, n)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					h.t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()
			fn(id)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < n; i++ {
		<-done
	}
}

// RunConcurrentWithErrors runs n goroutines and collects errors
func (h *ConcurrentTestHelper) RunConcurrentWithErrors(n int, fn func(id int) error) []error {
	type result struct {
		id  int
		err error
	}

	results := make(chan result, n)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					results <- result{id: id, err: fmt.Errorf("panic: %v", r)}
				}
			}()
			err := fn(id)
			results <- result{id: id, err: err}
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < n; i++ {
		res := <-results
		if res.err != nil {
			errors = append(errors, fmt.Errorf("goroutine %d: %w", res.id, res.err))
		}
	}

	return errors
}

// AssertNoErrors asserts that there are no errors in the slice
func (h *ConcurrentTestHelper) AssertNoErrors(errors []error) {
	if len(errors) > 0 {
		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}
		h.t.Fatalf("Concurrent test failed with %d errors:\n%s",
			len(errors), strings.Join(errMsgs, "\n"))
	}
}

// MockHTTPServer creates a simple mock HTTP server for testing
func MockHTTPServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// MockHTTPServerWithAuth creates a mock HTTP server with basic auth
func MockHTTPServerWithAuth(t *testing.T, username, password string, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}))
}
