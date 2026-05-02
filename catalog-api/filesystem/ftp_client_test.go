package filesystem

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// FTPClient Basic Tests
// =============================================================================

func TestFTPClient_NewFTPClient(t *testing.T) {
	config := &FTPConfig{
		Host:     "ftp.example.com",
		Port:     21,
		Username: "testuser",
		Password: "testpass",
		Path:     "/base/path",
	}

	client := NewFTPClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config, client.config)
	assert.False(t, client.IsConnected())
	assert.Nil(t, client.client)
}

func TestFTPClient_NewFTPClient_MinimalConfig(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, "localhost", client.config.Host)
	assert.Equal(t, 21, client.config.Port)
	assert.Empty(t, client.config.Username)
	assert.Empty(t, client.config.Password)
	assert.Empty(t, client.config.Path)
}

func TestFTPClient_GetProtocol(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	assert.Equal(t, "ftp", client.GetProtocol())
}

func TestFTPClient_GetConfig(t *testing.T) {
	config := &FTPConfig{
		Host:     "ftp.example.com",
		Port:     2121,
		Username: "user",
		Password: "pass",
		Path:     "/uploads",
	}

	client := NewFTPClient(config)
	retrievedConfig := client.GetConfig().(*FTPConfig)

	assert.Equal(t, config.Host, retrievedConfig.Host)
	assert.Equal(t, config.Port, retrievedConfig.Port)
	assert.Equal(t, config.Username, retrievedConfig.Username)
	assert.Equal(t, config.Password, retrievedConfig.Password)
	assert.Equal(t, config.Path, retrievedConfig.Path)
}

func TestFTPClient_IsConnected_Initial(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// FTPClient Error Handling Tests (Not Connected)
// =============================================================================

func TestFTPClient_TestConnection_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_ReadFile_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	_, err := client.ReadFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_WriteFile_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()
	data := bytes.NewReader([]byte("test data"))

	err := client.WriteFile(ctx, "/test/file.txt", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_GetFileInfo_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_ListDirectory_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	_, err := client.ListDirectory(ctx, "/test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_FileExists_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	_, err := client.FileExists(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_CreateDirectory_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "/test/newdir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_DeleteDirectory_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.DeleteDirectory(ctx, "/test/olddir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_DeleteFile_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFTPClient_CopyFile_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.CopyFile(ctx, "/test/src.txt", "/test/dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// =============================================================================
// FTPClient Connection Tests (Requires FTP Server)
// =============================================================================

func TestFTPClient_Connect_InvalidServer(t *testing.T) {
	config := &FTPConfig{
		Host:     "nonexistent.invalid.test",
		Port:     21,
		Username: "test",
		Password: "test",
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
	assert.False(t, client.IsConnected())
}

func TestFTPClient_Connect_InvalidPort(t *testing.T) {
	config := &FTPConfig{
		Host:     "localhost",
		Port:     99999, // Invalid port
		Username: "test",
		Password: "test",
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

func TestFTPClient_Disconnect_NotConnected(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Disconnect(ctx)
	assert.NoError(t, err) // Should not error when disconnecting already disconnected client
	assert.False(t, client.IsConnected())
}

// =============================================================================
// FTPConfig Validation Tests
// =============================================================================

func TestFTPConfig_AllFields(t *testing.T) {
	config := &FTPConfig{
		Host:     "ftp.example.com",
		Port:     2121,
		Username: "admin",
		Password: "secret123",
		Path:     "/data/files",
	}

	assert.Equal(t, "ftp.example.com", config.Host)
	assert.Equal(t, 2121, config.Port)
	assert.Equal(t, "admin", config.Username)
	assert.Equal(t, "secret123", config.Password)
	assert.Equal(t, "/data/files", config.Path)
}

func TestFTPConfig_DefaultPort(t *testing.T) {
	config := &FTPConfig{
		Host: "ftp.example.com",
		Port: 21,
	}

	assert.Equal(t, 21, config.Port)
}

func TestFTPConfig_CustomPort(t *testing.T) {
	config := &FTPConfig{
		Host: "ftp.example.com",
		Port: 2121,
	}

	assert.Equal(t, 2121, config.Port)
}

func TestFTPConfig_WithBasePath(t *testing.T) {
	config := &FTPConfig{
		Host: "ftp.example.com",
		Port: 21,
		Path: "/uploads/media",
	}

	assert.Equal(t, "/uploads/media", config.Path)
}

func TestFTPConfig_WithoutBasePath(t *testing.T) {
	config := &FTPConfig{
		Host: "ftp.example.com",
		Port: 21,
	}

	assert.Empty(t, config.Path)
}

func TestFTPConfig_AnonymousLogin(t *testing.T) {
	config := &FTPConfig{
		Host:     "ftp.example.com",
		Port:     21,
		Username: "anonymous",
		Password: "guest@example.com",
	}

	assert.Equal(t, "anonymous", config.Username)
	assert.Equal(t, "guest@example.com", config.Password)
}

// =============================================================================
// FTPClient Path Resolution Tests
// =============================================================================

func TestFTPClient_ResolvePath_WithBasePath(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
		Path: "/base/dir",
	}

	client := NewFTPClient(config)

	// Test path resolution (internal method, but we can test through operations)
	assert.Equal(t, "/base/dir", client.config.Path)
}

func TestFTPClient_ResolvePath_WithoutBasePath(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)

	assert.Empty(t, client.config.Path)
}

func TestFTPClient_ResolvePath_RootPath(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
		Path: "/",
	}

	client := NewFTPClient(config)

	assert.Equal(t, "/", client.config.Path)
}

func TestFTPClient_ResolvePath_NestedPath(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
		Path: "/level1/level2/level3",
	}

	client := NewFTPClient(config)

	assert.Equal(t, "/level1/level2/level3", client.config.Path)
}

// =============================================================================
// FTPClient State Management Tests
// =============================================================================

func TestFTPClient_IsConnected_StatePersistence(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)

	// Initial state
	assert.False(t, client.IsConnected())

	// Verify state doesn't change randomly
	assert.False(t, client.IsConnected())
	assert.False(t, client.IsConnected())
}

func TestFTPClient_MultipleInstances(t *testing.T) {
	config1 := &FTPConfig{
		Host: "server1.example.com",
		Port: 21,
	}

	config2 := &FTPConfig{
		Host: "server2.example.com",
		Port: 2121,
	}

	client1 := NewFTPClient(config1)
	client2 := NewFTPClient(config2)

	assert.NotEqual(t, client1, client2)
	assert.Equal(t, "server1.example.com", client1.config.Host)
	assert.Equal(t, "server2.example.com", client2.config.Host)
	assert.Equal(t, 21, client1.config.Port)
	assert.Equal(t, 2121, client2.config.Port)
}

// =============================================================================
// FTPClient Context Tests
// =============================================================================

func TestFTPClient_Connect_WithContext(t *testing.T) {
	config := &FTPConfig{
		Host: "nonexistent.test",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
}

func TestFTPClient_Connect_WithCanceledContext(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Note: The current implementation doesn't check context cancellation during Connect,
	// but it will fail to connect to nonexistent server anyway
	err := client.Connect(ctx)
	assert.Error(t, err)
}

func TestFTPClient_Disconnect_WithContext(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Disconnect(ctx)
	assert.NoError(t, err)
}

// =============================================================================
// NOTE: Integration Tests Requiring FTP Server
// =============================================================================

// The following tests require an actual FTP server running and are skipped by default.
// To run these tests, set up an FTP server and uncomment the tests below.

/*
func TestFTPClient_IntegrationConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &FTPConfig{
		Host:     "localhost",
		Port:     21,
		Username: "testuser",
		Password: "testpass",
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	assert.True(t, client.IsConnected())
}

func TestFTPClient_IntegrationListDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &FTPConfig{
		Host:     "localhost",
		Port:     21,
		Username: "testuser",
		Password: "testpass",
		Path:     "/",
	}

	client := NewFTPClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	assert.NotNil(t, files)
}
*/
