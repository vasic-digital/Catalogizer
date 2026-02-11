//go:build windows
// +build windows

package filesystem

import (
	"context"
	"fmt"
	"io"
)

// NFSConfig contains NFS connection configuration
type NFSConfig struct {
	Host       string `json:"host"`
	Path       string `json:"path"`        // Export path on NFS server
	MountPoint string `json:"mount_point"` // Local mount point
	Options    string `json:"options"`     // Mount options
}

// NFSClient for Windows (NFS is not natively supported)
type NFSClient struct {
	config     NFSConfig
	mountPoint string
	connected  bool
}

func NewNFSClient(config NFSConfig) (*NFSClient, error) {
	return &NFSClient{
		config:     config,
		mountPoint: config.MountPoint,
	}, nil
}

func (c *NFSClient) Connect(ctx context.Context) error {
	return fmt.Errorf("NFS mounting not supported on Windows platform")
}

func (c *NFSClient) Disconnect(ctx context.Context) error {
	c.connected = false
	return nil
}

func (c *NFSClient) TestConnection(ctx context.Context) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) ListDirectory(ctx context.Context, path string) ([]*FileInfo, error) {
	return nil, fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	return nil, fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) CreateDirectory(ctx context.Context, path string) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) DeleteDirectory(ctx context.Context, path string) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) DeleteFile(ctx context.Context, path string) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) CopyFile(ctx context.Context, src, dst string) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) MoveFile(ctx context.Context, src, dst string) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) WriteFile(ctx context.Context, path string, data io.Reader) error {
	return fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) FileExists(ctx context.Context, path string) (bool, error) {
	return false, fmt.Errorf("NFS not supported on Windows platform")
}

func (c *NFSClient) IsConnected() bool {
	return c.connected
}

func (c *NFSClient) GetRootPath() string {
	return c.mountPoint
}

func (c *NFSClient) GetProtocol() string {
	return "nfs"
}

func (c *NFSClient) GetConfig() interface{} {
	return c.config
}
