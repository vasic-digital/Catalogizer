package providers

import (
	"net/http"
	"testing"

	"go.uber.org/zap"
)

func newBenchProviderManager() *ProviderManager {
	logger, _ := zap.NewProduction()
	return &ProviderManager{
		providers: make(map[string]MetadataProvider),
		logger:    logger,
		client:    http.DefaultClient,
	}
}

func BenchmarkCalculateRelevanceScore(b *testing.B) {
	pm := newBenchProviderManager()
	year2010 := 2010
	rating8 := 8.0

	benchmarks := []struct {
		name   string
		result SearchResult
		query  string
		year   *int
	}{
		{
			"exact_title_with_year_and_rating",
			SearchResult{
				Title:     "Inception",
				Year:      &year2010,
				Rating:    &rating8,
				Relevance: 0.8,
			},
			"Inception",
			&year2010,
		},
		{
			"partial_title_match",
			SearchResult{
				Title:     "Inception: The IMAX Experience",
				Relevance: 0.6,
			},
			"inception",
			nil,
		},
		{
			"no_match",
			SearchResult{
				Title:     "Completely Different Title",
				Relevance: 0.5,
			},
			"Inception",
			nil,
		},
		{
			"year_match_no_title_match",
			SearchResult{
				Title:     "Another Movie",
				Year:      &year2010,
				Relevance: 0.4,
			},
			"Inception",
			&year2010,
		},
		{
			"long_title_comparison",
			SearchResult{
				Title:     "The Lord of the Rings: The Fellowship of the Ring Extended Edition",
				Year:      &year2010,
				Rating:    &rating8,
				Relevance: 0.7,
			},
			"the lord of the rings the fellowship of the ring",
			&year2010,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = pm.calculateRelevanceScore(bm.result, bm.query, bm.year)
			}
		})
	}
}

func BenchmarkGetProvidersForMediaType(b *testing.B) {
	pm := newBenchProviderManager()

	mediaTypes := []string{"movie", "tv_show", "anime", "music", "pc_game", "ebook", "unknown_type"}

	for _, mt := range mediaTypes {
		b.Run(mt, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = pm.getProvidersForMediaType(mt)
			}
		})
	}
}
