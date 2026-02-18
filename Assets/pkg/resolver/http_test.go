package resolver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPResolver_CanResolve(t *testing.T) {
	r := NewHTTPResolver(1)

	tests := []struct {
		hint string
		can  bool
	}{
		{"https://example.com/image.jpg", true},
		{"http://example.com/image.jpg", true},
		{"/local/path", false},
		{"ftp://example.com/file", false},
		{"", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.can, r.CanResolve(context.Background(), &ResolveRequest{SourceHint: tt.hint}))
	}
}

func TestHTTPResolver_Resolve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("fake-image-data"))
	}))
	defer server.Close()

	r := NewHTTPResolver(1)
	req := &ResolveRequest{
		AssetID:    asset.NewID(),
		SourceHint: server.URL + "/image.jpg",
	}

	result, err := r.Resolve(context.Background(), req)
	require.NoError(t, err)
	defer result.Content.Close()

	data, err := io.ReadAll(result.Content)
	require.NoError(t, err)
	assert.Equal(t, "fake-image-data", string(data))
	assert.Equal(t, "image/jpeg", result.ContentType)
}

func TestHTTPResolver_Resolve404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	r := NewHTTPResolver(1)
	req := &ResolveRequest{
		AssetID:    asset.NewID(),
		SourceHint: server.URL + "/notfound",
	}

	_, err := r.Resolve(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
