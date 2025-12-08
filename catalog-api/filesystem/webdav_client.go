package filesystem

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// WebDAVConfig contains WebDAV connection configuration
type WebDAVConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"` // Base path on the WebDAV server
}

// WebDAVClient implements FileSystemClient for WebDAV protocol
type WebDAVClient struct {
	config    *WebDAVConfig
	client    *http.Client
	baseURL   *url.URL
	connected bool
}

// NewWebDAVClient creates a new WebDAV client
func NewWebDAVClient(config *WebDAVConfig) *WebDAVClient {
	baseURL, _ := url.Parse(config.URL)
	if config.Path != "" && config.Path != "/" {
		baseURL.Path = config.Path
	}

	return &WebDAVClient{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

// Connect establishes the WebDAV connection
func (c *WebDAVClient) Connect(ctx context.Context) error {
	// Test the connection with a PROPFIND request
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", c.baseURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create PROPFIND request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	req.Header.Set("Depth", "0")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to WebDAV server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WebDAV server returned status %d", resp.StatusCode)
	}

	c.connected = true
	return nil
}

// Disconnect closes the WebDAV connection
func (c *WebDAVClient) Disconnect(ctx context.Context) error {
	c.connected = false
	return nil
}

// IsConnected returns true if the client is connected
func (c *WebDAVClient) IsConnected() bool {
	return c.connected
}

// TestConnection tests the WebDAV connection
func (c *WebDAVClient) TestConnection(ctx context.Context) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}
	return c.Connect(ctx) // Re-test connection
}

// resolveURL resolves a relative path to a full WebDAV URL
func (c *WebDAVClient) resolveURL(path string) string {
	// Clean the path and prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		// Prevent directory traversal attacks
		cleanPath = strings.ReplaceAll(cleanPath, "..", "")
	}

	u := *c.baseURL
	u.Path = filepath.Join(u.Path, cleanPath)
	return u.String()
}

// ReadFile reads a file from the WebDAV server
func (c *WebDAVClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve WebDAV file %s: %w", fullURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("WebDAV server returned status %d for file %s", resp.StatusCode, fullURL)
	}

	return resp.Body, nil
}

// WriteFile writes a file to the WebDAV server
func (c *WebDAVClient) WriteFile(ctx context.Context, path string, data io.Reader) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "PUT", fullURL, data)
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload WebDAV file %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV server returned status %d for file %s", resp.StatusCode, fullURL)
	}

	return nil
}

// GetFileInfo gets information about a file
func (c *WebDAVClient) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "HEAD", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get WebDAV file info %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WebDAV server returned status %d for file %s", resp.StatusCode, fullURL)
	}

	// Parse content length
	size := int64(0)
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		if s, err := strconv.ParseInt(cl, 10, 64); err == nil {
			size = s
		}
	}

	// Parse last modified
	modTime := time.Now()
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := time.Parse(time.RFC1123, lm); err == nil {
			modTime = t
		}
	}

	// Check if it's a directory (simplified check)
	isDir := strings.HasSuffix(path, "/") || resp.Header.Get("Content-Type") == "httpd/unix-directory"

	return &FileInfo{
		Name:    filepath.Base(path),
		Size:    size,
		ModTime: modTime,
		IsDir:   isDir,
		Mode:    0644, // Default mode
		Path:    path,
	}, nil
}

// ListDirectory lists files in a directory
func (c *WebDAVClient) ListDirectory(ctx context.Context, path string) ([]*FileInfo, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create PROPFIND request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "application/xml")

	// Create a minimal PROPFIND request body
	body := `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:">
	<D:prop>
		<D:displayname/>
		<D:getcontentlength/>
		<D:getlastmodified/>
		<D:resourcetype/>
	</D:prop>
</D:propfind>`

	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = int64(len(body))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list WebDAV directory %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("WebDAV server returned status %d for directory %s", resp.StatusCode, fullURL)
	}

	// Parse XML response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read WebDAV response: %w", err)
	}

	// Simple XML parsing for WebDAV multistatus response
	var files []*FileInfo

	// Parse the XML to extract file information
	// This is a simplified parser - in production you might want to use a proper XML decoder
	responseStr := string(bodyBytes)

	// Find all <D:response> elements
	responses := strings.Split(responseStr, "<D:response>")

	for i := 1; i < len(responses); i++ { // Skip first element as it's before the first response
		response := responses[i]

		// Find the end of this response
		endIndex := strings.Index(response, "</D:response>")
		if endIndex == -1 {
			continue
		}
		response = response[:endIndex]

		// Extract href
		hrefStart := strings.Index(response, "<D:href>")
		hrefEnd := strings.Index(response, "</D:href>")
		if hrefStart == -1 || hrefEnd == -1 {
			continue
		}
		href := response[hrefStart+8 : hrefEnd]

		// Skip the directory itself (usually the first response)
		if href == fullURL || href == strings.TrimSuffix(fullURL, "/") {
			continue
		}

		// Extract display name
		displayName := filepath.Base(href)
		nameStart := strings.Index(response, "<D:displayname>")
		nameEnd := strings.Index(response, "</D:displayname>")
		if nameStart != -1 && nameEnd != -1 {
			displayName = response[nameStart+16 : nameEnd]
		}

		// Extract content length
		var size int64 = 0
		sizeStart := strings.Index(response, "<D:getcontentlength>")
		sizeEnd := strings.Index(response, "</D:getcontentlength>")
		if sizeStart != -1 && sizeEnd != -1 {
			sizeStr := response[sizeStart+20 : sizeEnd]
			if s, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				size = s
			}
		}

		// Extract last modified date
		modTime := time.Now()
		modStart := strings.Index(response, "<D:getlastmodified>")
		modEnd := strings.Index(response, "</D:getlastmodified>")
		if modStart != -1 && modEnd != -1 {
			modStr := response[modStart+20 : modEnd]
			// Try RFC1123 format first
			if t, err := time.Parse(time.RFC1123, modStr); err == nil {
				modTime = t
			} else if t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", modStr); err == nil {
				modTime = t
			}
		}

		// Check if it's a directory
		isDir := false
		if strings.Contains(response, "<D:resourcetype><D:collection/></D:resourcetype>") ||
			strings.Contains(response, "<D:resourcetype><D:directory/></D:resourcetype>") {
			isDir = true
		}

		// Create relative path from full URL
		relPath := strings.TrimPrefix(href, fullURL)
		if relPath == "" {
			relPath = displayName
		} else {
			relPath = strings.TrimPrefix(relPath, "/")
		}

		files = append(files, &FileInfo{
			Name:    displayName,
			Size:    size,
			ModTime: modTime,
			IsDir:   isDir,
			Mode:    0644, // Default mode
			Path:    relPath,
		})
	}

	return files, nil
}

// FileExists checks if a file exists
func (c *WebDAVClient) FileExists(ctx context.Context, path string) (bool, error) {
	if !c.IsConnected() {
		return false, fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "HEAD", fullURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check WebDAV file existence %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// CreateDirectory creates a directory
func (c *WebDAVClient) CreateDirectory(ctx context.Context, path string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "MKCOL", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create MKCOL request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create WebDAV directory %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV server returned status %d for directory %s", resp.StatusCode, fullURL)
	}

	return nil
}

// DeleteDirectory deletes a directory
func (c *WebDAVClient) DeleteDirectory(ctx context.Context, path string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "DELETE", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete WebDAV directory %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV server returned status %d for directory %s", resp.StatusCode, fullURL)
	}

	return nil
}

// DeleteFile deletes a file
func (c *WebDAVClient) DeleteFile(ctx context.Context, path string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	fullURL := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, "DELETE", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete WebDAV file %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV server returned status %d for file %s", resp.StatusCode, fullURL)
	}

	return nil
}

// CopyFile copies a file on the WebDAV server
func (c *WebDAVClient) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	srcURL := c.resolveURL(srcPath)
	dstURL := c.resolveURL(dstPath)

	req, err := http.NewRequestWithContext(ctx, "COPY", srcURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create COPY request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	req.Header.Set("Destination", dstURL)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to copy WebDAV file from %s to %s: %w", srcURL, dstURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV server returned status %d for copy operation", resp.StatusCode)
	}

	return nil
}

// GetProtocol returns the protocol name
func (c *WebDAVClient) GetProtocol() string {
	return "webdav"
}

// GetConfig returns the WebDAV configuration
func (c *WebDAVClient) GetConfig() interface{} {
	return c.config
}
