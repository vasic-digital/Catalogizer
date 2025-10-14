package services

import (
	"bytes"
	"catalogizer/filesystem"
	"catalogizer/internal/config"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewFileSystemService(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()

	service := NewFileSystemService(cfg, logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.factory)
	assert.Equal(t, cfg, service.config)
	assert.Equal(t, logger, service.logger)
}

func TestFileSystemService_GetClient(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	tests := []struct {
		name    string
		config  *filesystem.StorageConfig
		wantErr bool
	}{
		{
			name: "Create local client",
			config: &filesystem.StorageConfig{
				ID:       "test-local",
				Name:     "Test Local Storage",
				Protocol: "local",
				Enabled:  true,
				Settings: map[string]interface{}{
					"base_path": "/tmp",
				},
			},
			wantErr: false,
		},
		{
			name: "Create SMB client",
			config: &filesystem.StorageConfig{
				ID:       "test-smb",
				Name:     "Test SMB Storage",
				Protocol: "smb",
				Enabled:  true,
				Settings: map[string]interface{}{
					"host":     "localhost",
					"port":     445,
					"share":    "test",
					"username": "testuser",
					"password": "testpass",
					"domain":   "WORKGROUP",
				},
			},
			wantErr: false,
		},
		{
			name: "Create FTP client",
			config: &filesystem.StorageConfig{
				ID:       "test-ftp",
				Name:     "Test FTP Storage",
				Protocol: "ftp",
				Enabled:  true,
				Settings: map[string]interface{}{
					"host":     "localhost",
					"port":     21,
					"username": "testuser",
					"password": "testpass",
					"path":     "/",
				},
			},
			wantErr: false,
		},
		{
			name: "Create NFS client",
			config: &filesystem.StorageConfig{
				ID:       "test-nfs",
				Name:     "Test NFS Storage",
				Protocol: "nfs",
				Enabled:  true,
				Settings: map[string]interface{}{
					"host":        "localhost",
					"path":        "/export",
					"mount_point": "/mnt/nfs",
					"options":     "vers=3",
				},
			},
			wantErr: false,
		},
		{
			name: "Create WebDAV client",
			config: &filesystem.StorageConfig{
				ID:       "test-webdav",
				Name:     "Test WebDAV Storage",
				Protocol: "webdav",
				Enabled:  true,
				Settings: map[string]interface{}{
					"url":      "http://localhost/webdav",
					"username": "testuser",
					"password": "testpass",
					"path":     "/",
				},
			},
			wantErr: false,
		},
		{
			name: "Unsupported protocol",
			config: &filesystem.StorageConfig{
				ID:       "test-invalid",
				Name:     "Test Invalid Storage",
				Protocol: "unsupported",
				Enabled:  true,
				Settings: map[string]interface{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := service.GetClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.config.Protocol, client.GetProtocol())
			}
		})
	}
}

func TestFileSystemService_ListFiles(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create a local client for testing
	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})

	// Test listing empty directory
	files, err := service.ListFiles(context.Background(), client, "")
	assert.NoError(t, err)
	if files == nil {
		files = []*filesystem.FileInfo{}
	}
	assert.Equal(t, 0, len(files))

	// Test that the client is connected after ListFiles
	assert.True(t, client.IsConnected())
}

func TestFileSystemService_ListFiles_WithDisconnectedClient(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create a local client but don't connect it
	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})

	// ListFiles should connect the client automatically
	assert.False(t, client.IsConnected())
	files, err := service.ListFiles(context.Background(), client, "")
	assert.NoError(t, err)
	if files == nil {
		files = []*filesystem.FileInfo{}
	}
	assert.True(t, client.IsConnected())
}

func TestFileSystemService_GetFileInfo(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	// Create a temporary test directory with a file
	tmpDir := t.TempDir()

	// Create a test file
	testFile := "test.txt"
	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	err := client.Connect(context.Background())
	assert.NoError(t, err)

	// Create the test file
	ctx := context.Background()
	err = client.WriteFile(ctx, testFile, bytes.NewBufferString("test content"))
	assert.NoError(t, err)

	// Get file info
	info, err := service.GetFileInfo(ctx, client, testFile)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, testFile, info.Name)
	assert.False(t, info.IsDir)
}

func TestFileSystemService_GetFileInfo_NonExistentFile(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create a local client for testing
	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	err := client.Connect(context.Background())
	assert.NoError(t, err)

	// Try to get info for non-existent file
	_, err = service.GetFileInfo(context.Background(), client, "nonexistent.txt")
	assert.Error(t, err)
}

func TestFileSystemService_GetFileInfo_WithDisconnectedClient(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create a local client but don't connect it
	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})

	// GetFileInfo should connect the client automatically
	assert.False(t, client.IsConnected())
	_, err := service.GetFileInfo(context.Background(), client, "test.txt")
	assert.True(t, client.IsConnected())
	// Error is expected because file doesn't exist, but client should be connected
	assert.Error(t, err)
}

func TestFileSystemService_Integration(t *testing.T) {
	cfg := &config.Config{}
	logger := zap.NewNop()
	service := NewFileSystemService(cfg, logger)

	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create a storage config
	storageConfig := &filesystem.StorageConfig{
		ID:       "test-integration",
		Name:     "Test Integration Storage",
		Protocol: "local",
		Enabled:  true,
		Settings: map[string]interface{}{
			"base_path": tmpDir,
		},
	}

	// Get a client
	client, err := service.GetClient(storageConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// List files (should auto-connect)
	files, err := service.ListFiles(context.Background(), client, "")
	assert.NoError(t, err)
	if files == nil {
		files = []*filesystem.FileInfo{}
	}
	assert.Equal(t, 0, len(files))

	// Client should be connected now
	assert.True(t, client.IsConnected())
}
