package store

import (
	"bytes"
	"context"
	"io"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_CRUD(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileStore(dir)
	require.NoError(t, err)

	ctx := context.Background()
	id := asset.NewID()
	content := []byte("hello asset world")

	// Put
	err = s.Put(ctx, id, bytes.NewReader(content), &Info{
		ContentType: "text/plain",
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
	assert.Equal(t, "text/plain", info.ContentType)
	assert.Equal(t, int64(len(content)), info.Size)

	// Delete
	err = s.Delete(ctx, id)
	require.NoError(t, err)

	exists, err = s.Exists(ctx, id)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFileStore_GetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileStore(dir)
	require.NoError(t, err)

	_, _, err = s.Get(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestFileStore_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileStore(dir)
	require.NoError(t, err)

	err = s.Delete(context.Background(), "nonexistent")
	assert.NoError(t, err)
}
