package store

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_CRUD(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	id := asset.NewID()
	content := []byte("test content")

	// Put
	err := s.Put(ctx, id, bytes.NewReader(content), &Info{
		ContentType: "image/png",
		Size:        int64(len(content)),
	})
	require.NoError(t, err)

	// Exists
	exists, err := s.Exists(ctx, id)
	require.NoError(t, err)
	assert.True(t, exists)

	// Get
	reader, info, err := s.Get(ctx, id)
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, data)
	assert.Equal(t, "image/png", info.ContentType)

	// Delete
	err = s.Delete(ctx, id)
	require.NoError(t, err)

	exists, err = s.Exists(ctx, id)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, _, err := s.Get(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := asset.NewID()
			content := []byte("concurrent test")
			err := s.Put(ctx, id, bytes.NewReader(content), &Info{ContentType: "text/plain"})
			assert.NoError(t, err)

			exists, err := s.Exists(ctx, id)
			assert.NoError(t, err)
			assert.True(t, exists)

			reader, _, err := s.Get(ctx, id)
			assert.NoError(t, err)
			reader.Close()
		}()
	}
	wg.Wait()
}
