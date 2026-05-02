//go:build linux
// +build linux

package filesystem

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NFSClient Basic Tests
// =============================================================================

func TestNFSClient_NewNFSClient(t *testing.T) {
	config := NFSConfig{
		Host:       "nfs.example.com",
		Path:       "/export/data",
		MountPoint: "/mnt/nfs",
		Options:    "vers=4",
	}

	client, err := NewNFSClient(config)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, config.Host, client.config.Host)
	assert.Equal(t, config.Path, client.config.Path)
	assert.Equal(t, config.MountPoint, client.config.MountPoint)
	assert.Equal(t, config.Options, client.config.Options)
	assert.False(t, client.IsConnected())
	assert.False(t, client.mounted)
	assert.Equal(t, "/mnt/nfs", client.mountPoint)
}

func TestNFSClient_NewNFSClient_MinimalConfig(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "localhost", client.config.Host)
	assert.Equal(t, "/export", client.config.Path)
	assert.Equal(t, "/mnt/test", client.config.MountPoint)
	assert.Empty(t, client.config.Options)
}

func TestNFSClient_NewNFSClient_NoMountPoint(t *testing.T) {
	config := NFSConfig{
		Host: "localhost",
		Path: "/export",
	}

	client, err := NewNFSClient(config)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "mount point is required")
}

func TestNFSClient_GetProtocol(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	assert.Equal(t, "nfs", client.GetProtocol())
}

func TestNFSClient_GetConfig(t *testing.T) {
	config := NFSConfig{
		Host:       "nfs.example.com",
		Path:       "/export/media",
		MountPoint: "/mnt/media",
		Options:    "vers=4,rw",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	retrievedConfig := client.GetConfig().(*NFSConfig)
	assert.Equal(t, config.Host, retrievedConfig.Host)
	assert.Equal(t, config.Path, retrievedConfig.Path)
	assert.Equal(t, config.MountPoint, retrievedConfig.MountPoint)
	assert.Equal(t, config.Options, retrievedConfig.Options)
}

func TestNFSClient_IsConnected_Initial(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	assert.False(t, client.IsConnected())
	assert.False(t, client.mounted)
	assert.False(t, client.connected)
}

// =============================================================================
// NFSClient Error Handling Tests (Not Connected)
// =============================================================================

func TestNFSClient_TestConnection_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.TestConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_ReadFile_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.ReadFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_WriteFile_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	data := bytes.NewReader([]byte("test data"))
	err = client.WriteFile(ctx, "/test/file.txt", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_GetFileInfo_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.GetFileInfo(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_ListDirectory_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.ListDirectory(ctx, "/test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_FileExists_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.FileExists(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_CreateDirectory_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.CreateDirectory(ctx, "/test/newdir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_DeleteDirectory_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.DeleteDirectory(ctx, "/test/olddir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_DeleteFile_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.DeleteFile(ctx, "/test/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_CopyFile_NotConnected(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.CopyFile(ctx, "/test/src.txt", "/test/dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// =============================================================================
// NFSConfig Validation Tests
// =============================================================================

func TestNFSConfig_AllFields(t *testing.T) {
	config := NFSConfig{
		Host:       "nfs.example.com",
		Path:       "/export/data",
		MountPoint: "/mnt/nfs",
		Options:    "vers=4,rw,soft",
	}

	assert.Equal(t, "nfs.example.com", config.Host)
	assert.Equal(t, "/export/data", config.Path)
	assert.Equal(t, "/mnt/nfs", config.MountPoint)
	assert.Equal(t, "vers=4,rw,soft", config.Options)
}

func TestNFSConfig_DefaultOptions(t *testing.T) {
	config := NFSConfig{
		Host:       "nfs.example.com",
		Path:       "/export",
		MountPoint: "/mnt/nfs",
	}

	assert.Empty(t, config.Options)
}

func TestNFSConfig_CustomMountPoint(t *testing.T) {
	testCases := []struct {
		name       string
		mountPoint string
	}{
		{"Root Mount", "/mnt/nfs"},
		{"Nested Mount", "/mnt/data/nfs"},
		{"Home Mount", "/home/user/nfs"},
		{"Tmp Mount", "/tmp/nfs"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NFSConfig{
				Host:       "localhost",
				Path:       "/export",
				MountPoint: tc.mountPoint,
			}

			assert.Equal(t, tc.mountPoint, config.MountPoint)
		})
	}
}

func TestNFSConfig_NFSVersionOptions(t *testing.T) {
	testCases := []struct {
		name    string
		options string
	}{
		{"NFSv3", "vers=3"},
		{"NFSv4", "vers=4"},
		{"NFSv4.1", "vers=4.1"},
		{"NFSv4.2", "vers=4.2"},
		{"Multiple Options", "vers=4,rw,soft,timeo=600"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NFSConfig{
				Host:       "localhost",
				Path:       "/export",
				MountPoint: "/mnt/test",
				Options:    tc.options,
			}

			assert.Equal(t, tc.options, config.Options)
		})
	}
}

func TestNFSConfig_ReadWriteOptions(t *testing.T) {
	testCases := []struct {
		name    string
		options string
	}{
		{"Read-Write", "rw"},
		{"Read-Only", "ro"},
		{"Soft Mount", "soft"},
		{"Hard Mount", "hard"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NFSConfig{
				Host:       "localhost",
				Path:       "/export",
				MountPoint: "/mnt/test",
				Options:    tc.options,
			}

			assert.Contains(t, config.Options, tc.options)
		})
	}
}

// =============================================================================
// NFSClient State Management Tests
// =============================================================================

func TestNFSClient_StatePersistence(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	// Verify initial state
	assert.False(t, client.IsConnected())
	assert.False(t, client.mounted)
	assert.False(t, client.connected)

	// Verify state doesn't change randomly
	assert.False(t, client.IsConnected())
	assert.False(t, client.mounted)
}

func TestNFSClient_MultipleInstances(t *testing.T) {
	config1 := NFSConfig{
		Host:       "server1.example.com",
		Path:       "/export1",
		MountPoint: "/mnt/nfs1",
	}

	config2 := NFSConfig{
		Host:       "server2.example.com",
		Path:       "/export2",
		MountPoint: "/mnt/nfs2",
	}

	client1, err1 := NewNFSClient(config1)
	client2, err2 := NewNFSClient(config2)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, client1, client2)
	assert.Equal(t, "server1.example.com", client1.config.Host)
	assert.Equal(t, "server2.example.com", client2.config.Host)
	assert.Equal(t, "/mnt/nfs1", client1.mountPoint)
	assert.Equal(t, "/mnt/nfs2", client2.mountPoint)
}

// =============================================================================
// NFSClient Context Tests
// =============================================================================

func TestNFSClient_Disconnect_WithContext(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Disconnect(ctx)
	assert.NoError(t, err) // Should not error when disconnecting already disconnected client
	assert.False(t, client.IsConnected())
	assert.False(t, client.mounted)
}

func TestNFSClient_Disconnect_WhenNotMounted(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Disconnect(ctx)
	assert.NoError(t, err)
	assert.False(t, client.mounted)
	assert.False(t, client.connected)
}

// =============================================================================
// NFSClient Path Security Tests
// =============================================================================

func TestNFSClient_ResolvePath_Security(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple Path", "/data/file.txt", "/mnt/test/data/file.txt"},
		{"Relative Path", "data/file.txt", "/mnt/test/data/file.txt"},
		{"Root Path", "/", "/mnt/test"},
		{"Current Dir", ".", "/mnt/test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolved := client.resolvePath(tc.input)
			assert.Equal(t, tc.expected, resolved)
		})
	}
}

func TestNFSClient_ResolvePath_DirectoryTraversal(t *testing.T) {
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: "/mnt/test",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	// Test directory traversal prevention
	testCases := []string{
		"../etc/passwd",
		"../../etc/shadow",
		"data/../../etc/hosts",
		"/data/../../../root",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			resolved := client.resolvePath(input)
			// Should not contain ".." after resolution
			assert.NotContains(t, resolved, "..")
			// Should still be within mount point
			assert.Contains(t, resolved, "/mnt/test")
		})
	}
}

// =============================================================================
// NOTE: Integration Tests Requiring NFS Server
// =============================================================================

// The following tests require an actual NFS server and mount permissions.
// These tests are skipped by default and require root privileges.

/*
func TestNFSClient_IntegrationConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test - requires NFS server and root privileges")
	}

	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export/test",
		MountPoint: "/tmp/nfs_test_mount",
		Options:    "vers=4",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	assert.True(t, client.IsConnected())
	assert.True(t, client.mounted)
}

func TestNFSClient_IntegrationListDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test - requires NFS server and root privileges")
	}

	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export/test",
		MountPoint: "/tmp/nfs_test_mount",
		Options:    "vers=4",
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Connect(ctx)
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	assert.NotNil(t, files)
}
*/
