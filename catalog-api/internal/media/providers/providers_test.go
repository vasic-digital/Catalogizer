package providers

import (
	"catalogizer/internal/concurrency"
	mediamodels "catalogizer/internal/media/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger(t *testing.T) *zap.Logger {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	return logger
}

// --- BaseProvider tests ---

func TestBaseProvider_GetName(t *testing.T) {
	bp := NewBaseProvider("test_provider", "https://example.com", "key123", http.DefaultClient, testLogger(t))
	assert.Equal(t, "test_provider", bp.GetName())
}

func TestBaseProvider_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		enabled bool
	}{
		{"enabled with API key", "some-key", true},
		{"disabled without API key", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bp := NewBaseProvider("test", "https://example.com", tc.apiKey, http.DefaultClient, testLogger(t))
			assert.Equal(t, tc.enabled, bp.IsEnabled())
		})
	}
}

func TestBaseProvider_MakeRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": "ok"}`))
		}))
		defer server.Close()

		bp := NewBaseProvider("test", server.URL, "test-key", server.Client(), testLogger(t))
		body, err := bp.makeRequest(context.Background(), server.URL+"/test", nil)
		require.NoError(t, err)
		assert.Contains(t, string(body), "ok")
	})

	t.Run("non-200 response returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		bp := NewBaseProvider("test", server.URL, "key", server.Client(), testLogger(t))
		_, err := bp.makeRequest(context.Background(), server.URL+"/missing", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("custom headers are sent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
		}))
		defer server.Close()

		bp := NewBaseProvider("test", server.URL, "key", server.Client(), testLogger(t))
		_, err := bp.makeRequest(context.Background(), server.URL, map[string]string{"Accept": "application/json"})
		require.NoError(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		bp := NewBaseProvider("test", server.URL, "key", server.Client(), testLogger(t))
		_, err := bp.makeRequest(ctx, server.URL, nil)
		require.Error(t, err)
	})
}

// --- ProviderManager tests ---

func TestNewProviderManager(t *testing.T) {
	pm := NewProviderManager(testLogger(t))
	require.NotNil(t, pm)
	assert.NotEmpty(t, pm.lazyProviders)
	// Verify key providers are registered (lazily)
	assert.Contains(t, pm.lazyProviders, "tmdb")
	assert.Contains(t, pm.lazyProviders, "imdb")
	assert.Contains(t, pm.lazyProviders, "musicbrainz")
	assert.Contains(t, pm.lazyProviders, "igdb")
	assert.Contains(t, pm.lazyProviders, "github")

	// Verify lazy resolution works
	tmdb, ok := pm.resolveProvider("tmdb")
	assert.True(t, ok)
	assert.NotNil(t, tmdb)
	assert.Equal(t, "tmdb", tmdb.GetName())
}

func TestGetProvidersForMediaType(t *testing.T) {
	pm := NewProviderManager(testLogger(t))

	tests := []struct {
		name              string
		mediaType         string
		expectedProviders []string
	}{
		{"movie providers", "movie", []string{"tmdb", "imdb"}},
		{"tv_show providers", "tv_show", []string{"tmdb", "imdb", "tvdb"}},
		{"anime providers", "anime", []string{"anidb", "myanimelist", "tmdb"}},
		{"music providers", "music", []string{"musicbrainz", "spotify", "lastfm"}},
		{"pc_game providers", "pc_game", []string{"igdb", "steam"}},
		{"ebook providers", "ebook", []string{"goodreads", "openlibrary"}},
		{"unknown type falls back to defaults", "unknown_type", []string{"tmdb", "imdb"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			providers := pm.getProvidersForMediaType(tc.mediaType)
			assert.Equal(t, tc.expectedProviders, providers)
		})
	}
}

func TestGetDetails_ProviderNotFound(t *testing.T) {
	pm := NewProviderManager(testLogger(t))

	_, err := pm.GetDetails(context.Background(), "nonexistent_provider", "123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestGetDetails_ProviderDisabled(t *testing.T) {
	pm := NewProviderManager(testLogger(t))

	// TMDB is disabled by default (no API key)
	_, err := pm.GetDetails(context.Background(), "tmdb", "123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider disabled")
}

func TestCalculateRelevanceScore(t *testing.T) {
	pm := NewProviderManager(testLogger(t))
	year2010 := 2010
	rating8 := 8.0

	tests := []struct {
		name      string
		result    SearchResult
		query     string
		year      *int
		minScore  float64
	}{
		{
			name: "exact title match with year and rating",
			result: SearchResult{
				Title:     "Inception",
				Year:      &year2010,
				Rating:    &rating8,
				Relevance: 0.8,
			},
			query:    "Inception",
			year:     &year2010,
			minScore: 1.3, // 0.8 base + 0.3 exact + 0.2 year + 0.1 rating
		},
		{
			name: "partial title match",
			result: SearchResult{
				Title:     "Inception: The Movie",
				Relevance: 0.5,
			},
			query:    "inception",
			year:     nil,
			minScore: 0.7, // 0.5 base + 0.2 partial
		},
		{
			name: "no match at all",
			result: SearchResult{
				Title:     "Completely Different",
				Relevance: 0.5,
			},
			query:    "Inception",
			year:     nil,
			minScore: 0.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := pm.calculateRelevanceScore(tc.result, tc.query, tc.year)
			assert.GreaterOrEqual(t, score, tc.minScore)
		})
	}
}

func TestSearchAll_AllDisabledProviders(t *testing.T) {
	pm := NewProviderManager(testLogger(t))

	// All providers are disabled (no API keys), so SearchAll should return empty results
	results, err := pm.SearchAll(context.Background(), "test query", "movie", nil, nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetBestMatch_NoResults(t *testing.T) {
	pm := NewProviderManager(testLogger(t))

	// All providers disabled, so no results
	result, provider, err := pm.GetBestMatch(context.Background(), "nonexistent movie", "movie", nil)
	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Empty(t, provider)
}

// --- TMDBProvider tests with mock server ---

func TestTMDBProvider_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/search/movie")
		assert.Equal(t, "test-api-key", r.URL.Query().Get("api_key"))
		assert.Equal(t, "Inception", r.URL.Query().Get("query"))

		resp := map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":           12345,
					"title":        "Inception",
					"release_date": "2010-07-16",
					"overview":     "A mind-bending thriller",
					"poster_path":  "/poster.jpg",
					"vote_average": 8.8,
				},
				{
					"id":           99999,
					"title":        "Inception: The Game",
					"release_date": "2011-01-01",
					"overview":     "Not the movie",
					"poster_path":  "",
					"vote_average": 0,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "test-api-key", server.Client(), testLogger(t)),
	}

	results, err := provider.Search(context.Background(), "Inception", "movie", nil)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First result
	assert.Equal(t, "12345", results[0].ExternalID)
	assert.Equal(t, "Inception", results[0].Title)
	require.NotNil(t, results[0].Year)
	assert.Equal(t, 2010, *results[0].Year)
	require.NotNil(t, results[0].Rating)
	assert.Equal(t, 8.8, *results[0].Rating)
	require.NotNil(t, results[0].CoverURL)
	assert.Contains(t, *results[0].CoverURL, "poster.jpg")

	// Second result: no rating, no cover
	assert.Nil(t, results[1].Rating)
	assert.Nil(t, results[1].CoverURL)
}

func TestTMDBProvider_Search_TVShow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/search/tv")
		resp := map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":             67890,
					"name":           "Breaking Bad",
					"first_air_date": "2008-01-20",
					"overview":       "A teacher turns to crime",
					"poster_path":    "/bb.jpg",
					"vote_average":   9.5,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "key", server.Client(), testLogger(t)),
	}

	results, err := provider.Search(context.Background(), "Breaking Bad", "tv_show", nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Breaking Bad", results[0].Title)
	require.NotNil(t, results[0].Year)
	assert.Equal(t, 2008, *results[0].Year)
}

func TestTMDBProvider_Search_WithYear(t *testing.T) {
	year := 2010
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2010", r.URL.Query().Get("year"))
		resp := map[string]interface{}{"results": []map[string]interface{}{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "key", server.Client(), testLogger(t)),
	}

	results, err := provider.Search(context.Background(), "Inception", "movie", &year)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestTMDBProvider_Search_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "key", server.Client(), testLogger(t)),
	}

	_, err := provider.Search(context.Background(), "test", "movie", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestTMDBProvider_Search_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "key", server.Client(), testLogger(t)),
	}

	_, err := provider.Search(context.Background(), "test", "movie", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestTMDBProvider_GetDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/movie/12345")
		resp := map[string]interface{}{
			"title":        "Inception",
			"vote_average": 8.8,
			"poster_path":  "/poster.jpg",
			"homepage":     "https://inception-movie.com",
			"videos": map[string]interface{}{
				"results": []map[string]interface{}{
					{"site": "YouTube", "key": "abc123", "type": "Trailer"},
					{"site": "Vimeo", "key": "def456", "type": "Teaser"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "key", server.Client(), testLogger(t)),
	}

	metadata, err := provider.GetDetails(context.Background(), "12345")
	require.NoError(t, err)
	require.NotNil(t, metadata)

	assert.Equal(t, "tmdb", metadata.Provider)
	assert.Equal(t, "12345", metadata.ExternalID)
	require.NotNil(t, metadata.Rating)
	assert.Equal(t, 8.8, *metadata.Rating)
	require.NotNil(t, metadata.CoverURL)
	assert.Contains(t, *metadata.CoverURL, "poster.jpg")
	require.NotNil(t, metadata.ReviewURL)
	assert.Equal(t, "https://inception-movie.com", *metadata.ReviewURL)
	require.NotNil(t, metadata.TrailerURL)
	assert.Contains(t, *metadata.TrailerURL, "abc123")
}

func TestTMDBProvider_GetDetails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", server.URL, "bad-key", server.Client(), testLogger(t)),
	}

	_, err := provider.GetDetails(context.Background(), "999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// --- providers parseInt tests ---

func TestProviders_ParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"2010", 2010},
		{"0", 0},
		{"", 0},
		{"abc", 0},
		{"2010-07", 201007},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseInt(tc.input))
		})
	}
}

// --- SearchAll with mock provider ---

type mockProvider struct {
	name    string
	enabled bool
	results []SearchResult
	err     error
}

func (m *mockProvider) GetName() string { return m.name }
func (m *mockProvider) IsEnabled() bool { return m.enabled }
func (m *mockProvider) Search(_ context.Context, _ string, _ string, _ *int) ([]SearchResult, error) {
	return m.results, m.err
}
func (m *mockProvider) GetDetails(_ context.Context, _ string) (*mediamodels.ExternalMetadata, error) {
	return nil, nil
}

func TestSearchAll_WithMockProviders(t *testing.T) {
	logger := testLogger(t)
	pm := &ProviderManager{
		providers:     make(map[string]MetadataProvider),
		lazyProviders: make(map[string]*LazyProvider),
		logger:        logger,
		client:        http.DefaultClient,
		searchSem:     concurrency.NewBoundedSemaphore("test", DefaultMaxConcurrentSearches, 5*time.Second),
	}

	year2010 := 2010
	rating8 := 8.0

	pm.providers["mock1"] = &mockProvider{
		name:    "mock1",
		enabled: true,
		results: []SearchResult{
			{ExternalID: "1", Title: "Result 1", Year: &year2010, Rating: &rating8, Relevance: 0.9},
		},
	}
	pm.providers["mock2"] = &mockProvider{
		name:    "mock2",
		enabled: false, // disabled
		results: []SearchResult{
			{ExternalID: "2", Title: "Result 2", Relevance: 0.5},
		},
	}
	pm.providers["mock3"] = &mockProvider{
		name:    "mock3",
		enabled: true,
		results: []SearchResult{}, // empty
	}

	results, err := pm.SearchAll(context.Background(), "test", "movie", nil, []string{"mock1", "mock2", "mock3"})
	require.NoError(t, err)

	// mock1 has results, mock2 is disabled, mock3 returns empty
	assert.Len(t, results, 1)
	assert.Contains(t, results, "mock1")
	assert.NotContains(t, results, "mock2")
	assert.NotContains(t, results, "mock3")
}

func TestGetBestMatch_WithMockProviders(t *testing.T) {
	logger := testLogger(t)
	pm := &ProviderManager{
		providers:     make(map[string]MetadataProvider),
		lazyProviders: make(map[string]*LazyProvider),
		logger:        logger,
		client:        http.DefaultClient,
		searchSem:     concurrency.NewBoundedSemaphore("test", DefaultMaxConcurrentSearches, 5*time.Second),
	}

	year := 2010
	rating := 8.0

	pm.providers["tmdb"] = &mockProvider{
		name:    "tmdb",
		enabled: true,
		results: []SearchResult{
			{ExternalID: "1", Title: "Inception", Year: &year, Rating: &rating, Relevance: 0.8},
		},
	}
	pm.providers["imdb"] = &mockProvider{
		name:    "imdb",
		enabled: true,
		results: []SearchResult{
			{ExternalID: "tt999", Title: "Something Else", Relevance: 0.3},
		},
	}

	result, provider, err := pm.GetBestMatch(context.Background(), "Inception", "movie", &year)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "tmdb", provider)
	assert.Equal(t, "Inception", result.Title)
}

// --- OpenLibrary Provider tests with mock server ---

func TestOpenLibraryProvider_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/search.json")
		assert.Equal(t, "Lord of the Rings", r.URL.Query().Get("q"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		resp := map[string]interface{}{
			"docs": []map[string]interface{}{
				{
					"key":                "/works/OL27448W",
					"title":              "The Lord of the Rings",
					"first_publish_year": 1954,
					"author_name":        []string{"J.R.R. Tolkien"},
					"cover_i":            123456,
				},
				{
					"key":                "/works/OL99999W",
					"title":              "The Lord of the Rings: The Two Towers",
					"first_publish_year": 1954,
					"author_name":        []string{"J.R.R. Tolkien"},
				},
				{
					"key":   "/works/OL88888W",
					"title": "Lord of the Rings Study Guide",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	results, err := provider.Search(context.Background(), "Lord of the Rings", "ebook", nil)
	require.NoError(t, err)
	require.Len(t, results, 3)

	// First result: has author and cover
	assert.Equal(t, "/works/OL27448W", results[0].ExternalID)
	assert.Equal(t, "The Lord of the Rings", results[0].Title)
	require.NotNil(t, results[0].Year)
	assert.Equal(t, 1954, *results[0].Year)
	require.NotNil(t, results[0].Description)
	assert.Contains(t, *results[0].Description, "J.R.R. Tolkien")
	require.NotNil(t, results[0].CoverURL)
	assert.Contains(t, *results[0].CoverURL, "123456")
	assert.Contains(t, *results[0].CoverURL, "-M.jpg")
	assert.Equal(t, 0.7, results[0].Relevance)

	// Second result: has author, no cover
	assert.Equal(t, "/works/OL99999W", results[1].ExternalID)
	require.NotNil(t, results[1].Description)
	assert.Nil(t, results[1].CoverURL)

	// Third result: no author, no cover
	assert.Equal(t, "/works/OL88888W", results[2].ExternalID)
	assert.Nil(t, results[2].Description)
	assert.Nil(t, results[2].Year)
}

func TestOpenLibraryProvider_Search_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	_, err := provider.Search(context.Background(), "test", "ebook", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestOpenLibraryProvider_Search_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	_, err := provider.Search(context.Background(), "test", "ebook", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestOpenLibraryProvider_GetDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/works/OL27448W.json")

		resp := map[string]interface{}{
			"title":       "The Lord of the Rings",
			"description": "An epic high-fantasy novel by J.R.R. Tolkien.",
			"covers":      []int{123456, 789012},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	metadata, err := provider.GetDetails(context.Background(), "/works/OL27448W")
	require.NoError(t, err)
	require.NotNil(t, metadata)

	assert.Equal(t, "openlibrary", metadata.Provider)
	assert.Equal(t, "/works/OL27448W", metadata.ExternalID)
	assert.NotEmpty(t, metadata.Data)
	require.NotNil(t, metadata.CoverURL)
	assert.Contains(t, *metadata.CoverURL, "123456")
	assert.Contains(t, *metadata.CoverURL, "-L.jpg")
}

func TestOpenLibraryProvider_GetDetails_DescriptionAsObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"title": "Some Book",
			"description": map[string]interface{}{
				"type":  "/type/text",
				"value": "A description in object form.",
			},
			"covers": []int{555},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	metadata, err := provider.GetDetails(context.Background(), "/works/OL11111W")
	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.NotNil(t, metadata.CoverURL)
	assert.Contains(t, *metadata.CoverURL, "555")
}

func TestOpenLibraryProvider_GetDetails_NoCover(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"title":  "No Cover Book",
			"covers": []int{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	metadata, err := provider.GetDetails(context.Background(), "/works/OL22222W")
	require.NoError(t, err)
	require.NotNil(t, metadata)
	assert.Nil(t, metadata.CoverURL)
}

func TestOpenLibraryProvider_GetDetails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := &OpenLibraryProvider{
		BaseProvider: NewBaseProvider("openlibrary", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	_, err := provider.GetDetails(context.Background(), "/works/NONEXISTENT")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// --- MusicBrainz Provider tests with mock server ---

func TestMusicBrainzProvider_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/recording")
		assert.Equal(t, "Bohemian Rhapsody", r.URL.Query().Get("query"))
		assert.Equal(t, "json", r.URL.Query().Get("fmt"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		// Verify User-Agent header is set
		assert.Contains(t, r.Header.Get("User-Agent"), "Catalogizer")

		resp := map[string]interface{}{
			"recordings": []map[string]interface{}{
				{
					"id":    "uuid-1234-5678",
					"title": "Bohemian Rhapsody",
					"artist-credit": []map[string]interface{}{
						{"name": "Queen"},
					},
					"releases": []map[string]interface{}{
						{"title": "A Night at the Opera", "date": "1975-10-31"},
					},
				},
				{
					"id":    "uuid-9999-0000",
					"title": "Bohemian Rhapsody (Live)",
					"artist-credit": []map[string]interface{}{
						{"name": "Queen"},
						{"name": "Adam Lambert"},
					},
					"releases": []map[string]interface{}{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &MusicBrainzProvider{
		BaseProvider: NewBaseProvider("musicbrainz", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	results, err := provider.Search(context.Background(), "Bohemian Rhapsody", "music", nil)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First result: has artist, release with date
	assert.Equal(t, "uuid-1234-5678", results[0].ExternalID)
	assert.Equal(t, "Bohemian Rhapsody", results[0].Title)
	require.NotNil(t, results[0].Year)
	assert.Equal(t, 1975, *results[0].Year)
	require.NotNil(t, results[0].Description)
	assert.Contains(t, *results[0].Description, "Queen")
	assert.Contains(t, *results[0].Description, "A Night at the Opera")
	assert.Equal(t, 0.7, results[0].Relevance)

	// Second result: multiple artists, no releases → no year
	assert.Equal(t, "uuid-9999-0000", results[1].ExternalID)
	assert.Nil(t, results[1].Year)
	require.NotNil(t, results[1].Description)
	assert.Contains(t, *results[1].Description, "Queen")
	assert.Contains(t, *results[1].Description, "Adam Lambert")
}

func TestMusicBrainzProvider_Search_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"recordings": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &MusicBrainzProvider{
		BaseProvider: NewBaseProvider("musicbrainz", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	results, err := provider.Search(context.Background(), "nonexistent track xyz", "music", nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestMusicBrainzProvider_Search_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := &MusicBrainzProvider{
		BaseProvider: NewBaseProvider("musicbrainz", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	_, err := provider.Search(context.Background(), "test", "music", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestMusicBrainzProvider_Search_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	provider := &MusicBrainzProvider{
		BaseProvider: NewBaseProvider("musicbrainz", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	_, err := provider.Search(context.Background(), "test", "music", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestMusicBrainzProvider_GetDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/recording/uuid-1234-5678")
		assert.Equal(t, "artists+releases", r.URL.Query().Get("inc"))
		assert.Equal(t, "json", r.URL.Query().Get("fmt"))
		assert.Contains(t, r.Header.Get("User-Agent"), "Catalogizer")

		resp := map[string]interface{}{
			"title": "Bohemian Rhapsody",
			"artist-credit": []map[string]interface{}{
				{"name": "Queen"},
			},
			"releases": []map[string]interface{}{
				{"title": "A Night at the Opera", "date": "1975-10-31"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &MusicBrainzProvider{
		BaseProvider: NewBaseProvider("musicbrainz", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	metadata, err := provider.GetDetails(context.Background(), "uuid-1234-5678")
	require.NoError(t, err)
	require.NotNil(t, metadata)

	assert.Equal(t, "musicbrainz", metadata.Provider)
	assert.Equal(t, "uuid-1234-5678", metadata.ExternalID)
	assert.NotEmpty(t, metadata.Data)
	require.NotNil(t, metadata.ReviewURL)
	assert.Contains(t, *metadata.ReviewURL, "musicbrainz.org/recording/uuid-1234-5678")
	// MusicBrainz doesn't provide cover art or ratings
	assert.Nil(t, metadata.CoverURL)
	assert.Nil(t, metadata.Rating)
}

func TestMusicBrainzProvider_GetDetails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := &MusicBrainzProvider{
		BaseProvider: NewBaseProvider("musicbrainz", server.URL, "", server.Client(), testLogger(t)),
	}
	provider.enabled = true

	_, err := provider.GetDetails(context.Background(), "nonexistent-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// --- Stub provider graceful degradation tests ---

func TestStubProvider_SearchReturnsNilWhenDisabled(t *testing.T) {
	logger := testLogger(t)
	client := http.DefaultClient

	// All these providers have no API key → disabled
	stubProviders := []MetadataProvider{
		NewIMDBProvider(client, logger),
		NewTVDBProvider(client, logger),
		NewSpotifyProvider(client, logger),
		NewLastFMProvider(client, logger),
		NewIGDBProvider(client, logger),
		NewSteamProvider(client, logger),
		NewGoodreadsProvider(client, logger),
		NewAniDBProvider(client, logger),
		NewMyAnimeListProvider(client, logger),
		NewYouTubeProvider(client, logger),
		NewGitHubProvider(client, logger),
	}

	for _, provider := range stubProviders {
		t.Run(provider.GetName()+"_search", func(t *testing.T) {
			assert.False(t, provider.IsEnabled())
			results, err := provider.Search(context.Background(), "test query", "movie", nil)
			require.NoError(t, err)
			assert.Nil(t, results)
		})
	}
}

func TestStubProvider_GetDetailsReturnsNilWhenDisabled(t *testing.T) {
	logger := testLogger(t)
	client := http.DefaultClient

	stubProviders := []MetadataProvider{
		NewIMDBProvider(client, logger),
		NewTVDBProvider(client, logger),
		NewSpotifyProvider(client, logger),
		NewLastFMProvider(client, logger),
		NewIGDBProvider(client, logger),
		NewSteamProvider(client, logger),
		NewGoodreadsProvider(client, logger),
		NewAniDBProvider(client, logger),
		NewMyAnimeListProvider(client, logger),
		NewYouTubeProvider(client, logger),
		NewGitHubProvider(client, logger),
	}

	for _, provider := range stubProviders {
		t.Run(provider.GetName()+"_get_details", func(t *testing.T) {
			assert.False(t, provider.IsEnabled())
			metadata, err := provider.GetDetails(context.Background(), "some-id")
			require.NoError(t, err)
			assert.Nil(t, metadata)
		})
	}
}

// --- Free provider enabled-by-default tests ---

func TestFreeProviders_EnabledByDefault(t *testing.T) {
	logger := testLogger(t)
	client := http.DefaultClient

	// OpenLibrary and MusicBrainz should be enabled even without API keys
	olProvider := NewOpenLibraryProvider(client, logger)
	assert.True(t, olProvider.IsEnabled(), "OpenLibrary should be enabled by default (free API)")

	mbProvider := NewMusicBrainzProvider(client, logger)
	assert.True(t, mbProvider.IsEnabled(), "MusicBrainz should be enabled by default (free API)")
}

func TestProviderManager_FreeProvidersRegistered(t *testing.T) {
	pm := NewProviderManager(testLogger(t))

	// Verify free providers are enabled in the manager (resolved lazily)
	olProvider, exists := pm.resolveProvider("openlibrary")
	require.True(t, exists)
	assert.True(t, olProvider.IsEnabled())

	mbProvider, exists := pm.resolveProvider("musicbrainz")
	require.True(t, exists)
	assert.True(t, mbProvider.IsEnabled())
}
