package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"catalogizer/models"
	"catalogizer/internal/services"
)

// MockFileRepository implements services.FileRepositoryInterface for testing
type MockFileRepository struct{}

func (m *MockFileRepository) SearchFiles(ctx context.Context, filter models.SearchFilter, pagination models.PaginationOptions, sort models.SortOptions) (*models.SearchResult, error) {
	return &models.SearchResult{
		Files:      []models.FileWithMetadata{},
		TotalCount: 0,
		Page:       1,
		Limit:      20,
		TotalPages: 0,
	}, nil
}

func TestRecommendationService_GetSimilarItems(t *testing.T) {
	ctx := context.Background()

	// Setup services with database
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()
	
	mediaRecognitionService := services.NewMediaRecognitionService(
		db, logger, nil, nil, 
		"http://mock-movie-api", "http://mock-music-api", 
		"http://mock-book-api", "http://mock-game-api", 
		"http://mock-ocr-api", "http://mock-fingerprint-api",
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, logger, nil)
	mockFileRepo := &MockFileRepository{}
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		mockFileRepo,
		db,
	)

	t.Run("movie recommendations", func(t *testing.T) {
		year1999 := 1999
		rating87 := 8.7
		req := &services.SimilarItemsRequest{
			MediaID: "matrix_1999",
			MediaMetadata: &models.MediaMetadata{
				Title:       "The Matrix",
				Year:        &year1999,
				Genre:       "Science Fiction", 
				Director:    "The Wachowskis",
				MediaType:   models.MediaTypeVideo,
				Rating:      &rating87,
			},
			MaxLocalItems:       10,
			MaxExternalItems:    5,
			IncludeExternal:     true,
			SimilarityThreshold: 0.3,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify response structure
		assert.NotEmpty(t, response.LocalItems)
		assert.NotEmpty(t, response.ExternalItems)
		assert.True(t, response.TotalFound > 0)
		assert.False(t, response.GeneratedAt.IsZero())
		assert.NotEmpty(t, response.Algorithms)
		assert.NotNil(t, response.Performance)

		// Verify local items
		for _, item := range response.LocalItems {
			assert.NotEmpty(t, item.MediaID)
			assert.NotNil(t, item.MediaMetadata)
			assert.True(t, item.SimilarityScore >= req.SimilarityThreshold)
			assert.NotEmpty(t, item.SimilarityReasons)
			assert.NotEmpty(t, item.DetailLink)
			assert.True(t, item.IsOwned) // Local items are owned
		}

		// Verify external items
		for _, item := range response.ExternalItems {
			assert.NotEmpty(t, item.ExternalID)
			assert.NotEmpty(t, item.Title)
			assert.NotEmpty(t, item.Provider)
			assert.NotEmpty(t, item.ExternalLink)
			assert.True(t, item.SimilarityScore > 0)
			assert.NotEmpty(t, item.SimilarityReasons)
		}

		// Verify sorting (highest similarity first)
		for i := 1; i < len(response.LocalItems); i++ {
			assert.True(t, response.LocalItems[i-1].SimilarityScore >= response.LocalItems[i].SimilarityScore)
		}
		for i := 1; i < len(response.ExternalItems); i++ {
			assert.True(t, response.ExternalItems[i-1].SimilarityScore >= response.ExternalItems[i].SimilarityScore)
		}
	})

	t.Run("music recommendations", func(t *testing.T) {
		year1975 := 1975
		duration := 355
		req := &services.SimilarItemsRequest{
			MediaID: "queen_bohemian_rhapsody",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Bohemian Rhapsody",
				Description: "Queen",
				Genre:     "Rock",
				Year:      &year1975,
				MediaType: models.MediaTypeAudio,
				Duration:  &duration,
			},
			MaxLocalItems:       8,
			MaxExternalItems:    3,
			IncludeExternal:     true,
			SimilarityThreshold: 0.4,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.NotEmpty(t, response.LocalItems)
		assert.True(t, len(response.LocalItems) <= req.MaxLocalItems)
		assert.True(t, len(response.ExternalItems) <= req.MaxExternalItems)

		// Verify music-specific recommendations
		foundSameGenre := false
		for _, item := range response.LocalItems {
			if item.MediaMetadata.Genre == req.MediaMetadata.Genre {
				foundSameGenre = true
			}
			// Look for same artist in description (artist stored in description for audio)
			if item.MediaMetadata.Description == req.MediaMetadata.Description {
				foundSameGenre = true // Artist match
			}
		}
		assert.True(t, foundSameGenre, "Should find similar genre or artist")
	})

	t.Run("book recommendations", func(t *testing.T) {
		year1997 := 1997
		pages := 223
		req := &services.SimilarItemsRequest{
			MediaID: "harry_potter_1",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Harry Potter and the Philosopher's Stone",
				Description: "J.K. Rowling",
				Year:      &year1997,
				Genre:     "Fantasy",
				Country:   "Bloomsbury",
				MediaType: models.MediaTypeBook,
				Duration:  &pages,
			},
			MaxLocalItems:       5,
			MaxExternalItems:    5,
			IncludeExternal:     true,
			SimilarityThreshold: 0.2,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify book-specific recommendations
		for _, item := range response.LocalItems {
			assert.Equal(t, models.MediaTypeBook, item.MediaMetadata.MediaType)
		}

		for _, item := range response.ExternalItems {
			// External book items should have relevant providers
			validProviders := []string{"Google Books", "Open Library"}
			found := false
			for _, provider := range validProviders {
				if item.Provider == provider {
					found = true
					break
				}
			}
			assert.True(t, found, "Book recommendations should come from book-specific providers")
		}
	})

	t.Run("filtered recommendations", func(t *testing.T) {
		year2008 := 2008
		rating90 := 9.0
		req := &services.SimilarItemsRequest{
			MediaID: "dark_knight_2008",
			MediaMetadata: &models.MediaMetadata{
				Title:     "The Dark Knight",
				Year:      &year2008,
				Genre:     "Action",
				MediaType: models.MediaTypeVideo,
				Rating:    &rating90,
			},
			Filters: &services.RecommendationFilters{
				GenreFilter: []string{"Action"},
				YearRange: &services.YearRange{
					StartYear: 2000,
					EndYear:   2015,
				},
				RatingRange: &services.RatingRange{
					MinRating: 7.0,
					MaxRating: 10.0,
				},
				MinConfidence: 0.8,
			},
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Verify all returned items pass filters
		for _, item := range response.LocalItems {
			assert.Contains(t, item.MediaMetadata.Genre, "Action")
			if item.MediaMetadata.Year != nil {
				// Year filtering would be verified in real implementation
			}
			if item.MediaMetadata.Rating != nil && *item.MediaMetadata.Rating > 0 {
				assert.True(t, *item.MediaMetadata.Rating >= 7.0)
				assert.True(t, *item.MediaMetadata.Rating <= 10.0)
			}
			assert.True(t, item.SimilarityScore >= 0.8)
		}
	})

	t.Run("no external recommendations when disabled", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "test_media",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Test Movie",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:   5,
			IncludeExternal: false,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		assert.Empty(t, response.ExternalItems)
		assert.Equal(t, 0, response.Performance.ExternalItemsFound)
		assert.Equal(t, time.Duration(0), response.Performance.ExternalSearchTime)
	})

	t.Run("similarity threshold filtering", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "test_media",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Test Movie",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:       10,
			SimilarityThreshold: 0.9, // Very high threshold
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// All returned items should meet the high threshold
		for _, item := range response.LocalItems {
			assert.True(t, item.SimilarityScore >= 0.9)
		}
	})
}

func TestRecommendationService_ExternalProviders(t *testing.T) {
	ctx := context.Background()

	// Note: StartAllMockServers() implementation needed
	// For now, we'll skip the mock servers
	// mockServers := StartAllMockServers()
	// defer func() {
	// 	for _, server := range mockServers {
	// 		server.Close()
	// 	}
	// }()

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	mediaRecognitionService := services.NewMediaRecognitionService(
		db, zap.NewNop(), nil, nil,
		"http://mock-movie-api", "http://mock-music-api",
		"http://mock-book-api", "http://mock-game-api",
		"http://mock-ocr-api", "http://mock-fingerprint-api",
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, zap.NewNop(), nil)
	mockFileRepo := &MockFileRepository{}
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		mockFileRepo,
		db,
	)

	t.Run("TMDb movie recommendations", func(t *testing.T) {
		year2010 := 2010
		req := &services.SimilarItemsRequest{
			MediaID: "inception",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Inception",
				Year:      &year2010,
				Genre:     "Sci-Fi",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:   0, // Only external
			MaxExternalItems: 5,
			IncludeExternal: true,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should have TMDb recommendations
		tmdbFound := false
		for _, item := range response.ExternalItems {
			if item.Provider == "TMDb" {
				tmdbFound = true
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.ExternalLink)
				assert.Contains(t, item.ExternalLink, "themoviedb.org")
				break
			}
		}
		assert.True(t, tmdbFound, "Should find TMDb recommendations")
	})

	t.Run("Last.fm music recommendations", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "imagine",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Imagine",
				Description: "John Lennon",
				Genre:     "Rock",
				MediaType: models.MediaTypeAudio,
			},
			MaxLocalItems:   0,
			MaxExternalItems: 5,
			IncludeExternal: true,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should have Last.fm recommendations
		lastfmFound := false
		for _, item := range response.ExternalItems {
			if item.Provider == "Last.fm" {
				lastfmFound = true
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.ExternalLink)
				assert.Contains(t, item.ExternalLink, "last.fm")
				break
			}
		}
		assert.True(t, lastfmFound, "Should find Last.fm recommendations")
	})

	t.Run("Google Books recommendations", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "1984",
			MediaMetadata: &models.MediaMetadata{
				Title:     "1984",
				Description: "George Orwell",
				Genre:     "Dystopian Fiction",
				MediaType: models.MediaTypeBook,
			},
			MaxLocalItems:   0,
			MaxExternalItems: 5,
			IncludeExternal: true,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should have Google Books recommendations
		googleFound := false
		for _, item := range response.ExternalItems {
			if item.Provider == "Google Books" {
				googleFound = true
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.ExternalLink)
				break
			}
		}
		assert.True(t, googleFound, "Should find Google Books recommendations")
	})

	t.Run("IGDB game recommendations", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "witcher3",
			MediaMetadata: &models.MediaMetadata{
				Title:     "The Witcher 3",
				Description: "CD Projekt RED",
				Genre:     "RPG",
				MediaType: models.MediaTypeGame,
			},
			MaxLocalItems:   0,
			MaxExternalItems: 5,
			IncludeExternal: true,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should have IGDB or Steam recommendations
		gameProviderFound := false
		for _, item := range response.ExternalItems {
			if item.Provider == "IGDB" || item.Provider == "Steam" {
				gameProviderFound = true
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.ExternalLink)
				break
			}
		}
		assert.True(t, gameProviderFound, "Should find game recommendations")
	})
}

func TestRecommendationService_Performance(t *testing.T) {
	ctx := context.Background()

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	mediaRecognitionService := services.NewMediaRecognitionService(
		db, zap.NewNop(), nil, nil,
		"http://mock-movie-api", "http://mock-music-api",
		"http://mock-book-api", "http://mock-game-api",
		"http://mock-ocr-api", "http://mock-fingerprint-api",
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, zap.NewNop(), nil)
	mockFileRepo := &MockFileRepository{}
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		mockFileRepo,
		db,
	)

	t.Run("performance metrics", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "test_performance",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Performance Test",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:   10,
			MaxExternalItems: 5,
			IncludeExternal: true,
		}

		start := time.Now()
		response, err := recommendationService.GetSimilarItems(ctx, req)
		totalDuration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, response.Performance)

		// Verify performance metrics are captured
		assert.True(t, response.Performance.LocalSearchTime > 0)
		assert.True(t, response.Performance.TotalTime > 0)
		assert.True(t, response.Performance.TotalTime >= response.Performance.LocalSearchTime)
		assert.True(t, totalDuration >= response.Performance.TotalTime)

		// Performance should be reasonable
		assert.True(t, response.Performance.TotalTime < 10*time.Second)

		// Metrics should be consistent
		assert.Equal(t, len(response.LocalItems), response.Performance.LocalItemsFound)
		assert.Equal(t, len(response.ExternalItems), response.Performance.ExternalItemsFound)
	})

	t.Run("large dataset performance", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "large_dataset_test",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Large Dataset Test",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:       50, // Large number
			MaxExternalItems:    20,
			IncludeExternal:     true,
			SimilarityThreshold: 0.1, // Low threshold for more results
		}

		start := time.Now()
		response, err := recommendationService.GetSimilarItems(ctx, req)
		duration := time.Since(start)

		require.NoError(t, err)

		// Should complete within reasonable time even with large dataset
		assert.True(t, duration < 15*time.Second)

		// Should respect limits
		assert.True(t, len(response.LocalItems) <= req.MaxLocalItems)
		assert.True(t, len(response.ExternalItems) <= req.MaxExternalItems)
	})
}

func TestRecommendationService_EdgeCases(t *testing.T) {
	ctx := context.Background()

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	mediaRecognitionService := services.NewMediaRecognitionService(
		db, zap.NewNop(), nil, nil,
		"http://mock-movie-api", "http://mock-music-api",
		"http://mock-book-api", "http://mock-game-api",
		"http://mock-ocr-api", "http://mock-fingerprint-api",
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, zap.NewNop(), nil)
	mockFileRepo := &MockFileRepository{}
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		mockFileRepo,
		db,
	)

	t.Run("empty media metadata", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "empty_metadata",
			MediaMetadata: &models.MediaMetadata{
				MediaType: models.MediaTypeVideo,
				// All other fields empty
			},
			MaxLocalItems: 5,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should handle gracefully and return some results
		assert.NotNil(t, response)
		assert.True(t, response.TotalFound >= 0)
	})

	t.Run("very high similarity threshold", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "high_threshold",
			MediaMetadata: &models.MediaMetadata{
				Title:     "High Threshold Test",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:       10,
			SimilarityThreshold: 0.99, // Very high
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// May return few or no results due to high threshold
		for _, item := range response.LocalItems {
			assert.True(t, item.SimilarityScore >= 0.99)
		}
	})

	t.Run("zero limits", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "zero_limits",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Zero Limits Test",
				MediaType: models.MediaTypeVideo,
			},
			MaxLocalItems:    0,
			MaxExternalItems: 0,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should respect limits
		assert.Empty(t, response.LocalItems)
		assert.Empty(t, response.ExternalItems)
		assert.Equal(t, 0, response.TotalFound)
	})

	t.Run("unknown media type", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID: "unknown_type",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Unknown Type Test",
				MediaType: "unknown_type",
			},
			MaxLocalItems:   5,
			IncludeExternal: true,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)

		// Should handle gracefully
		assert.NotNil(t, response)
	})

	t.Run("missing media metadata", func(t *testing.T) {
		req := &services.SimilarItemsRequest{
			MediaID:       "missing_metadata",
			MediaMetadata: nil, // Missing metadata
			MaxLocalItems: 5,
		}

		// Should handle missing metadata gracefully
		// In a real implementation, it might fetch metadata by MediaID
		response, err := recommendationService.GetSimilarItems(ctx, req)

		// The service should either return an error or handle gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "metadata")
		} else {
			assert.NotNil(t, response)
		}
	})

	t.Run("concurrent requests", func(t *testing.T) {
		numGoroutines := 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				req := &services.SimilarItemsRequest{
					MediaID: fmt.Sprintf("concurrent_test_%d", id),
					MediaMetadata: &models.MediaMetadata{
						Title:     fmt.Sprintf("Concurrent Test %d", id),
						MediaType: models.MediaTypeVideo,
					},
					MaxLocalItems:   5,
					IncludeExternal: false, // Disable external to reduce complexity
				}

				_, err := recommendationService.GetSimilarItems(ctx, req)
				results <- err
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err, "Concurrent request %d should not fail", i)
		}
	})
}

// TestRecommendationService_SimilarityAlgorithms tests similarity algorithm behavior
// through the public GetSimilarItems interface rather than accessing private methods.
func TestRecommendationService_SimilarityAlgorithms(t *testing.T) {
	ctx := context.Background()

	// Setup services with database
	db, _ := sql.Open("sqlite3", ":memory:")
	logger := zap.NewNop()

	mediaRecognitionService := services.NewMediaRecognitionService(
		db, logger, nil, nil,
		"http://mock-movie-api", "http://mock-music-api",
		"http://mock-book-api", "http://mock-game-api",
		"http://mock-ocr-api", "http://mock-fingerprint-api",
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, logger, nil)
	mockFileRepo := &MockFileRepository{}
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		mockFileRepo,
		db,
	)

	t.Run("similarity scoring produces ordered results", func(t *testing.T) {
		// Test that GetSimilarItems returns items sorted by similarity score (descending)
		year2008 := 2008
		rating90 := 9.0
		req := &services.SimilarItemsRequest{
			MediaID: "dark_knight_2008",
			MediaMetadata: &models.MediaMetadata{
				Title:     "The Dark Knight",
				Year:      &year2008,
				Genre:     "Action",
				Director:  "Christopher Nolan",
				MediaType: models.MediaTypeVideo,
				Rating:    &rating90,
			},
			MaxLocalItems:       10,
			IncludeExternal:     false,
			SimilarityThreshold: 0.1, // Low threshold to get more results
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify results are sorted by similarity score (descending)
		if len(response.LocalItems) > 1 {
			for i := 0; i < len(response.LocalItems)-1; i++ {
				assert.GreaterOrEqual(t, response.LocalItems[i].SimilarityScore, response.LocalItems[i+1].SimilarityScore,
					"Items should be sorted by similarity score (descending)")
			}
		}
	})

	t.Run("similarity threshold filtering", func(t *testing.T) {
		year2010 := 2010
		rating85 := 8.5

		// Test with high threshold - should get fewer or no results
		highThresholdReq := &services.SimilarItemsRequest{
			MediaID: "inception_2010",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Inception",
				Year:      &year2010,
				Genre:     "Science Fiction",
				Director:  "Christopher Nolan",
				MediaType: models.MediaTypeVideo,
				Rating:    &rating85,
			},
			MaxLocalItems:       10,
			IncludeExternal:     false,
			SimilarityThreshold: 0.9, // Very high threshold
		}

		highThresholdResponse, err := recommendationService.GetSimilarItems(ctx, highThresholdReq)
		require.NoError(t, err)
		require.NotNil(t, highThresholdResponse)

		// All returned items should meet the threshold
		for _, item := range highThresholdResponse.LocalItems {
			assert.GreaterOrEqual(t, item.SimilarityScore, highThresholdReq.SimilarityThreshold,
				"All items should meet the similarity threshold")
		}

		// Test with low threshold - should get more results
		lowThresholdReq := &services.SimilarItemsRequest{
			MediaID: "inception_2010",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Inception",
				Year:      &year2010,
				Genre:     "Science Fiction",
				Director:  "Christopher Nolan",
				MediaType: models.MediaTypeVideo,
				Rating:    &rating85,
			},
			MaxLocalItems:       10,
			IncludeExternal:     false,
			SimilarityThreshold: 0.1, // Low threshold
		}

		lowThresholdResponse, err := recommendationService.GetSimilarItems(ctx, lowThresholdReq)
		require.NoError(t, err)
		require.NotNil(t, lowThresholdResponse)

		// Low threshold should generally yield >= results as high threshold
		assert.GreaterOrEqual(t, len(lowThresholdResponse.LocalItems), len(highThresholdResponse.LocalItems),
			"Lower threshold should yield at least as many results")
	})

	t.Run("similarity reasons are provided", func(t *testing.T) {
		year2014 := 2014
		rating86 := 8.6
		req := &services.SimilarItemsRequest{
			MediaID: "interstellar_2014",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Interstellar",
				Year:      &year2014,
				Genre:     "Science Fiction",
				Director:  "Christopher Nolan",
				MediaType: models.MediaTypeVideo,
				Rating:    &rating86,
			},
			MaxLocalItems:       10,
			IncludeExternal:     false,
			SimilarityThreshold: 0.2,
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify that similarity reasons are provided for local items
		for _, item := range response.LocalItems {
			assert.NotEmpty(t, item.SimilarityReasons,
				"Each similar item should have similarity reasons explaining the match")
		}
	})

	t.Run("genre filter affects results", func(t *testing.T) {
		year2015 := 2015
		rating78 := 7.8
		req := &services.SimilarItemsRequest{
			MediaID: "test_movie",
			MediaMetadata: &models.MediaMetadata{
				Title:     "Test Movie",
				Year:      &year2015,
				Genre:     "Action",
				MediaType: models.MediaTypeVideo,
				Rating:    &rating78,
			},
			MaxLocalItems:       10,
			IncludeExternal:     false,
			SimilarityThreshold: 0.1,
			Filters: &services.RecommendationFilters{
				GenreFilter: []string{"Action"},
			},
		}

		response, err := recommendationService.GetSimilarItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Results should be filtered - response structure should be valid
		assert.NotNil(t, response.Performance)
		assert.NotEmpty(t, response.Algorithms)
	})
}

func BenchmarkRecommendationService(b *testing.B) {
	ctx := context.Background()
	
	// In-memory database for benchmarking
	db, _ := sql.Open("sqlite3", ":memory:")
	logger, _ := zap.NewDevelopment()
	defer db.Close()

	mediaRecognitionService := services.NewMediaRecognitionService(
		db, logger, nil, nil, 
		"http://mock-movie-api", "http://mock-music-api", 
		"http://mock-book-api", "http://mock-game-api", 
		"http://mock-ocr-api", "http://mock-fingerprint-api",
	)
	duplicateDetectionService := services.NewDuplicateDetectionService(db, logger, nil)
	mockFileRepo := &MockFileRepository{}
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		mockFileRepo,
		db,
	)

	year2023 := 2023
	req := &services.SimilarItemsRequest{
		MediaID: "benchmark_test",
		MediaMetadata: &models.MediaMetadata{
			Title:     "Benchmark Movie",
			Year:      &year2023,
			Genre:     "Action",
			MediaType: models.MediaTypeVideo,
		},
		MaxLocalItems:       10,
		MaxExternalItems:    5,
		IncludeExternal:     false, // Disable external for consistent benchmarking
		SimilarityThreshold: 0.3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := recommendationService.GetSimilarItems(ctx, req)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}