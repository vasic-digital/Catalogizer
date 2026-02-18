package resolver

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalFileResolver_CanResolve(t *testing.T) {
	r := NewLocalFileResolver(1)

	// Non-absolute path
	assert.False(t, r.CanResolve(context.Background(), &ResolveRequest{SourceHint: "relative/path"}))

	// Non-existent file
	assert.False(t, r.CanResolve(context.Background(), &ResolveRequest{SourceHint: "/nonexistent/file"}))

	// Create a temp file
	dir := t.TempDir()
	f := filepath.Join(dir, "test.jpg")
	require.NoError(t, os.WriteFile(f, []byte("jpeg-data"), 0644))

	assert.True(t, r.CanResolve(context.Background(), &ResolveRequest{SourceHint: f}))

	// Directory should not resolve
	assert.False(t, r.CanResolve(context.Background(), &ResolveRequest{SourceHint: dir}))
}

func TestLocalFileResolver_Resolve(t *testing.T) {
	r := NewLocalFileResolver(1)
	dir := t.TempDir()

	content := []byte("local image content")
	f := filepath.Join(dir, "cover.png")
	require.NoError(t, os.WriteFile(f, content, 0644))

	req := &ResolveRequest{
		AssetID:    asset.NewID(),
		SourceHint: f,
	}

	result, err := r.Resolve(context.Background(), req)
	require.NoError(t, err)
	defer result.Content.Close()

	data, err := io.ReadAll(result.Content)
	require.NoError(t, err)
	assert.Equal(t, content, data)
	assert.Equal(t, "image/png", result.ContentType)
	assert.Equal(t, int64(len(content)), result.Size)
}

func TestLocalFileResolver_ResolveNotFound(t *testing.T) {
	r := NewLocalFileResolver(1)
	req := &ResolveRequest{
		AssetID:    asset.NewID(),
		SourceHint: "/nonexistent/file.jpg",
	}

	_, err := r.Resolve(context.Background(), req)
	assert.Error(t, err)
}
