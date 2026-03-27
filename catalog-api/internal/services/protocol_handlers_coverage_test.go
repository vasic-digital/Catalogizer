package services

import (
	"catalogizer/filesystem"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockFSClient implements filesystem.FileSystemClient for protocol handler tests
type mockFSClient struct {
	fileExistsFn      func(ctx context.Context, path string) (bool, error)
	copyFileFn        func(ctx context.Context, src, dst string) error
	deleteFileFn      func(ctx context.Context, path string) error
	listDirectoryFn   func(ctx context.Context, path string) ([]*filesystem.FileInfo, error)
	createDirectoryFn func(ctx context.Context, path string) error
	deleteDirectoryFn func(ctx context.Context, path string) error
}

func (m *mockFSClient) Connect(ctx context.Context) error              { return nil }
func (m *mockFSClient) Disconnect(ctx context.Context) error           { return nil }
func (m *mockFSClient) IsConnected() bool                              { return true }
func (m *mockFSClient) TestConnection(ctx context.Context) error       { return nil }
func (m *mockFSClient) GetProtocol() string                            { return "mock" }
func (m *mockFSClient) GetConfig() interface{}                         { return nil }
func (m *mockFSClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFSClient) WriteFile(ctx context.Context, path string, data io.Reader) error {
	return errors.New("not implemented")
}
func (m *mockFSClient) GetFileInfo(ctx context.Context, path string) (*filesystem.FileInfo, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFSClient) FileExists(ctx context.Context, path string) (bool, error) {
	if m.fileExistsFn != nil {
		return m.fileExistsFn(ctx, path)
	}
	return false, nil
}
func (m *mockFSClient) DeleteFile(ctx context.Context, path string) error {
	if m.deleteFileFn != nil {
		return m.deleteFileFn(ctx, path)
	}
	return nil
}
func (m *mockFSClient) CopyFile(ctx context.Context, src, dst string) error {
	if m.copyFileFn != nil {
		return m.copyFileFn(ctx, src, dst)
	}
	return nil
}
func (m *mockFSClient) ListDirectory(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
	if m.listDirectoryFn != nil {
		return m.listDirectoryFn(ctx, path)
	}
	return nil, nil
}
func (m *mockFSClient) CreateDirectory(ctx context.Context, path string) error {
	if m.createDirectoryFn != nil {
		return m.createDirectoryFn(ctx, path)
	}
	return nil
}
func (m *mockFSClient) DeleteDirectory(ctx context.Context, path string) error {
	if m.deleteDirectoryFn != nil {
		return m.deleteDirectoryFn(ctx, path)
	}
	return nil
}

// =============================================================================
// LocalProtocolHandler
// =============================================================================

func TestLocalProtocolHandler_GetFileIdentifier(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	id, err := h.GetFileIdentifier(context.Background(), "/test/file.mp4", 1024, false)
	assert.NoError(t, err)
	assert.Equal(t, "local:1024:false", id)

	id2, err := h.GetFileIdentifier(context.Background(), "/test/dir", 0, true)
	assert.NoError(t, err)
	assert.Equal(t, "local:0:true", id2)
}

func TestLocalProtocolHandler_PerformMove(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	err := h.PerformMove(context.Background(), nil, "/old", "/new", false)
	assert.NoError(t, err)
}

func TestLocalProtocolHandler_ValidateMove_SamePath(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	err := h.ValidateMove(context.Background(), nil, "/same", "/same")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same")
}

func TestLocalProtocolHandler_ValidateMove_DestExists(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) {
			return true, nil
		},
	}
	err := h.ValidateMove(context.Background(), client, "/old", "/new")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestLocalProtocolHandler_ValidateMove_CheckError(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) {
			return false, errors.New("check failed")
		},
	}
	err := h.ValidateMove(context.Background(), client, "/old", "/new")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check")
}

func TestLocalProtocolHandler_ValidateMove_Success(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) {
			return false, nil
		},
	}
	err := h.ValidateMove(context.Background(), client, "/old", "/new")
	assert.NoError(t, err)
}

func TestLocalProtocolHandler_GetMoveWindow(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	assert.Equal(t, 2*time.Second, h.GetMoveWindow())
}

func TestLocalProtocolHandler_SupportsRealTimeNotification(t *testing.T) {
	h := NewLocalProtocolHandler(zap.NewNop())
	assert.True(t, h.SupportsRealTimeNotification())
}

// =============================================================================
// SMBProtocolHandler
// =============================================================================

func TestSMBProtocolHandler_GetFileIdentifier(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	id, err := h.GetFileIdentifier(context.Background(), "/path", 2048, false)
	assert.NoError(t, err)
	assert.Equal(t, "smb:2048:false", id)
}

func TestSMBProtocolHandler_PerformMove_File(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		copyFileFn:   func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old.txt", "/new.txt", false)
	assert.NoError(t, err)
}

func TestSMBProtocolHandler_PerformMove_File_CopyFails(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		copyFileFn: func(ctx context.Context, src, dst string) error {
			return errors.New("copy failed")
		},
	}
	err := h.PerformMove(context.Background(), client, "/old.txt", "/new.txt", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "copy")
}

func TestSMBProtocolHandler_PerformMove_File_DeleteFails(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		copyFileFn: func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn: func(ctx context.Context, path string) error {
			return errors.New("delete failed")
		},
	}
	err := h.PerformMove(context.Background(), client, "/old.txt", "/new.txt", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete original")
}

func TestSMBProtocolHandler_PerformMove_Directory(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			return []*filesystem.FileInfo{
				{Name: "file1.txt", IsDir: false},
			}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		copyFileFn:        func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn:      func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/olddir", "/newdir", true)
	assert.NoError(t, err)
}

func TestSMBProtocolHandler_PerformMove_Directory_ListError(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			return nil, errors.New("list failed")
		},
	}
	err := h.PerformMove(context.Background(), client, "/olddir", "/newdir", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list directory")
}

func TestSMBProtocolHandler_ValidateMove(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())

	// Same path
	err := h.ValidateMove(context.Background(), nil, "/a", "/a")
	assert.Error(t, err)

	// Dest exists
	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) { return true, nil },
	}
	err = h.ValidateMove(context.Background(), client, "/a", "/b")
	assert.Error(t, err)

	// Success
	client2 := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) { return false, nil },
	}
	err = h.ValidateMove(context.Background(), client2, "/a", "/b")
	assert.NoError(t, err)
}

func TestSMBProtocolHandler_GetMoveWindow(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	assert.Equal(t, 10*time.Second, h.GetMoveWindow())
}

func TestSMBProtocolHandler_SupportsRealTimeNotification(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	assert.False(t, h.SupportsRealTimeNotification())
}

// =============================================================================
// FTPProtocolHandler
// =============================================================================

func TestFTPProtocolHandler_GetFileIdentifier(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	id, err := h.GetFileIdentifier(context.Background(), "/path", 4096, true)
	assert.NoError(t, err)
	assert.Equal(t, "ftp:4096:true", id)
}

func TestFTPProtocolHandler_PerformMove_File(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	client := &mockFSClient{}
	err := h.PerformMove(context.Background(), client, "/old", "/new", false)
	assert.NoError(t, err)
}

func TestFTPProtocolHandler_PerformMove_File_CopyFails(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		copyFileFn: func(ctx context.Context, src, dst string) error {
			return errors.New("ftp copy failed")
		},
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", false)
	assert.Error(t, err)
}

func TestFTPProtocolHandler_PerformMove_Directory(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			return []*filesystem.FileInfo{}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.NoError(t, err)
}

func TestFTPProtocolHandler_ValidateMove(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	err := h.ValidateMove(context.Background(), nil, "/a", "/a")
	assert.Error(t, err)

	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) { return false, nil },
	}
	err = h.ValidateMove(context.Background(), client, "/a", "/b")
	assert.NoError(t, err)
}

func TestFTPProtocolHandler_GetMoveWindow(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	assert.Equal(t, 30*time.Second, h.GetMoveWindow())
}

func TestFTPProtocolHandler_SupportsRealTimeNotification(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	assert.False(t, h.SupportsRealTimeNotification())
}

// =============================================================================
// NFSProtocolHandler
// =============================================================================

func TestNFSProtocolHandler_GetFileIdentifier(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	id, err := h.GetFileIdentifier(context.Background(), "/nfs/path", 8192, false)
	assert.NoError(t, err)
	assert.Equal(t, "nfs:8192:false", id)
}

func TestNFSProtocolHandler_PerformMove_File(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	client := &mockFSClient{}
	err := h.PerformMove(context.Background(), client, "/old", "/new", false)
	assert.NoError(t, err)
}

func TestNFSProtocolHandler_PerformMove_Directory(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			if path == "/old" {
				return []*filesystem.FileInfo{
					{Name: "sub", IsDir: true},
				}, nil
			}
			return []*filesystem.FileInfo{}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.NoError(t, err)
}

func TestNFSProtocolHandler_ValidateMove(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	err := h.ValidateMove(context.Background(), nil, "/x", "/x")
	assert.Error(t, err)

	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) { return true, nil },
	}
	err = h.ValidateMove(context.Background(), client, "/a", "/b")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestNFSProtocolHandler_GetMoveWindow(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	assert.Equal(t, 5*time.Second, h.GetMoveWindow())
}

func TestNFSProtocolHandler_SupportsRealTimeNotification(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	assert.False(t, h.SupportsRealTimeNotification())
}

// =============================================================================
// WebDAVProtocolHandler
// =============================================================================

func TestWebDAVProtocolHandler_GetFileIdentifier(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	id, err := h.GetFileIdentifier(context.Background(), "/dav", 512, true)
	assert.NoError(t, err)
	assert.Equal(t, "webdav:512:true", id)
}

func TestWebDAVProtocolHandler_PerformMove_File(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	client := &mockFSClient{}
	err := h.PerformMove(context.Background(), client, "/old", "/new", false)
	assert.NoError(t, err)
}

func TestWebDAVProtocolHandler_PerformMove_Directory(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			return []*filesystem.FileInfo{}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.NoError(t, err)
}

func TestWebDAVProtocolHandler_ValidateMove(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	err := h.ValidateMove(context.Background(), nil, "/same", "/same")
	assert.Error(t, err)

	client := &mockFSClient{
		fileExistsFn: func(ctx context.Context, path string) (bool, error) {
			return false, errors.New("check err")
		},
	}
	err = h.ValidateMove(context.Background(), client, "/a", "/b")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check")
}

func TestWebDAVProtocolHandler_GetMoveWindow(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	assert.Equal(t, 15*time.Second, h.GetMoveWindow())
}

func TestWebDAVProtocolHandler_SupportsRealTimeNotification(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	assert.False(t, h.SupportsRealTimeNotification())
}

// =============================================================================
// ProtocolHandlerFactory
// =============================================================================

func TestProtocolHandlerFactory_CreateHandler_AllProtocols(t *testing.T) {
	factory := NewProtocolHandlerFactory(zap.NewNop())

	for _, proto := range []string{"local", "smb", "ftp", "nfs", "webdav"} {
		handler, err := factory.CreateHandler(proto)
		assert.NoError(t, err, proto)
		assert.NotNil(t, handler, proto)
	}
}

func TestProtocolHandlerFactory_CreateHandler_CaseInsensitive(t *testing.T) {
	factory := NewProtocolHandlerFactory(zap.NewNop())

	for _, proto := range []string{"LOCAL", "SMB", "FTP", "NFS", "WEBDAV"} {
		handler, err := factory.CreateHandler(proto)
		assert.NoError(t, err, proto)
		assert.NotNil(t, handler, proto)
	}
}

func TestProtocolHandlerFactory_CreateHandler_Unknown(t *testing.T) {
	factory := NewProtocolHandlerFactory(zap.NewNop())
	handler, err := factory.CreateHandler("s3")
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.Contains(t, err.Error(), "unsupported protocol")
}

func TestProtocolHandlerFactory_GetSupportedProtocols_AllFive(t *testing.T) {
	factory := NewProtocolHandlerFactory(zap.NewNop())
	protocols := factory.GetSupportedProtocols()
	assert.Len(t, protocols, 5)
	assert.Contains(t, protocols, "local")
	assert.Contains(t, protocols, "smb")
	assert.Contains(t, protocols, "ftp")
	assert.Contains(t, protocols, "nfs")
	assert.Contains(t, protocols, "webdav")
}

// =============================================================================
// GetProtocolCapabilities
// =============================================================================

func TestGetProtocolCapabilities_Local(t *testing.T) {
	caps, err := GetProtocolCapabilities("local", zap.NewNop())
	require.NoError(t, err)
	assert.Equal(t, "local", caps.Protocol)
	assert.True(t, caps.SupportsRealTimeNotification)
	assert.True(t, caps.SupportsAtomicMove)
	assert.False(t, caps.RequiresPolling)
	assert.Equal(t, 2*time.Second, caps.MoveWindow)
}

func TestGetProtocolCapabilities_SMB(t *testing.T) {
	caps, err := GetProtocolCapabilities("smb", zap.NewNop())
	require.NoError(t, err)
	assert.Equal(t, "smb", caps.Protocol)
	assert.False(t, caps.SupportsRealTimeNotification)
	assert.False(t, caps.SupportsAtomicMove)
	assert.True(t, caps.RequiresPolling)
}

func TestGetProtocolCapabilities_NFS(t *testing.T) {
	caps, err := GetProtocolCapabilities("nfs", zap.NewNop())
	require.NoError(t, err)
	assert.True(t, caps.SupportsAtomicMove)
}

func TestGetProtocolCapabilities_Unknown(t *testing.T) {
	_, err := GetProtocolCapabilities("s3", zap.NewNop())
	assert.Error(t, err)
}

// =============================================================================
// Nested directory moves with subdirectories
// =============================================================================

func TestSMBProtocolHandler_MoveDirectory_WithSubdirs(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	callCount := 0
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			callCount++
			if callCount == 1 {
				return []*filesystem.FileInfo{
					{Name: "subdir", IsDir: true},
					{Name: "file.txt", IsDir: false},
				}, nil
			}
			return []*filesystem.FileInfo{}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		copyFileFn:        func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn:      func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.NoError(t, err)
}

func TestFTPProtocolHandler_MoveDirectory_WithFiles(t *testing.T) {
	h := NewFTPProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			if path == "/old" {
				return []*filesystem.FileInfo{
					{Name: "file1.txt", IsDir: false},
					{Name: "file2.txt", IsDir: false},
				}, nil
			}
			return nil, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		copyFileFn:        func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn:      func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.NoError(t, err)
}

func TestWebDAVProtocolHandler_MoveDirectory_WithNestedContent(t *testing.T) {
	h := NewWebDAVProtocolHandler(zap.NewNop())
	callCount := 0
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			callCount++
			if callCount == 1 {
				return []*filesystem.FileInfo{
					{Name: "nested", IsDir: true},
				}, nil
			}
			return []*filesystem.FileInfo{
				{Name: "deep.txt", IsDir: false},
			}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error { return nil },
		copyFileFn:        func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn:      func(ctx context.Context, path string) error { return nil },
		deleteDirectoryFn: func(ctx context.Context, path string) error { return nil },
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.NoError(t, err)
}

// Error cases for directory moves
func TestSMBProtocolHandler_MoveDirectory_CreateDirFails(t *testing.T) {
	h := NewSMBProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		listDirectoryFn: func(ctx context.Context, path string) ([]*filesystem.FileInfo, error) {
			return []*filesystem.FileInfo{}, nil
		},
		createDirectoryFn: func(ctx context.Context, path string) error {
			return errors.New("create dir failed")
		},
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create destination directory")
}

func TestNFSProtocolHandler_MoveFile_DeleteFails(t *testing.T) {
	h := NewNFSProtocolHandler(zap.NewNop())
	client := &mockFSClient{
		copyFileFn: func(ctx context.Context, src, dst string) error { return nil },
		deleteFileFn: func(ctx context.Context, path string) error {
			return errors.New("delete failed")
		},
	}
	err := h.PerformMove(context.Background(), client, "/old", "/new", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete original")
}
