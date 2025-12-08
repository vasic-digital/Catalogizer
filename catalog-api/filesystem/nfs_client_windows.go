//go:build windows
// +build windows

package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// NFSClient for Windows (placeholder implementation)
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
	// NFS mounting is not natively supported on Windows
	return fmt.Errorf("NFS mounting not supported on Windows platform")
}

func (c *NFSClient) Disconnect(ctx context.Context) error {
	c.connected = false
	return nil
}

func (c *NFSClient) ListDirectory(ctx context.Context, path string) ([]FileInfo, error) {
	return nil, fmt.Errorf("not connected")
}

func (c *NFSClient) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	return nil, fmt.Errorf("not connected")
}

func (c *NFSClient) CreateDirectory(ctx context.Context, path string) error {
	return fmt.Errorf("not connected")
}

func (c *NFSClient) DeleteFile(ctx context.Context, path string) error {
	return fmt.Errorf("not connected")
}

func (c *NFSClient) CopyFile(ctx context.Context, src, dst string) error {
	return fmt.Errorf("not connected")
}

func (c *NFSClient) MoveFile(ctx context.Context, src, dst string) error {
	return fmt.Errorf("not connected")
}

func (c *NFSClient) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return nil, fmt.Errorf("not connected")
}

func (c *NFSClient) WriteFile(ctx context.Context, path string, data []byte) error {
	return fmt.Errorf("not connected")
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
