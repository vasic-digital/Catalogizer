package integration

import (
	"catalogizer/internal/media/models"
	"catalogizer/internal/media/providers"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockProvider implements providers.MetadataProvider for testing.
type mockProvider struct {
	name      string
	enabled   bool
	searchFn  func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error)
	detailsFn func(ctx context.Context, externalID string) (*models.ExternalMetadata, error)
}

func (m *mockProvider) GetName() string { return m.name }
func (m *mockProvider) IsEnabled() bool { return m.enabled }

func (m *mockProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, mediaType, year)
	}
	return nil, nil
}

func (m *mockProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if m.detailsFn != nil {
		return m.detailsFn(ctx, externalID)
	}
	return nil, nil
}

// newTestProviderManager creates a ProviderManager with mock providers injected via RegisterProvider.
func newTestProviderManager(t *testing.T, mocks map[string]providers.MetadataProvider) *providers.ProviderManager {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	pm := providers.NewProviderManager(logger)
	for name, p := range mocks {
		pm.RegisterProvider(name, p)
	}
	return pm
}

// --- Tests ---

func TestProviderPipeline_SearchAll_MultipleProviders(t *testing.T) {
	year2020 := 2020
	rating8 := 8.1
	rating7 := 7.5

	providerA := &mockProvider{
		name:    "provider_a",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "a-1", Title: "The Matrix", Year: &year2020, Rating: &rating8, Relevance: 0.9},
				{ExternalID: "a-2", Title: "Matrix Reloaded", Relevance: 0.6},
			}, nil
		},
	}

	providerB := &mockProvider{
		name:    "provider_b",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "b-1", Title: "The Matrix", Year: &year2020, Rating: &rating7, Relevance: 0.85},
			}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"provider_a": providerA,
		"provider_b": providerB,
	})

	ctx := context.Background()
	results, err := pm.SearchAll(ctx, "The Matrix", "movie", &year2020, []string{"provider_a", "provider_b"})
	require.NoError(t, err)

	assert.Len(t, results, 2, "should have results from two providers")
	assert.Len(t, results["provider_a"], 2, "provider_a should return 2 results")
	assert.Len(t, results["provider_b"], 1, "provider_b should return 1 result")
	assert.Equal(t, "The Matrix", results["provider_a"][0].Title)
	assert.Equal(t, "The Matrix", results["provider_b"][0].Title)
}

func TestProviderPipeline_GetBestMatch_HighestScore(t *testing.T) {
	year1999 := 1999
	ratingHigh := 9.0
	ratingLow := 6.0

	// provider_a returns an exact title match with high rating and year match
	providerA := &mockProvider{
		name:    "tmdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "tmdb-1", Title: "Inception", Year: &year1999, Rating: &ratingHigh, Relevance: 0.9},
			}, nil
		},
	}

	// provider_b returns a partial match with lower rating
	providerB := &mockProvider{
		name:    "imdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "imdb-1", Title: "Inception 2", Rating: &ratingLow, Relevance: 0.5},
			}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb": providerA,
		"imdb": providerB,
	})

	ctx := context.Background()
	best, providerName, err := pm.GetBestMatch(ctx, "Inception", "movie", &year1999)
	require.NoError(t, err)
	require.NotNil(t, best, "should find a best match")

	// The TMDB result has: base relevance 0.9 + exact title match 0.3 + year match 0.2 + has rating 0.1 = 1.5
	// The IMDB result has: base relevance 0.5 + contains query 0.2 + has rating 0.1 = 0.8
	assert.Equal(t, "tmdb", providerName, "TMDB result should score highest")
	assert.Equal(t, "tmdb-1", best.ExternalID)
	assert.Equal(t, "Inception", best.Title)
}

func TestProviderPipeline_GetDetails_SpecificProvider(t *testing.T) {
	rating := 8.5
	coverURL := "https://example.com/poster.jpg"

	providerA := &mockProvider{
		name:    "test_provider",
		enabled: true,
		detailsFn: func(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
			return &models.ExternalMetadata{
				Provider:    "test_provider",
				ExternalID:  externalID,
				Data:        `{"title":"Test Movie"}`,
				Rating:      &rating,
				CoverURL:    &coverURL,
				LastFetched: time.Now(),
			}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"test_provider": providerA,
	})

	ctx := context.Background()
	meta, err := pm.GetDetails(ctx, "test_provider", "ext-123")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, "test_provider", meta.Provider)
	assert.Equal(t, "ext-123", meta.ExternalID)
	assert.InDelta(t, 8.5, *meta.Rating, 0.001)
	assert.Equal(t, "https://example.com/poster.jpg", *meta.CoverURL)
}

func TestProviderPipeline_GetDetails_ProviderNotFound(t *testing.T) {
	pm := newTestProviderManager(t, nil)

	ctx := context.Background()
	meta, err := pm.GetDetails(ctx, "nonexistent_provider", "id-1")
	assert.Error(t, err)
	assert.Nil(t, meta)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestProviderPipeline_GetDetails_ProviderDisabled(t *testing.T) {
	disabled := &mockProvider{
		name:    "disabled_provider",
		enabled: false,
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"disabled_provider": disabled,
	})

	ctx := context.Background()
	meta, err := pm.GetDetails(ctx, "disabled_provider", "id-1")
	assert.Error(t, err)
	assert.Nil(t, meta)
	assert.Contains(t, err.Error(), "provider disabled")
}

func TestProviderPipeline_GracefulDegradation_ProviderReturnsError(t *testing.T) {
	testYear := 2021
	testRating := 7.0

	// This provider fails with an error
	failingProvider := &mockProvider{
		name:    "tmdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return nil, fmt.Errorf("HTTP 500: Internal Server Error")
		},
	}

	// This provider succeeds
	workingProvider := &mockProvider{
		name:    "imdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "imdb-42", Title: "Good Movie", Year: &testYear, Rating: &testRating, Relevance: 0.8},
			}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb": failingProvider,
		"imdb": workingProvider,
	})

	ctx := context.Background()
	results, err := pm.SearchAll(ctx, "Good Movie", "movie", nil, []string{"tmdb", "imdb"})

	// SearchAll should not return an error -- it logs and continues
	require.NoError(t, err)
	assert.NotContains(t, results, "tmdb", "failing provider should not appear in results")
	assert.Contains(t, results, "imdb", "working provider should appear in results")
	assert.Len(t, results["imdb"], 1)
	assert.Equal(t, "Good Movie", results["imdb"][0].Title)
}

func TestProviderPipeline_DisabledProviders_Skipped(t *testing.T) {
	searchCalled := false

	disabled := &mockProvider{
		name:    "tmdb",
		enabled: false,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			searchCalled = true
			return []providers.SearchResult{{ExternalID: "should-not-appear", Title: "Ghost"}}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb": disabled,
	})

	ctx := context.Background()
	results, err := pm.SearchAll(ctx, "anything", "movie", nil, []string{"tmdb"})
	require.NoError(t, err)

	assert.False(t, searchCalled, "disabled provider's Search should not be called")
	assert.Empty(t, results, "no results should be returned from disabled providers")
}

func TestProviderPipeline_GetBestMatch_NoResults(t *testing.T) {
	emptyProvider := &mockProvider{
		name:    "tmdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return nil, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb": emptyProvider,
		"imdb": emptyProvider,
	})

	ctx := context.Background()
	best, providerName, err := pm.GetBestMatch(ctx, "Nonexistent Movie XYZ", "movie", nil)
	require.NoError(t, err)
	assert.Nil(t, best, "should return nil when no results found")
	assert.Empty(t, providerName, "provider name should be empty when no match")
}

func TestProviderPipeline_ContextCancellation(t *testing.T) {
	// Provider that respects context cancellation
	cancelProvider := &mockProvider{
		name:    "tmdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return []providers.SearchResult{
					{ExternalID: "1", Title: "Result", Relevance: 0.5},
				}, nil
			}
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb": cancelProvider,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	results, err := pm.SearchAll(ctx, "test", "movie", nil, []string{"tmdb"})
	// The provider returns ctx.Err(), which SearchAll logs and continues
	require.NoError(t, err)
	assert.Empty(t, results["tmdb"], "cancelled context should yield no results")
}

func TestProviderPipeline_SearchAll_EmptyResults_Excluded(t *testing.T) {
	// Provider that returns empty results
	emptyProvider := &mockProvider{
		name:    "provider_empty",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{}, nil
		},
	}

	// Provider that returns nil results
	nilProvider := &mockProvider{
		name:    "provider_nil",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return nil, nil
		},
	}

	// Provider that returns actual results
	realProvider := &mockProvider{
		name:    "provider_real",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "r-1", Title: "Real Result", Relevance: 0.7},
			}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"provider_empty": emptyProvider,
		"provider_nil":   nilProvider,
		"provider_real":  realProvider,
	})

	ctx := context.Background()
	results, err := pm.SearchAll(ctx, "test", "movie", nil, []string{"provider_empty", "provider_nil", "provider_real"})
	require.NoError(t, err)

	// Empty and nil results should be excluded from the map
	assert.NotContains(t, results, "provider_empty", "empty results should be excluded")
	assert.NotContains(t, results, "provider_nil", "nil results should be excluded")
	assert.Contains(t, results, "provider_real", "real results should be included")
}

func TestProviderPipeline_GetBestMatch_YearBonus(t *testing.T) {
	year2010 := 2010
	year2020 := 2020
	rating := 7.0

	// Both providers return same title, but only one matches the query year
	providerA := &mockProvider{
		name:    "tmdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "a-1", Title: "Inception", Year: &year2010, Rating: &rating, Relevance: 0.8},
			}, nil
		},
	}

	providerB := &mockProvider{
		name:    "imdb",
		enabled: true,
		searchFn: func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
			return []providers.SearchResult{
				{ExternalID: "b-1", Title: "Inception", Year: &year2020, Rating: &rating, Relevance: 0.8},
			}, nil
		},
	}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb": providerA,
		"imdb": providerB,
	})

	ctx := context.Background()
	best, providerName, err := pm.GetBestMatch(ctx, "Inception", "movie", &year2010)
	require.NoError(t, err)
	require.NotNil(t, best)

	// Provider A result: 0.8 + exact title 0.3 + year match 0.2 + rating 0.1 = 1.4
	// Provider B result: 0.8 + exact title 0.3 + no year match + rating 0.1 = 1.2
	assert.Equal(t, "tmdb", providerName, "result with matching year should score higher")
	assert.Equal(t, 2010, *best.Year)
}
