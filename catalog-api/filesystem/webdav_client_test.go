package filesystem

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// WebDAVClient Basic Tests
// =============================================================================

func TestWebDAVClient_NewWebDAVClient(t *testing.T) {
	config := &WebDAVConfig{
		URL:      "https://webdav.example.com",
		Username: "testuser",
		Password: "testpass",
		Path:     "/remote/path",
	}

	client := NewWebDAVClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config, client.config)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.baseURL)
	assert.False(t, client.IsConnected())
	assert.Equal(t, "/remote/path", client.baseURL.Path)
}

func TestWebDAVClient_NewWebDAVClient_MinimalConfig(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost:8080/webdav",
	}

	client := NewWebDAVClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, "https://localhost:8080/webdav", config.URL)
	assert.Empty(t, config.Username)
	assert.Empty(t, config.Password)
	assert.Empty(t, config.Path)
}

func TestWebDAVClient_NewWebDAVClient_WithBasePath(t *testing.T) {
	config := &WebDAVConfig{
		URL:  "https://example.com",
		Path: "/dav/files",
	}

	client := NewWebDAVClient(config)

	assert.Equal(t, "/dav/files", client.baseURL.Path)
}

func TestWebDAVClient_NewWebDAVClient_RootPath(t *testing.T) {
	config := &WebDAVConfig{
		URL:  "https://example.com",
		Path: "/",
	}

	client := NewWebDAVClient(config)

	// Root path should be empty after processing
	assert.Equal(t, "", client.baseURL.Path)
}

func TestWebDAVClient_GetProtocol(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	assert.Equal(t, "webdav", client.GetProtocol())
}

func TestWebDAVClient_GetConfig(t *testing.T) {
	config := &WebDAVConfig{
		URL:      "https://webdav.example.com",
		Username: "user",
		Password: "pass",
		Path:     "/files",
	}

	client := NewWebDAVClient(config)
	retrievedConfig := client.GetConfig().(*WebDAVConfig)

	assert.Equal(t, config.URL, retrievedConfig.URL)
	assert.Equal(t, config.Username, retrievedConfig.Username)
	assert.Equal(t, config.Password, retrievedConfig.Password)
	assert.Equal(t, config.Path, retrievedConfig.Path)
}

func TestWebDAVClient_IsConnected_Initial(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// WebDAVClient Error Handling Tests (Not Connected)
// =============================================================================

func TestWebDAVClient_TestConnection_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_ReadFile_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	_, err := client.ReadFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_WriteFile_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()
	data := bytes.NewReader([]byte("test data"))

	err := client.WriteFile(ctx, "/test/file.txt", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_GetFileInfo_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_ListDirectory_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	_, err := client.ListDirectory(ctx, "/test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_FileExists_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	_, err := client.FileExists(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_CreateDirectory_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "/test/newdir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_DeleteDirectory_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.DeleteDirectory(ctx, "/test/olddir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_DeleteFile_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_CopyFile_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.CopyFile(ctx, "/test/src.txt", "/test/dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// =============================================================================
// WebDAVConfig Validation Tests
// =============================================================================

func TestWebDAVConfig_AllFields(t *testing.T) {
	config := &WebDAVConfig{
		URL:      "https://webdav.example.com:8080",
		Username: "admin",
		Password: "secret123",
		Path:     "/data/files",
	}

	assert.Equal(t, "https://webdav.example.com:8080", config.URL)
	assert.Equal(t, "admin", config.Username)
	assert.Equal(t, "secret123", config.Password)
	assert.Equal(t, "/data/files", config.Path)
}

func TestWebDAVConfig_HTTPSEndpoint(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://secure.example.com",
	}

	assert.Contains(t, config.URL, "https://")
}

func TestWebDAVConfig_HTTPEndpoint(t *testing.T) {
	config := &WebDAVConfig{
		URL: "http://insecure.example.com",
	}

	assert.Contains(t, config.URL, "http://")
}

func TestWebDAVConfig_CustomPort(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{"Standard HTTPS", "https://example.com:443"},
		{"Standard HTTP", "http://example.com:80"},
		{"Custom Port", "https://example.com:8443"},
		{"Alternative Port", "http://example.com:8080"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &WebDAVConfig{
				URL: tc.url,
			}

			assert.Equal(t, tc.url, config.URL)
		})
	}
}

func TestWebDAVConfig_BasePath(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{"Root", "/"},
		{"Simple Path", "/webdav"},
		{"Nested Path", "/remote/dav/files"},
		{"User Path", "/users/john/files"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &WebDAVConfig{
				URL:  "https://example.com",
				Path: tc.path,
			}

			assert.Equal(t, tc.path, config.Path)
		})
	}
}

func TestWebDAVConfig_Authentication(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"Basic Auth", "user", "pass"},
		{"Empty Password", "user", ""},
		{"Complex Password", "admin", "P@ssw0rd!#$%"},
		{"Email Username", "user@example.com", "password"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &WebDAVConfig{
				URL:      "https://example.com",
				Username: tc.username,
				Password: tc.password,
			}

			assert.Equal(t, tc.username, config.Username)
			assert.Equal(t, tc.password, config.Password)
		})
	}
}

// =============================================================================
// WebDAVClient State Management Tests
// =============================================================================

func TestWebDAVClient_Disconnect_NotConnected(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.Disconnect(ctx)
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestWebDAVClient_StatePersistence(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://localhost",
	}

	client := NewWebDAVClient(config)

	// Verify initial state
	assert.False(t, client.IsConnected())

	// Verify state doesn't change randomly
	assert.False(t, client.IsConnected())
	assert.False(t, client.IsConnected())
}

func TestWebDAVClient_MultipleInstances(t *testing.T) {
	config1 := &WebDAVConfig{
		URL:  "https://server1.example.com",
		Path: "/files1",
	}

	config2 := &WebDAVConfig{
		URL:  "https://server2.example.com",
		Path: "/files2",
	}

	client1 := NewWebDAVClient(config1)
	client2 := NewWebDAVClient(config2)

	assert.NotEqual(t, client1, client2)
	assert.Equal(t, "https://server1.example.com", client1.config.URL)
	assert.Equal(t, "https://server2.example.com", client2.config.URL)
	assert.Equal(t, "/files1", client1.baseURL.Path)
	assert.Equal(t, "/files2", client2.baseURL.Path)
}

// =============================================================================
// WebDAVClient URL Resolution Tests
// =============================================================================

func TestWebDAVClient_ResolveURL_SimplePaths(t *testing.T) {
	config := &WebDAVConfig{
		URL:  "https://example.com",
		Path: "/webdav",
	}

	client := NewWebDAVClient(config)

	testCases := []struct {
		name     string
		input    string
		contains string
	}{
		{"Simple File", "/file.txt", "https://example.com/webdav/file.txt"},
		{"Nested File", "/data/file.txt", "https://example.com/webdav/data/file.txt"},
		{"Root", "/", "https://example.com/webdav"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolved := client.resolveURL(tc.input)
			assert.Contains(t, resolved, tc.contains)
		})
	}
}

func TestWebDAVClient_ResolveURL_DirectoryTraversal(t *testing.T) {
	config := &WebDAVConfig{
		URL:  "https://example.com",
		Path: "/webdav",
	}

	client := NewWebDAVClient(config)

	// Test directory traversal prevention
	testCases := []string{
		"../etc/passwd",
		"../../etc/shadow",
		"data/../../etc/hosts",
		"/data/../../../root",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			resolved := client.resolveURL(input)
			// Should not contain ".." after resolution
			assert.NotContains(t, resolved, "..")
			// Should still contain base URL
			assert.Contains(t, resolved, "https://example.com")
		})
	}
}

// =============================================================================
// WebDAVClient Context Tests
// =============================================================================

func TestWebDAVClient_Connect_InvalidServer(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://nonexistent.invalid.test",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

func TestWebDAVClient_Connect_WithAuthentication(t *testing.T) {
	config := &WebDAVConfig{
		URL:      "https://nonexistent.test",
		Username: "user",
		Password: "pass",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err) // Should fail to connect to nonexistent server
	assert.False(t, client.IsConnected())
}

func TestWebDAVClient_Connect_WithCanceledContext(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://example.com",
	}

	client := NewWebDAVClient(config)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// WebDAVClient HTTP Client Tests
// =============================================================================

func TestWebDAVClient_HTTPClientInitialization(t *testing.T) {
	config := &WebDAVConfig{
		URL: "https://example.com",
	}

	client := NewWebDAVClient(config)

	assert.NotNil(t, client.client)
	assert.Equal(t, 30000000000, int(client.client.Timeout)) // 30 seconds in nanoseconds
}

func TestWebDAVClient_BaseURLParsing(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectedHost string
		expectedScheme string
	}{
		{"HTTPS URL", "https://example.com", "example.com", "https"},
		{"HTTP URL", "http://example.com", "example.com", "http"},
		{"With Port", "https://example.com:8443", "example.com:8443", "https"},
		{"With Path", "https://example.com/webdav", "example.com", "https"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &WebDAVConfig{
				URL: tc.url,
			}

			client := NewWebDAVClient(config)

			assert.Equal(t, tc.expectedHost, client.baseURL.Host)
			assert.Equal(t, tc.expectedScheme, client.baseURL.Scheme)
		})
	}
}

// =============================================================================
// NOTE: Integration Tests Requiring WebDAV Server
// =============================================================================

// The following tests require an actual WebDAV server running.
// These tests are skipped by default.

/*
func TestWebDAVClient_IntegrationConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &WebDAVConfig{
		URL:      "http://localhost:8080/webdav",
		Username: "testuser",
		Password: "testpass",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	assert.True(t, client.IsConnected())
}

func TestWebDAVClient_IntegrationListDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &WebDAVConfig{
		URL:      "http://localhost:8080/webdav",
		Username: "testuser",
		Password: "testpass",
		Path:     "/",
	}

	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	assert.NotNil(t, files)
}
*/
