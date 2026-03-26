package integration

import (
	"catalogizer/internal/media/models"
	"catalogizer/internal/media/providers"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockProvider implements providers.MetadataProvider for testing.
type mockProvider struct {
	name       string
	enabled    bool
	searchFn   func(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error)
	detailsFn  func(ctx context.Context, externalID string) (*models.ExternalMetadata, error)
}

func (m *mockProvider) GetName() string  { return m.name }
func (m *mockProvider) IsEnabled() bool  { return m.enabled }

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

// --- TMDB mock server helpers ---

// newTMDBSearchServer returns an httptest server that responds with TMDB-style search results.
func newTMDBSearchServer(t *testing.T, results []tmdbSearchItem) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := tmdbSearchResponse{Results: results}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(ts.Close)
	return ts
}

type tmdbSearchItem struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Name        string  `json:"name"`
	ReleaseDate string  `json:"release_date"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
}

type tmdbSearchResponse struct {
	Results []tmdbSearchItem `json:"results"`
}

// newTMDBDetailsServer returns an httptest server that responds with TMDB-style movie details.
func newTMDBDetailsServer(t *testing.T, title string, rating float64, posterPath string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"title":        title,
			"vote_average": rating,
			"poster_path":  posterPath,
			"homepage":     "https://example.com/movie",
			"videos":       map[string]interface{}{"results": []interface{}{}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(ts.Close)
	return ts
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

func TestProviderPipeline_GracefulDegradation_ServerReturns500(t *testing.T) {
	// Create a mock TMDB server that returns 500
	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	t.Cleanup(server500.Close)

	logger, _ := zap.NewDevelopment()
	client := &http.Client{Timeout: 5 * time.Second}

	// Build a TMDB provider pointing at the failing server
	bp := providers.NewBaseProvider("tmdb_test", server500.URL, "fake-api-key", client, logger)
	tmdbProvider := &tmdbTestProvider{BaseProvider: bp}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"tmdb_test": tmdbProvider,
	})

	ctx := context.Background()
	results, err := pm.SearchAll(ctx, "Test Movie", "movie", nil, []string{"tmdb_test"})
	require.NoError(t, err, "SearchAll should not fail when a provider returns 500")
	assert.Empty(t, results["tmdb_test"], "provider returning 500 should yield no results")
}

func TestProviderPipeline_TimeoutHandling(t *testing.T) {
	// Create a server that delays longer than the client timeout
	var requestCount int32
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		// Block until context is cancelled or 10 seconds
		select {
		case <-r.Context().Done():
			return
		case <-time.After(10 * time.Second):
			return
		}
	}))
	t.Cleanup(slowServer.Close)

	logger, _ := zap.NewDevelopment()
	client := &http.Client{Timeout: 500 * time.Millisecond} // Very short timeout

	bp := providers.NewBaseProvider("slow_provider", slowServer.URL, "fake-key", client, logger)
	slowProvider := &tmdbTestProvider{BaseProvider: bp}

	pm := newTestProviderManager(t, map[string]providers.MetadataProvider{
		"slow_provider": slowProvider,
	})

	ctx := context.Background()
	start := time.Now()
	results, err := pm.SearchAll(ctx, "Timeout Test", "movie", nil, []string{"slow_provider"})
	elapsed := time.Since(start)

	require.NoError(t, err, "SearchAll should not fail on timeout -- it logs and continues")
	assert.Empty(t, results["slow_provider"], "timed-out provider should yield no results")
	assert.Less(t, elapsed, 5*time.Second, "should timeout quickly, not block for 10s")
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount), "server should have received the request")
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

func TestProviderPipeline_TMDBProvider_MockServer_Search(t *testing.T) {
	// Create a mock TMDB API server
	ts := newTMDBSearchServer(t, []tmdbSearchItem{
		{ID: 550, Title: "Fight Club", ReleaseDate: "1999-10-15", Overview: "An insomniac office worker...", PosterPath: "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg", VoteAverage: 8.4},
		{ID: 551, Title: "Fight Club 2", ReleaseDate: "2025-01-01", Overview: "Sequel", VoteAverage: 0},
	})

	logger, _ := zap.NewDevelopment()
	client := &http.Client{Timeout: 5 * time.Second}

	// Create a TMDB provider pointing at our mock server
	bp := providers.NewBaseProvider("tmdb", ts.URL, "fake-api-key", client, logger)
	tmdb := &tmdbTestProvider{BaseProvider: bp}

	ctx := context.Background()
	results, err := tmdb.Search(ctx, "Fight Club", "movie", nil)
	require.NoError(t, err)
	require.Len(t, results, 2, "should parse both results from mock response")

	assert.Equal(t, "550", results[0].ExternalID)
	assert.Equal(t, "Fight Club", results[0].Title)
	assert.NotNil(t, results[0].Year)
	assert.Equal(t, 1999, *results[0].Year)
	assert.NotNil(t, results[0].Rating)
	assert.InDelta(t, 8.4, *results[0].Rating, 0.01)
	assert.NotNil(t, results[0].CoverURL)
	assert.Contains(t, *results[0].CoverURL, "image.tmdb.org")
}

func TestProviderPipeline_TMDBProvider_MockServer_GetDetails(t *testing.T) {
	ts := newTMDBDetailsServer(t, "Fight Club", 8.4, "/poster.jpg")

	logger, _ := zap.NewDevelopment()
	client := &http.Client{Timeout: 5 * time.Second}

	bp := providers.NewBaseProvider("tmdb", ts.URL, "fake-api-key", client, logger)
	tmdb := &tmdbTestProvider{BaseProvider: bp}

	ctx := context.Background()
	meta, err := tmdb.GetDetails(ctx, "550")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, "tmdb", meta.Provider)
	assert.Equal(t, "550", meta.ExternalID)
	assert.NotNil(t, meta.Rating)
	assert.InDelta(t, 8.4, *meta.Rating, 0.01)
	assert.NotNil(t, meta.CoverURL)
	assert.Contains(t, *meta.CoverURL, "image.tmdb.org")
	assert.NotNil(t, meta.ReviewURL)
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

// tmdbTestProvider wraps BaseProvider to provide TMDB-compatible Search and GetDetails
// that use the configurable baseURL (for httptest servers).
type tmdbTestProvider struct {
	*providers.BaseProvider
}

func (t *tmdbTestProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]providers.SearchResult, error) {
	if !t.IsEnabled() {
		return nil, nil
	}

	// Build the request URL using the base URL (which points to httptest server)
	params := url.Values{}
	params.Add("api_key", "fake")
	params.Add("query", query)
	requestURL := fmt.Sprintf("%s/search/movie?%s", t.GetBaseURL(), params.Encode())

	body, err := t.MakeRequest(ctx, requestURL, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []struct {
			ID          int     `json:"id"`
			Title       string  `json:"title"`
			Name        string  `json:"name"`
			ReleaseDate string  `json:"release_date"`
			Overview    string  `json:"overview"`
			PosterPath  string  `json:"poster_path"`
			VoteAverage float64 `json:"vote_average"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]providers.SearchResult, 0, len(response.Results))
	for _, item := range response.Results {
		title := item.Title
		if title == "" {
			title = item.Name
		}

		var yr *int
		if len(item.ReleaseDate) >= 4 {
			y := parseIntFromDate(item.ReleaseDate[:4])
			if y > 1900 {
				yr = &y
			}
		}

		var coverURL *string
		if item.PosterPath != "" {
			url := fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", item.PosterPath)
			coverURL = &url
		}

		var rating *float64
		if item.VoteAverage > 0 {
			rating = &item.VoteAverage
		}

		results = append(results, providers.SearchResult{
			ExternalID:  fmt.Sprintf("%d", item.ID),
			Title:       title,
			Year:        yr,
			Rating:      rating,
			Description: &item.Overview,
			CoverURL:    coverURL,
			Relevance:   0.8,
		})
	}

	return results, nil
}

func (t *tmdbTestProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !t.IsEnabled() {
		return nil, nil
	}

	requestURL := fmt.Sprintf("%s/movie/%s?api_key=fake", t.GetBaseURL(), externalID)

	body, err := t.MakeRequest(ctx, requestURL, nil)
	if err != nil {
		return nil, err
	}

	metadata := &models.ExternalMetadata{
		Provider:    t.GetName(),
		ExternalID:  externalID,
		Data:        string(body),
		LastFetched: time.Now(),
	}

	var details struct {
		Title       string  `json:"title"`
		VoteAverage float64 `json:"vote_average"`
		PosterPath  string  `json:"poster_path"`
		Homepage    string  `json:"homepage"`
	}

	if err := json.Unmarshal(body, &details); err == nil {
		if details.VoteAverage > 0 {
			metadata.Rating = &details.VoteAverage
		}
		if details.PosterPath != "" {
			url := fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", details.PosterPath)
			metadata.CoverURL = &url
		}
		if details.Homepage != "" {
			metadata.ReviewURL = &details.Homepage
		}
	}

	return metadata, nil
}

func parseIntFromDate(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}
