package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"digital.vasic.assets/pkg/asset"
)

type memEntry struct {
	data []byte
	info Info
}

// MemoryStore is an in-memory Store implementation for testing.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[asset.ID]*memEntry
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make(map[asset.ID]*memEntry),
	}
}

// Get retrieves asset content from memory.
func (s *MemoryStore) Get(_ context.Context, id asset.ID) (io.ReadCloser, *Info, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[id]
	if !ok {
		return nil, nil, fmt.Errorf("asset not found: %s", id)
	}

	info := entry.info
	return io.NopCloser(bytes.NewReader(entry.data)), &info, nil
}

// Put stores asset content in memory.
func (s *MemoryStore) Put(_ context.Context, id asset.ID, content io.Reader, info *Info) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("read content: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[id] = &memEntry{
		data: data,
		info: Info{
			ContentType: info.ContentType,
			Size:        int64(len(data)),
		},
	}
	return nil
}

// Delete removes asset content from memory.
func (s *MemoryStore) Delete(_ context.Context, id asset.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, id)
	return nil
}

// Exists checks if asset content exists in memory.
func (s *MemoryStore) Exists(_ context.Context, id asset.ID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.entries[id]
	return ok, nil
}
