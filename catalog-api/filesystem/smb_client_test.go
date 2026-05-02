package filesystem

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// SmbClient Basic Tests
// =============================================================================

func TestSmbClient_NewSmbClient(t *testing.T) {
	config := &SmbConfig{
		Host:     "smb.example.com",
		Port:     445,
		Share:    "data",
		Username: "testuser",
		Password: "testpass",
		Domain:   "WORKGROUP",
	}

	client := NewSmbClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config, client.config)
	assert.False(t, client.IsConnected())
	assert.Nil(t, client.conn)
	assert.Nil(t, client.session)
	assert.Nil(t, client.share)
}

func TestSmbClient_NewSmbClient_MinimalConfig(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "public",
	}

	client := NewSmbClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, "localhost", client.config.Host)
	assert.Equal(t, 445, client.config.Port)
	assert.Equal(t, "public", client.config.Share)
	assert.Empty(t, client.config.Username)
	assert.Empty(t, client.config.Password)
	assert.Empty(t, client.config.Domain)
}

func TestSmbClient_GetProtocol(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	assert.Equal(t, "smb", client.GetProtocol())
}

func TestSmbClient_GetConfig(t *testing.T) {
	config := &SmbConfig{
		Host:     "smb.example.com",
		Port:     445,
		Share:    "files",
		Username: "admin",
		Password: "secret",
		Domain:   "DOMAIN",
	}

	client := NewSmbClient(config)
	retrievedConfig := client.GetConfig().(*SmbConfig)

	assert.Equal(t, config.Host, retrievedConfig.Host)
	assert.Equal(t, config.Port, retrievedConfig.Port)
	assert.Equal(t, config.Share, retrievedConfig.Share)
	assert.Equal(t, config.Username, retrievedConfig.Username)
	assert.Equal(t, config.Password, retrievedConfig.Password)
	assert.Equal(t, config.Domain, retrievedConfig.Domain)
}

func TestSmbClient_IsConnected_Initial(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// SmbClient Error Handling Tests (Not Connected)
// =============================================================================

func TestSmbClient_TestConnection_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_ReadFile_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	_, err := client.ReadFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_WriteFile_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()
	data := bytes.NewReader([]byte("test data"))

	err := client.WriteFile(ctx, "/test/file.txt", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_GetFileInfo_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_ListDirectory_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	_, err := client.ListDirectory(ctx, "/test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_FileExists_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	_, err := client.FileExists(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_CreateDirectory_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "/test/newdir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_DeleteDirectory_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.DeleteDirectory(ctx, "/test/olddir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_DeleteFile_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestSmbClient_CopyFile_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.CopyFile(ctx, "/test/src.txt", "/test/dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// =============================================================================
// SmbClient Connection Tests
// =============================================================================

func TestSmbClient_Connect_InvalidServer(t *testing.T) {
	config := &SmbConfig{
		Host:     "nonexistent.invalid.test",
		Port:     445,
		Share:    "test",
		Username: "user",
		Password: "pass",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
	assert.False(t, client.IsConnected())
}

func TestSmbClient_Disconnect_NotConnected(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.Disconnect(ctx)
	assert.NoError(t, err) // Should not error when disconnecting already disconnected client
	assert.False(t, client.IsConnected())
}

// =============================================================================
// SmbConfig Validation Tests
// =============================================================================

func TestSmbConfig_AllFields(t *testing.T) {
	config := &SmbConfig{
		Host:     "smb.example.com",
		Port:     445,
		Share:    "data",
		Username: "admin",
		Password: "secret123",
		Domain:   "DOMAIN",
	}

	assert.Equal(t, "smb.example.com", config.Host)
	assert.Equal(t, 445, config.Port)
	assert.Equal(t, "data", config.Share)
	assert.Equal(t, "admin", config.Username)
	assert.Equal(t, "secret123", config.Password)
	assert.Equal(t, "DOMAIN", config.Domain)
}

func TestSmbConfig_DefaultPort(t *testing.T) {
	config := &SmbConfig{
		Host:  "smb.example.com",
		Port:  445,
		Share: "files",
	}

	assert.Equal(t, 445, config.Port)
}

func TestSmbConfig_CustomPort(t *testing.T) {
	config := &SmbConfig{
		Host:  "smb.example.com",
		Port:  1445,
		Share: "files",
	}

	assert.Equal(t, 1445, config.Port)
}

func TestSmbConfig_ShareNames(t *testing.T) {
	testCases := []struct {
		name  string
		share string
	}{
		{"Simple Share", "public"},
		{"Data Share", "data"},
		{"User Share", "users"},
		{"Drive Letter", "C$"},
		{"Admin Share", "IPC$"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &SmbConfig{
				Host:  "localhost",
				Port:  445,
				Share: tc.share,
			}

			assert.Equal(t, tc.share, config.Share)
		})
	}
}

func TestSmbConfig_Domains(t *testing.T) {
	testCases := []struct {
		name   string
		domain string
	}{
		{"Workgroup", "WORKGROUP"},
		{"Domain", "DOMAIN"},
		{"Enterprise", "ENTERPRISE"},
		{"Empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &SmbConfig{
				Host:   "localhost",
				Port:   445,
				Share:  "test",
				Domain: tc.domain,
			}

			assert.Equal(t, tc.domain, config.Domain)
		})
	}
}

func TestSmbConfig_Authentication(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"Basic Auth", "user", "pass"},
		{"Empty Password", "user", ""},
		{"Complex Password", "admin", "P@ssw0rd!#$%"},
		{"Domain User", "DOMAIN\\user", "password"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &SmbConfig{
				Host:     "localhost",
				Port:     445,
				Share:    "test",
				Username: tc.username,
				Password: tc.password,
			}

			assert.Equal(t, tc.username, config.Username)
			assert.Equal(t, tc.password, config.Password)
		})
	}
}

// =============================================================================
// SmbClient State Management Tests
// =============================================================================

func TestSmbClient_StatePersistence(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}

	client := NewSmbClient(config)

	// Verify initial state
	assert.False(t, client.IsConnected())

	// Verify state doesn't change randomly
	assert.False(t, client.IsConnected())
	assert.False(t, client.IsConnected())
}

func TestSmbClient_MultipleInstances(t *testing.T) {
	config1 := &SmbConfig{
		Host:  "server1.example.com",
		Port:  445,
		Share: "share1",
	}

	config2 := &SmbConfig{
		Host:  "server2.example.com",
		Port:  445,
		Share: "share2",
	}

	client1 := NewSmbClient(config1)
	client2 := NewSmbClient(config2)

	assert.NotEqual(t, client1, client2)
	assert.Equal(t, "server1.example.com", client1.config.Host)
	assert.Equal(t, "server2.example.com", client2.config.Host)
	assert.Equal(t, "share1", client1.config.Share)
	assert.Equal(t, "share2", client2.config.Share)
}

// =============================================================================
// NOTE: Integration Tests Requiring SMB Server
// =============================================================================

// The following tests require an actual SMB server running.
// These tests are skipped by default.

/*
func TestSmbClient_IntegrationConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &SmbConfig{
		Host:     "localhost",
		Port:     445,
		Share:    "test",
		Username: "testuser",
		Password: "testpass",
		Domain:   "WORKGROUP",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	assert.True(t, client.IsConnected())
}

func TestSmbClient_IntegrationListDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &SmbConfig{
		Host:     "localhost",
		Port:     445,
		Share:    "test",
		Username: "testuser",
		Password: "testpass",
		Domain:   "WORKGROUP",
	}

	client := NewSmbClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	assert.NotNil(t, files)
}
*/
