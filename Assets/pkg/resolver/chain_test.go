package resolver

import (
	"context"
	"io"
	"strings"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResolver struct {
	name       string
	priority   int
	canResolve bool
	result     *ResolveResult
	err        error
}

func (m *mockResolver) Name() string                                                     { return m.name }
func (m *mockResolver) Priority() int                                                    { return m.priority }
func (m *mockResolver) CanResolve(_ context.Context, _ *ResolveRequest) bool             { return m.canResolve }
func (m *mockResolver) Resolve(_ context.Context, _ *ResolveRequest) (*ResolveResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestChainResolver_PriorityOrder(t *testing.T) {
	low := &mockResolver{name: "low", priority: 10, canResolve: true, result: &ResolveResult{
		Content: io.NopCloser(strings.NewReader("low")), ContentType: "text/plain",
	}}
	high := &mockResolver{name: "high", priority: 1, canResolve: true, result: &ResolveResult{
		Content: io.NopCloser(strings.NewReader("high")), ContentType: "text/plain",
	}}

	chain := NewChain(low, high)
	req := &ResolveRequest{AssetID: asset.NewID()}

	result, err := chain.Resolve(context.Background(), req)
	require.NoError(t, err)

	data, _ := io.ReadAll(result.Content)
	assert.Equal(t, "high", string(data))
}

func TestChainResolver_SkipsUnable(t *testing.T) {
	unable := &mockResolver{name: "unable", priority: 1, canResolve: false}
	able := &mockResolver{name: "able", priority: 2, canResolve: true, result: &ResolveResult{
		Content: io.NopCloser(strings.NewReader("found")), ContentType: "text/plain",
	}}

	chain := NewChain(unable, able)
	req := &ResolveRequest{AssetID: asset.NewID()}

	result, err := chain.Resolve(context.Background(), req)
	require.NoError(t, err)

	data, _ := io.ReadAll(result.Content)
	assert.Equal(t, "found", string(data))
}

func TestChainResolver_AllFail(t *testing.T) {
	r1 := &mockResolver{name: "r1", priority: 1, canResolve: false}
	r2 := &mockResolver{name: "r2", priority: 2, canResolve: false}

	chain := NewChain(r1, r2)
	req := &ResolveRequest{AssetID: asset.NewID()}

	_, err := chain.Resolve(context.Background(), req)
	assert.Error(t, err)
}

func TestChainResolver_CanResolve(t *testing.T) {
	r1 := &mockResolver{canResolve: false}
	r2 := &mockResolver{canResolve: true}

	chain := NewChain(r1, r2)
	assert.True(t, chain.CanResolve(context.Background(), &ResolveRequest{}))

	chain2 := NewChain(r1)
	assert.False(t, chain2.CanResolve(context.Background(), &ResolveRequest{}))
}
