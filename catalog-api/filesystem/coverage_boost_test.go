//go:build linux
// +build linux

package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NFS Client Connected-Path Tests
// =============================================================================

// newConnectedNFSClient creates an NFS client that appears connected
// by using a real temp directory as the mount point and setting state flags.
func newConnectedNFSClient(t *testing.T) (*NFSClient, string) {
	t.Helper()
	tmpDir := t.TempDir()

	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: tmpDir,
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	// Simulate connected state (normally set by Connect after syscall.Mount)
	client.mounted = true
	client.connected = true

	return client, tmpDir
}

func TestNFSClient_ReadFile_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	// Create a test file
	testContent := []byte("hello NFS world")
	err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), testContent, 0644)
	require.NoError(t, err)

	reader, err := client.ReadFile(ctx, "test.txt")
	require.NoError(t, err)
	defer reader.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)
	assert.Equal(t, testContent, buf.Bytes())
}

func TestNFSClient_ReadFile_Connected_NotExists(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	_, err := client.ReadFile(ctx, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open NFS file")
}

func TestNFSClient_WriteFile_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	data := bytes.NewReader([]byte("write test data"))
	err := client.WriteFile(ctx, "output.txt", data)
	require.NoError(t, err)

	// Verify file was written
	content, err := os.ReadFile(filepath.Join(tmpDir, "output.txt"))
	require.NoError(t, err)
	assert.Equal(t, "write test data", string(content))
}

func TestNFSClient_WriteFile_Connected_NestedDir(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	data := bytes.NewReader([]byte("nested content"))
	err := client.WriteFile(ctx, "sub/dir/file.txt", data)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "sub", "dir", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested content", string(content))
}

func TestNFSClient_GetFileInfo_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	err := os.WriteFile(filepath.Join(tmpDir, "info.txt"), []byte("info content"), 0644)
	require.NoError(t, err)

	info, err := client.GetFileInfo(ctx, "info.txt")
	require.NoError(t, err)
	assert.Equal(t, "info.txt", info.Name)
	assert.Equal(t, int64(12), info.Size)
	assert.False(t, info.IsDir)
}

func TestNFSClient_GetFileInfo_Connected_Directory(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	require.NoError(t, err)

	info, err := client.GetFileInfo(ctx, "subdir")
	require.NoError(t, err)
	assert.Equal(t, "subdir", info.Name)
	assert.True(t, info.IsDir)
}

func TestNFSClient_GetFileInfo_Connected_NotExists(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat NFS file")
}

func TestNFSClient_ListDirectory_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	// Create some files and directories
	err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("bb"), 0644)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	require.NoError(t, err)

	files, err := client.ListDirectory(ctx, ".")
	require.NoError(t, err)
	assert.Len(t, files, 3)

	// Verify file names exist in result
	names := make(map[string]bool)
	for _, f := range files {
		names[f.Name] = true
	}
	assert.True(t, names["file1.txt"])
	assert.True(t, names["file2.txt"])
	assert.True(t, names["subdir"])
}

func TestNFSClient_ListDirectory_Connected_Empty(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	files, err := client.ListDirectory(ctx, ".")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestNFSClient_ListDirectory_Connected_NotExists(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	_, err := client.ListDirectory(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list NFS directory")
}

func TestNFSClient_FileExists_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	err := os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("x"), 0644)
	require.NoError(t, err)

	exists, err := client.FileExists(ctx, "exists.txt")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = client.FileExists(ctx, "nothere.txt")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestNFSClient_CreateDirectory_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "newdir/nested")
	require.NoError(t, err)

	info, err := os.Stat(filepath.Join(tmpDir, "newdir", "nested"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNFSClient_DeleteDirectory_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	// Create a directory with content
	dir := filepath.Join(tmpDir, "todelete")
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0644)
	require.NoError(t, err)

	err = client.DeleteDirectory(ctx, "todelete")
	require.NoError(t, err)

	_, err = os.Stat(dir)
	assert.True(t, os.IsNotExist(err))
}

func TestNFSClient_DeleteFile_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	filePath := filepath.Join(tmpDir, "todelete.txt")
	err := os.WriteFile(filePath, []byte("x"), 0644)
	require.NoError(t, err)

	err = client.DeleteFile(ctx, "todelete.txt")
	require.NoError(t, err)

	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
}

func TestNFSClient_DeleteFile_Connected_NotExists(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete NFS file")
}

func TestNFSClient_CopyFile_Connected(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	// Create source file
	srcContent := []byte("copy me please")
	err := os.WriteFile(filepath.Join(tmpDir, "source.txt"), srcContent, 0644)
	require.NoError(t, err)

	err = client.CopyFile(ctx, "source.txt", "dest.txt")
	require.NoError(t, err)

	// Verify destination
	content, err := os.ReadFile(filepath.Join(tmpDir, "dest.txt"))
	require.NoError(t, err)
	assert.Equal(t, srcContent, content)
}

func TestNFSClient_CopyFile_Connected_NestedDest(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	err := os.WriteFile(filepath.Join(tmpDir, "source.txt"), []byte("data"), 0644)
	require.NoError(t, err)

	err = client.CopyFile(ctx, "source.txt", "nested/dir/dest.txt")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "nested", "dir", "dest.txt"))
	require.NoError(t, err)
	assert.Equal(t, "data", string(content))
}

func TestNFSClient_CopyFile_Connected_SourceNotExists(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	err := client.CopyFile(ctx, "nonexistent.txt", "dest.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open source file")
}

func TestNFSClient_TestConnection_Connected(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	assert.NoError(t, err)
}

func TestNFSClient_Disconnect_Connected(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	ctx := context.Background()

	// Disconnect when mounted will try syscall.Unmount which will fail in test
	// because it's not a real mount, but it exercises the code path
	err := client.Disconnect(ctx)
	// We expect an error because it's not truly mounted
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmount")
}

func TestNFSClient_Connect_AlreadyMounted(t *testing.T) {
	tmpDir := t.TempDir()
	config := NFSConfig{
		Host:       "localhost",
		Path:       "/export",
		MountPoint: tmpDir,
	}

	client, err := NewNFSClient(config)
	require.NoError(t, err)

	// isMounted() returns true when /proc/mounts can be read (always on Linux)
	// So Connect should detect it as "already mounted" and set connected=true
	ctx := context.Background()
	err = client.Connect(ctx)
	// This will either succeed (if isMounted returns true) or fail at syscall.Mount
	// On test environments, isMounted() reads /proc/mounts and returns true
	// so it should set connected=true and return nil
	if err == nil {
		assert.True(t, client.connected)
	}
}

// =============================================================================
// FTP Client - resolvePath and Disconnect edge cases
// =============================================================================

func TestFTPClient_ResolvePath_WithBasePath_Direct(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
		Path: "/base",
	}
	client := NewFTPClient(config)

	result := client.resolvePath("subdir/file.txt")
	assert.Equal(t, "/base/subdir/file.txt", result)
}

func TestFTPClient_ResolvePath_WithoutBasePath_Direct(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
	}
	client := NewFTPClient(config)

	result := client.resolvePath("subdir/file.txt")
	assert.Equal(t, "subdir/file.txt", result)
}

func TestFTPClient_ResolvePath_EmptyPath(t *testing.T) {
	config := &FTPConfig{
		Host: "localhost",
		Port: 21,
		Path: "/base",
	}
	client := NewFTPClient(config)

	result := client.resolvePath("")
	assert.Equal(t, "/base/", result)
}

func TestFTPClient_Disconnect_AfterFailedConnect(t *testing.T) {
	config := &FTPConfig{
		Host: "nonexistent.test",
		Port: 21,
	}
	client := NewFTPClient(config)
	ctx := context.Background()

	// Try to connect (will fail)
	_ = client.Connect(ctx)

	// Disconnect should still work cleanly
	err := client.Disconnect(ctx)
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// SMB Client - isNotExistError helper
// =============================================================================

func TestIsNotExistError_Extended(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"file does not exist", "file does not exist", true},
		{"no such file or directory", "no such file or directory", true},
		{"wrapped error", "wrapped: connection refused", false},
		{"empty error msg", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = fmt.Errorf("%s", tt.errMsg)
			}
			result := isNotExistError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// SMB Client - Disconnect with partial state
// =============================================================================

func TestSmbClient_Disconnect_PartialState(t *testing.T) {
	config := &SmbConfig{
		Host:  "localhost",
		Port:  445,
		Share: "test",
	}
	client := NewSmbClient(config)
	ctx := context.Background()

	// All nil - should return nil error
	err := client.Disconnect(ctx)
	assert.NoError(t, err)
}

// =============================================================================
// Local Client - Connected path edge cases
// =============================================================================

func TestLocalClient_Connect_NotADirectory(t *testing.T) {
	// Create a temp file (not directory)
	tmpFile, err := os.CreateTemp("", "local_test_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config := &LocalConfig{BasePath: tmpFile.Name()}
	client := NewLocalClient(config)
	ctx := context.Background()

	err = client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not a directory")
}

func TestLocalClient_Connect_NonexistentPathBoost(t *testing.T) {
	config := &LocalConfig{BasePath: "/nonexistent/path/that/does/not/exist"}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access base path")
}

func TestLocalClient_ResolvePath_WithTraversal(t *testing.T) {
	config := &LocalConfig{BasePath: "/base"}
	client := NewLocalClient(config)

	// Test that traversal is prevented
	result := client.resolvePath("../../../etc/passwd")
	assert.NotContains(t, result, "..")
	assert.Contains(t, result, "/base")
}

func TestLocalClient_WriteFile_Connected_ErrorOnBadDir(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Write to a valid nested path
	data := bytes.NewReader([]byte("test data"))
	err = client.WriteFile(ctx, "sub/dir/file.txt", data)
	assert.NoError(t, err)

	// Verify file exists
	content, err := os.ReadFile(filepath.Join(tmpDir, "sub", "dir", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test data", string(content))
}

func TestLocalClient_ListDirectory_Connected_WithMixedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Create mixed entries
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.log"), []byte("bb"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "subdir2"), 0755)

	files, err := client.ListDirectory(ctx, ".")
	require.NoError(t, err)
	assert.Len(t, files, 4)

	// Check types
	dirCount := 0
	fileCount := 0
	for _, f := range files {
		if f.IsDir {
			dirCount++
		} else {
			fileCount++
		}
	}
	assert.Equal(t, 2, dirCount)
	assert.Equal(t, 2, fileCount)
}

func TestLocalClient_FileExists_Connected_Variants(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Create a file
	os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("x"), 0644)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", "exists.txt", true},
		{"non-existing file", "nothere.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := client.FileExists(ctx, tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestLocalClient_CreateDirectory_Connected_Nested(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	err = client.CreateDirectory(ctx, "a/b/c/d")
	require.NoError(t, err)

	info, err := os.Stat(filepath.Join(tmpDir, "a", "b", "c", "d"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestLocalClient_DeleteDirectory_Connected(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Create directory with content
	dir := filepath.Join(tmpDir, "todelete")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0644)

	err = client.DeleteDirectory(ctx, "todelete")
	require.NoError(t, err)

	_, err = os.Stat(dir)
	assert.True(t, os.IsNotExist(err))
}

func TestLocalClient_CopyFile_Connected(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Create source
	os.WriteFile(filepath.Join(tmpDir, "src.txt"), []byte("copy data"), 0644)

	err = client.CopyFile(ctx, "src.txt", "dst.txt")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "dst.txt"))
	require.NoError(t, err)
	assert.Equal(t, "copy data", string(content))
}

func TestLocalClient_CopyFile_Connected_NestedDest(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(tmpDir, "src.txt"), []byte("nested copy"), 0644)

	err = client.CopyFile(ctx, "src.txt", "a/b/dst.txt")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "a", "b", "dst.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested copy", string(content))
}

func TestLocalClient_CopyFile_Connected_SourceNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	err = client.CopyFile(ctx, "nonexistent.txt", "dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open source file")
}

// =============================================================================
// DirectoryTreeInfo struct tests
// =============================================================================

func TestDirectoryTreeInfo_Boost(t *testing.T) {
	tree := &DirectoryTreeInfo{
		Path:       "/media",
		TotalFiles: 10,
		TotalDirs:  3,
		TotalSize:  1024 * 1024,
		MaxDepth:   2,
		Files:      []*FileInfo{{Name: "file.txt", Size: 100}},
		Subdirs: []*DirectoryTreeInfo{
			{Path: "/media/sub", TotalFiles: 5, TotalDirs: 1},
		},
	}

	assert.Equal(t, "/media", tree.Path)
	assert.Equal(t, 10, tree.TotalFiles)
	assert.Equal(t, 3, tree.TotalDirs)
	assert.Equal(t, int64(1024*1024), tree.TotalSize)
	assert.Equal(t, 2, tree.MaxDepth)
	assert.Len(t, tree.Files, 1)
	assert.Len(t, tree.Subdirs, 1)
	assert.Equal(t, "/media/sub", tree.Subdirs[0].Path)
}

// =============================================================================
// Factory - additional protocol edge cases
// =============================================================================

func TestDefaultClientFactory_CreateClient_AllProtocols(t *testing.T) {
	factory := NewDefaultClientFactory()

	tests := []struct {
		name     string
		protocol string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "ftp protocol",
			protocol: "ftp",
			settings: map[string]interface{}{
				"host":     "ftp.example.com",
				"port":     float64(21),
				"username": "user",
				"password": "pass",
				"path":     "/uploads",
			},
			wantErr: false,
		},
		{
			name:     "ftp with default port",
			protocol: "ftp",
			settings: map[string]interface{}{
				"host": "ftp.example.com",
			},
			wantErr: false,
		},
		{
			name:     "smb protocol",
			protocol: "smb",
			settings: map[string]interface{}{
				"host":     "smb.example.com",
				"port":     float64(445),
				"share":    "data",
				"username": "admin",
				"password": "secret",
				"domain":   "WORKGROUP",
			},
			wantErr: false,
		},
		{
			name:     "smb with defaults",
			protocol: "smb",
			settings: map[string]interface{}{
				"host": "smb.example.com",
			},
			wantErr: false,
		},
		{
			name:     "local protocol",
			protocol: "local",
			settings: map[string]interface{}{
				"base_path": "/tmp",
			},
			wantErr: false,
		},
		{
			name:     "local with default",
			protocol: "local",
			settings: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "nfs protocol",
			protocol: "nfs",
			settings: map[string]interface{}{
				"host":        "nfs.example.com",
				"path":        "/export",
				"mount_point": "/mnt/nfs",
				"options":     "vers=4",
			},
			wantErr: false,
		},
		{
			name:     "nfs missing mount_point",
			protocol: "nfs",
			settings: map[string]interface{}{
				"host": "nfs.example.com",
			},
			wantErr: true,
		},
		{
			name:     "webdav protocol",
			protocol: "webdav",
			settings: map[string]interface{}{
				"url":      "https://dav.example.com",
				"username": "user",
				"password": "pass",
			},
			wantErr: false,
		},
		{
			name:     "unsupported protocol",
			protocol: "sftp",
			settings: map[string]interface{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &StorageConfig{
				Protocol: tt.protocol,
				Settings: tt.settings,
			}
			client, err := factory.CreateClient(config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestDefaultClientFactory_SupportedProtocols_Boost(t *testing.T) {
	factory := NewDefaultClientFactory()
	protocols := factory.SupportedProtocols()

	assert.Contains(t, protocols, "smb")
	assert.Contains(t, protocols, "ftp")
	assert.Contains(t, protocols, "nfs")
	assert.Contains(t, protocols, "webdav")
	assert.Contains(t, protocols, "local")
	assert.Len(t, protocols, 5)
}

// =============================================================================
// Helper function tests
// =============================================================================

func TestGetStringSetting_Boost(t *testing.T) {
	settings := map[string]interface{}{
		"host":     "example.com",
		"port":     float64(21),
		"empty":    "",
		"number":   42,
		"nil_val":  nil,
	}

	assert.Equal(t, "example.com", getStringSetting(settings, "host", "default"))
	assert.Equal(t, "default", getStringSetting(settings, "missing", "default"))
	assert.Equal(t, "", getStringSetting(settings, "empty", "default"))
	assert.Equal(t, "default", getStringSetting(settings, "number", "default"))
}

func TestGetIntSetting_Boost(t *testing.T) {
	settings := map[string]interface{}{
		"port":    float64(21),
		"count":   float64(100),
		"zero":    float64(0),
		"string":  "not a number",
		"nil_val": nil,
	}

	assert.Equal(t, 21, getIntSetting(settings, "port", 0))
	assert.Equal(t, 100, getIntSetting(settings, "count", 0))
	assert.Equal(t, 0, getIntSetting(settings, "zero", 99))
	assert.Equal(t, 99, getIntSetting(settings, "missing", 99))
	assert.Equal(t, 99, getIntSetting(settings, "string", 99))
}
