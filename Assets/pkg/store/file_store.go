package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"digital.vasic.assets/pkg/asset"
)

// FileStore is a filesystem-backed Store implementation.
type FileStore struct {
	baseDir string
}

// NewFileStore creates a new FileStore at the given directory.
func NewFileStore(baseDir string) (*FileStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}
	return &FileStore{baseDir: baseDir}, nil
}

func (s *FileStore) contentPath(id asset.ID) string {
	return filepath.Join(s.baseDir, id.String())
}

func (s *FileStore) metaPath(id asset.ID) string {
	return filepath.Join(s.baseDir, id.String()+".meta")
}

// Get retrieves asset content and info from the filesystem.
func (s *FileStore) Get(_ context.Context, id asset.ID) (io.ReadCloser, *Info, error) {
	info, err := s.readMeta(id)
	if err != nil {
		return nil, nil, fmt.Errorf("read metadata: %w", err)
	}

	f, err := os.Open(s.contentPath(id))
	if err != nil {
		return nil, nil, fmt.Errorf("open content: %w", err)
	}

	return f, info, nil
}

// Put stores asset content and info to the filesystem.
func (s *FileStore) Put(_ context.Context, id asset.ID, content io.Reader, info *Info) error {
	f, err := os.Create(s.contentPath(id))
	if err != nil {
		return fmt.Errorf("create content file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, content)
	if err != nil {
		os.Remove(s.contentPath(id))
		return fmt.Errorf("write content: %w", err)
	}

	meta := &Info{
		ContentType: info.ContentType,
		Size:        written,
	}
	if err := s.writeMeta(id, meta); err != nil {
		os.Remove(s.contentPath(id))
		return fmt.Errorf("write metadata: %w", err)
	}

	return nil
}

// Delete removes asset content and info from the filesystem.
func (s *FileStore) Delete(_ context.Context, id asset.ID) error {
	os.Remove(s.metaPath(id))
	if err := os.Remove(s.contentPath(id)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete content: %w", err)
	}
	return nil
}

// Exists checks if asset content exists in the filesystem.
func (s *FileStore) Exists(_ context.Context, id asset.ID) (bool, error) {
	_, err := os.Stat(s.contentPath(id))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *FileStore) readMeta(id asset.ID) (*Info, error) {
	data, err := os.ReadFile(s.metaPath(id))
	if err != nil {
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *FileStore) writeMeta(id asset.ID, info *Info) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(id), data, 0644)
}
