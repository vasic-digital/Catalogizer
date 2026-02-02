package services

import (
	"context"
	"testing"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileRepository is a mock implementation of FileRepositoryInterface
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) SearchFiles(ctx context.Context, filter models.SearchFilter, pagination models.PaginationOptions, sort models.SortOptions) (*models.SearchResult, error) {
	args := m.Called(ctx, filter, pagination, sort)
	return args.Get(0).(*models.SearchResult), args.Error(1)
}

func TestNewRecommendationService(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.fileRepository)
	assert.Equal(t, mockMediaRecognition, service.mediaRecognitionService)
	assert.Equal(t, mockDuplicateDetection, service.duplicateDetectionService)
	assert.NotNil(t, service.httpClient)
	assert.Equal(t, "https://api.themoviedb.org/3", service.tmdbBaseURL)
}

func TestRecommendationService_GetSimilarItems(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	req := &SimilarItemsRequest{
		MediaMetadata: &models.MediaMetadata{
			ID:        1,
			Title:     "Test Movie",
			MediaType: "movie",
			Genre:     "Action",
		},
		MaxLocalItems:    10,
		MaxExternalItems: 5,
	}

	// Mock the repository search
	mockFiles := []models.FileWithMetadata{
		{
			File: models.File{
				ID:   2,
				Name: "Similar Movie 1",
				Size: 1000000,
			},
			Metadata: []models.FileMetadata{
				{Key: "media_type", Value: "movie"},
				{Key: "title", Value: "Similar Movie 1"},
				{Key: "genre", Value: "Action"},
			},
		},
	}

	mockResult := &models.SearchResult{
		Files:      mockFiles,
		TotalCount: 1,
	}

	mockRepo.On("SearchFiles", mock.Anything, mock.AnythingOfType("models.SearchFilter"), mock.AnythingOfType("models.PaginationOptions"), mock.AnythingOfType("models.SortOptions")).Return(mockResult, nil)

	result, err := service.GetSimilarItems(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(result.LocalItems), 0)
	assert.NotNil(t, result.Performance)
	assert.Greater(t, result.TotalFound, 0)

	mockRepo.AssertExpectations(t)
}

func TestRecommendationService_FindLocalSimilarItems(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	req := &SimilarItemsRequest{
		MediaMetadata: &models.MediaMetadata{
			ID:        1,
			Title:     "Test Movie",
			MediaType: "movie",
			Genre:     "Action",
		},
		MaxLocalItems: 10,
	}

	// Mock repository search
	mockFiles := []models.FileWithMetadata{
		{
			File: models.File{
				ID:   2,
				Name: "Similar Movie",
				Size: 1000000,
			},
			Metadata: []models.FileMetadata{
				{Key: "media_type", Value: "movie"},
				{Key: "title", Value: "Similar Movie"},
				{Key: "genre", Value: "Action"},
			},
		},
	}

	mockResult := &models.SearchResult{
		Files:      mockFiles,
		TotalCount: 1,
	}

	mockRepo.On("SearchFiles", mock.Anything, mock.AnythingOfType("models.SearchFilter"), mock.AnythingOfType("models.PaginationOptions"), mock.AnythingOfType("models.SortOptions")).Return(mockResult, nil)

	items, err := service.findLocalSimilarItems(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Greater(t, len(items), 0)
	assert.Equal(t, "2", items[0].MediaID)

	mockRepo.AssertExpectations(t)
}

func TestRecommendationService_ConvertFileToMediaMetadata(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	fileWithMeta := &models.FileWithMetadata{
		File: models.File{
			ID:   1,
			Name: "Test Movie.mp4",
			Size: 1000000,
		},
		Metadata: []models.FileMetadata{
			{Key: "media_type", Value: "movie"},
			{Key: "title", Value: "Test Movie"},
			{Key: "genre", Value: "Action"},
			{Key: "year", Value: "2023"},
			{Key: "rating", Value: "8.5"},
			{Key: "duration", Value: "120"},
			{Key: "director", Value: "Test Director"},
		},
	}

	metadata := service.convertFileToMediaMetadata(fileWithMeta)

	assert.NotNil(t, metadata)
	assert.Equal(t, int64(1), metadata.ID)
	assert.Equal(t, "Test Movie", metadata.Title)
	assert.Equal(t, "movie", metadata.MediaType)
	assert.Equal(t, "Action", metadata.Genre)
	assert.Equal(t, 2023, *metadata.Year)
	assert.Equal(t, 8.5, *metadata.Rating)
	assert.Equal(t, 120, *metadata.Duration)
	assert.Equal(t, "Test Director", metadata.Director)
	assert.Equal(t, int64(1000000), *metadata.FileSize)
}

func TestRecommendationService_QuerySimilarMediaFromDatabase(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	originalMetadata := &models.MediaMetadata{
		ID:        1,
		Title:     "Test Movie",
		MediaType: "movie",
		Genre:     "Action",
	}

	// Mock repository search
	mockFiles := []models.FileWithMetadata{
		{
			File: models.File{
				ID:   2,
				Name: "Similar Movie",
			},
			Metadata: []models.FileMetadata{
				{Key: "media_type", Value: "movie"},
			},
		},
	}

	mockResult := &models.SearchResult{
		Files:      mockFiles,
		TotalCount: 1,
	}

	mockRepo.On("SearchFiles", mock.Anything, mock.AnythingOfType("models.SearchFilter"), mock.AnythingOfType("models.PaginationOptions"), mock.AnythingOfType("models.SortOptions")).Return(mockResult, nil).Twice()

	results, err := service.querySimilarMediaFromDatabase(context.Background(), originalMetadata)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	mockRepo.AssertExpectations(t)
}

func TestRecommendationService_CalculateLocalSimilarity(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	original := &models.MediaMetadata{
		ID:        1,
		Title:     "Test Movie",
		MediaType: "movie",
		Genre:     "Action",
		Year:      &[]int{2023}[0],
		Rating:    &[]float64{8.5}[0],
	}

	candidate := &models.MediaMetadata{
		ID:        2,
		Title:     "Similar Movie",
		MediaType: "movie",
		Genre:     "Action",
		Year:      &[]int{2023}[0],
		Rating:    &[]float64{8.0}[0],
	}

	similarity, reasons := service.calculateLocalSimilarity(original, candidate)

	assert.Greater(t, similarity, 0.0)
	assert.LessOrEqual(t, similarity, 1.0)
	assert.NotEmpty(t, reasons)
}

func TestRecommendationService_PassesFilters(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	media := &models.MediaMetadata{
		Title:  "Test Movie",
		Genre:  "Action",
		Year:   &[]int{2023}[0],
		Rating: &[]float64{8.5}[0],
	}

	filters := &RecommendationFilters{
		GenreFilter: []string{"Action"},
		YearRange: &YearRange{
			StartYear: 2020,
			EndYear:   2025,
		},
		RatingRange: &RatingRange{
			MinRating: 8.0,
			MaxRating: 9.0,
		},
	}

	passes := service.passesFilters(media, 0.8, filters)
	assert.True(t, passes)

	// Test with non-matching genre
	filters.GenreFilter = []string{"Comedy"}
	passes = service.passesFilters(media, 0.8, filters)
	assert.False(t, passes)
}

func TestRecommendationService_FindExternalSimilarItems(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	req := &SimilarItemsRequest{
		MediaMetadata: &models.MediaMetadata{
			ID:        1,
			Title:     "Test Movie",
			MediaType: "movie",
			Genre:     "Action",
		},
		IncludeExternal: true,
	}

	// This test verifies the method exists and doesn't panic
	// External API calls are not mocked, so it returns an empty slice
	items, err := service.findExternalSimilarItems(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, 0, len(items))
}

func TestRecommendationService_GenerateDetailLink(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	media := &models.MediaMetadata{
		ID: 123,
	}

	link := service.generateDetailLink(media)
	assert.Equal(t, "/detail/123", link)
}

func TestRecommendationService_GeneratePlayLink(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	media := &models.MediaMetadata{
		ID: 123,
	}

	link := service.generatePlayLink(media)
	assert.Equal(t, "/play/123", link)
}

func TestRecommendationService_GenerateDownloadLink(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	media := &models.MediaMetadata{
		ID: 123,
	}

	link := service.generateDownloadLink(media)
	assert.Equal(t, "/download/123", link)
}

func TestRecommendationService_GenerateMediaID(t *testing.T) {
	mockRepo := &MockFileRepository{}
	mockMediaRecognition := &MediaRecognitionService{}
	mockDuplicateDetection := &DuplicateDetectionService{}

	service := NewRecommendationService(mockMediaRecognition, mockDuplicateDetection, mockRepo, nil)

	media := &models.MediaMetadata{
		ID: 123,
	}

	id := service.generateMediaID(media)
	assert.Equal(t, "123", id)
}
