package services

import (
	"catalogizer/database"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewCoverArtService(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()

	service := NewCoverArtService(mockDB, mockLogger)

	assert.NotNil(t, service)
}

func TestCoverArtService_GetCoverArt_NilDB(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCoverArtService(mockDB, mockLogger)

	// GetCoverArt with nil DB panics because it directly uses s.db without nil check.
	// Verify it panics as expected rather than silently failing.
	assert.Panics(t, func() {
		_, _ = service.GetCoverArt(context.Background(), 1)
	})
}

func TestCoverArtService_GenerateCacheKey(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCoverArtService(mockDB, mockLogger)

	tests := []struct {
		name    string
		request *CoverArtSearchRequest
	}{
		{
			name: "basic request",
			request: &CoverArtSearchRequest{
				Artist:  "Test Artist",
				Title:   "Test Title",
				Quality: "high",
			},
		},
		{
			name: "empty fields",
			request: &CoverArtSearchRequest{
				Artist:  "",
				Title:   "",
				Quality: "",
			},
		},
		{
			name: "special characters",
			request: &CoverArtSearchRequest{
				Artist:  "Artist / With & Special",
				Title:   "Album (2024) [Deluxe]",
				Quality: "medium",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := service.generateCacheKey(tt.request)
			assert.NotEmpty(t, key)

			// Same inputs should produce same key
			key2 := service.generateCacheKey(tt.request)
			assert.Equal(t, key, key2)
		})
	}
}

func TestCoverArtService_GenerateCacheKey_Uniqueness(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCoverArtService(mockDB, mockLogger)

	key1 := service.generateCacheKey(&CoverArtSearchRequest{Artist: "Artist A", Title: "Title A", Quality: "high"})
	key2 := service.generateCacheKey(&CoverArtSearchRequest{Artist: "Artist B", Title: "Title B", Quality: "high"})

	assert.NotEqual(t, key1, key2)
}

func TestCoverArtService_GenerateTimestamps(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCoverArtService(mockDB, mockLogger)

	tests := []struct {
		name     string
		duration float64
		count    int
	}{
		{
			name:     "normal video",
			duration: 120.0,
			count:    5,
		},
		{
			name:     "short video",
			duration: 10.0,
			count:    3,
		},
		{
			name:     "zero duration",
			duration: 0.0,
			count:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamps := service.generateTimestamps(tt.duration, tt.count)
			assert.NotNil(t, timestamps)
		})
	}
}

func TestCoverArtService_SortCoverArtResults(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCoverArtService(mockDB, mockLogger)

	results := []CoverArtSearchResult{
		{MatchScore: 0.5, Width: 200},
		{MatchScore: 0.9, Width: 800},
		{MatchScore: 0.3, Width: 100},
	}

	service.sortCoverArtResults(results)

	// Results should be sorted by match score descending
	assert.GreaterOrEqual(t, results[0].MatchScore, results[1].MatchScore)
}

func TestCoverArtService_GenerateCoverArtID(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCoverArtService(mockDB, mockLogger)

	id1 := service.generateCoverArtID()
	id2 := service.generateCoverArtID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.Contains(t, id1, "cover_")
}
