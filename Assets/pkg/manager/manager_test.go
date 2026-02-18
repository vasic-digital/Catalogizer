package manager

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"digital.vasic.assets/pkg/asset"
	"digital.vasic.assets/pkg/event"
	"digital.vasic.assets/pkg/resolver"
	"digital.vasic.assets/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_GetReturnsDefault_WhenNotResolved(t *testing.T) {
	memStore := store.NewMemoryStore()
	mgr := New(WithStore(memStore))
	defer mgr.Stop()

	content, info, isDefault, err := mgr.Get(context.Background(), "nonexistent")
	require.NoError(t, err)
	defer content.Close()

	assert.True(t, isDefault)
	assert.Equal(t, "image/png", info.ContentType)
}

func TestManager_GetReturnsContent_WhenResolved(t *testing.T) {
	memStore := store.NewMemoryStore()
	id := asset.NewID()

	// Pre-populate store
	err := memStore.Put(context.Background(), id, bytes.NewReader([]byte("real-image")),
		&store.Info{ContentType: "image/jpeg", Size: 10})
	require.NoError(t, err)

	mgr := New(WithStore(memStore))
	defer mgr.Stop()

	content, info, isDefault, err := mgr.Get(context.Background(), id)
	require.NoError(t, err)
	defer content.Close()

	assert.False(t, isDefault)
	assert.Equal(t, "image/jpeg", info.ContentType)

	data, _ := io.ReadAll(content)
	assert.Equal(t, "real-image", string(data))
}

func TestManager_RequestAndResolve(t *testing.T) {
	memStore := store.NewMemoryStore()
	bus := event.NewInMemoryBus()

	readyCh := make(chan event.Event, 1)
	bus.Subscribe(func(evt event.Event) {
		if evt.Type == event.AssetReady {
			readyCh <- evt
		}
	})

	mock := &mockResolver{
		canResolve: true,
		result: &resolver.ResolveResult{
			Content:     io.NopCloser(strings.NewReader("resolved-content")),
			ContentType: "image/png",
			Size:        16,
		},
	}
	chain := resolver.NewChain(mock)

	mgr := New(
		WithStore(memStore),
		WithResolver(chain),
		WithEventBus(bus),
		WithWorkers(1),
	)
	defer mgr.Stop()

	req := &resolver.ResolveRequest{
		AssetType:  asset.TypeImage,
		SourceHint: "https://example.com/image.png",
		EntityType: "file",
		EntityID:   "42",
	}

	id, err := mgr.Request(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// Wait for resolution
	select {
	case evt := <-readyCh:
		assert.Equal(t, event.AssetReady, evt.Type)
		assert.Equal(t, id, evt.AssetID)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for asset resolution")
	}

	// Now Get should return the resolved content
	content, info, isDefault, err := mgr.Get(context.Background(), id)
	require.NoError(t, err)
	defer content.Close()

	assert.False(t, isDefault)
	assert.Equal(t, "image/png", info.ContentType)
}

func TestManager_Invalidate(t *testing.T) {
	memStore := store.NewMemoryStore()
	bus := event.NewInMemoryBus()
	id := asset.NewID()

	// Pre-populate
	err := memStore.Put(context.Background(), id, bytes.NewReader([]byte("old-content")),
		&store.Info{ContentType: "image/jpeg"})
	require.NoError(t, err)

	invalidatedCh := make(chan event.Event, 1)
	bus.Subscribe(func(evt event.Event) {
		if evt.Type == event.AssetInvalidated {
			invalidatedCh <- evt
		}
	})

	mgr := New(WithStore(memStore), WithEventBus(bus))
	defer mgr.Stop()

	err = mgr.Invalidate(context.Background(), id, nil)
	require.NoError(t, err)

	// Content should be deleted
	exists, _ := memStore.Exists(context.Background(), id)
	assert.False(t, exists)

	// Invalidated event published
	select {
	case evt := <-invalidatedCh:
		assert.Equal(t, event.AssetInvalidated, evt.Type)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for invalidation event")
	}
}

// mockResolver for manager tests
type mockResolver struct {
	canResolve bool
	result     *resolver.ResolveResult
	err        error
}

func (m *mockResolver) Name() string                                                     { return "mock" }
func (m *mockResolver) Priority() int                                                    { return 1 }
func (m *mockResolver) CanResolve(_ context.Context, _ *resolver.ResolveRequest) bool    { return m.canResolve }
func (m *mockResolver) Resolve(_ context.Context, _ *resolver.ResolveRequest) (*resolver.ResolveResult, error) {
	return m.result, m.err
}
