package services

import (
	"catalog-api/filesystem"
	"catalog-api/internal/config"
	"context"

	"go.uber.org/zap"
)

type FileSystemService struct {
	config  *config.Config
	logger  *zap.Logger
	factory filesystem.ClientFactory
}

func NewFileSystemService(cfg *config.Config, logger *zap.Logger) *FileSystemService {
	return &FileSystemService{
		config:  cfg,
		logger:  logger,
		factory: filesystem.NewDefaultClientFactory(),
	}
}

func (fs *FileSystemService) GetClient(config *filesystem.StorageConfig) (filesystem.FileSystemClient, error) {
	return fs.factory.CreateClient(config)
}

func (fs *FileSystemService) ListFiles(ctx context.Context, client filesystem.FileSystemClient, path string) ([]*filesystem.FileInfo, error) {
	if !client.IsConnected() {
		if err := client.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return client.ListDirectory(ctx, path)
}

func (fs *FileSystemService) GetFileInfo(ctx context.Context, client filesystem.FileSystemClient, path string) (*filesystem.FileInfo, error) {
	if !client.IsConnected() {
		if err := client.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return client.GetFileInfo(ctx, path)
}
