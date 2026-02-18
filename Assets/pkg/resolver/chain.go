package resolver

import (
	"context"
	"fmt"
	"sort"
)

// ChainResolver tries multiple resolvers in priority order.
type ChainResolver struct {
	resolvers []Resolver
}

// NewChain creates a ChainResolver from the given resolvers,
// sorted by priority (lowest number tried first).
func NewChain(resolvers ...Resolver) *ChainResolver {
	sorted := make([]Resolver, len(resolvers))
	copy(sorted, resolvers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})
	return &ChainResolver{resolvers: sorted}
}

func (c *ChainResolver) Name() string     { return "chain" }
func (c *ChainResolver) Priority() int    { return 0 }

// CanResolve returns true if any resolver in the chain can resolve.
func (c *ChainResolver) CanResolve(ctx context.Context, req *ResolveRequest) bool {
	for _, r := range c.resolvers {
		if r.CanResolve(ctx, req) {
			return true
		}
	}
	return false
}

// Resolve tries each resolver in priority order until one succeeds.
func (c *ChainResolver) Resolve(ctx context.Context, req *ResolveRequest) (*ResolveResult, error) {
	for _, r := range c.resolvers {
		if !r.CanResolve(ctx, req) {
			continue
		}
		result, err := r.Resolve(ctx, req)
		if err == nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("no resolver could handle asset %s", req.AssetID)
}
