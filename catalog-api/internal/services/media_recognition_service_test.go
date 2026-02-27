package services

import (
	"catalogizer/database"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockMediaRecognitionCacheService is a mock implementation of CacheServiceInterface
type MockMediaRecognitionCacheService struct {
	mock.Mock
}

func (m *MockMediaRecognitionCacheService) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	args := m.Called(ctx, key, dest)
	return args.Bool(0), args.Error(1)
}

func (m *MockMediaRecognitionCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// MockMediaRecognitionTranslationService is a mock implementation of TranslationServiceInterface
type MockMediaRecognitionTranslationService struct {
	mock.Mock
}

func (m *MockMediaRecognitionTranslationService) TranslateText(ctx context.Context, request TranslationRequest) (*TranslationResult, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*TranslationResult), args.Error(1)
}

// MockRecognitionProvider is a mock implementation of RecognitionProvider
type MockRecognitionProvider struct {
	mock.Mock
}

func (m *MockRecognitionProvider) RecognizeMedia(ctx context.Context, req *MediaRecognitionRequest) (*MediaRecognitionResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*MediaRecognitionResult), args.Error(1)
}

func (m *MockRecognitionProvider) GetProviderName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRecognitionProvider) SupportsMediaType(mediaType MediaType) bool {
	args := m.Called(mediaType)
	return args.Bool(0)
}

func (m *MockRecognitionProvider) GetConfidenceThreshold() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func TestNewMediaRecognitionService(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(
		mockDB,
		mockLogger,
		mockCache,
		mockTranslation,
		"https://api.themoviedb.org/3",
		"https://api.musicbrainz.org",
		"https://www.googleapis.com/books/v1",
		"https://api.igdb.com/v4",
		"https://api.ocr.space",
		"https://api.acoustid.org",
	)

	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockLogger, service.logger)
	assert.Equal(t, mockCache, service.cacheService)
	assert.Equal(t, mockTranslation, service.translationService)
	assert.Equal(t, "https://api.themoviedb.org/3", service.movieAPIBaseURL)
	assert.Equal(t, "https://api.musicbrainz.org", service.musicAPIBaseURL)
	assert.Equal(t, "https://www.googleapis.com/books/v1", service.bookAPIBaseURL)
	assert.Equal(t, "https://api.igdb.com/v4", service.gameAPIBaseURL)
	assert.Equal(t, "https://api.ocr.space", service.ocrAPIBaseURL)
	assert.Equal(t, "https://api.acoustid.org", service.fingerprintAPIBaseURL)
}

func TestMediaRecognitionService_RecognizeMedia_Cached(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	req := &MediaRecognitionRequest{
		FilePath: "/path/to/movie.mp4",
		FileHash: "abc123",
	}

	cachedResult := MediaRecognitionResult{
		MediaID:    "cached_123",
		MediaType:  MediaTypeMovie,
		Title:      "Cached Movie",
		Confidence: 0.95,
	}

	// Mock cache hit
	mockCache.On("Get", mock.Anything, "media_recognition:abc123", mock.AnythingOfType("*services.MediaRecognitionResult")).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*MediaRecognitionResult)
		*dest = cachedResult
	}).Return(true, nil)

	result, err := service.RecognizeMedia(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "cached_123", result.MediaID)
	assert.Equal(t, "Cached Movie", result.Title)

	mockCache.AssertExpectations(t)
}

func TestMediaRecognitionService_RecognizeMedia_NoCache(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	req := &MediaRecognitionRequest{
		FilePath: "/path/to/movie.mp4",
		FileHash: "abc123",
		FileName: "Test Movie.mp4",
		MimeType: "video/mp4",
	}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, "media_recognition:abc123", mock.AnythingOfType("*services.MediaRecognitionResult")).Return(false, nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, "media_recognition:abc123", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	// Since no providers are registered, this should fail
	_, err := service.RecognizeMedia(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recognition providers available")
}

func TestMediaRecognitionService_DetectMediaType(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	tests := []struct {
		name         string
		req          *MediaRecognitionRequest
		expectedType MediaType
	}{
		{
			name: "Video file",
			req: &MediaRecognitionRequest{
				MimeType: "video/mp4",
				FileName: "movie.mp4",
			},
			expectedType: MediaTypeMovie,
		},
		{
			name: "Audio file",
			req: &MediaRecognitionRequest{
				MimeType: "audio/mpeg",
				FileName: "song.mp3",
			},
			expectedType: MediaTypeMusic,
		},
		{
			name: "Document file (docx)",
			req: &MediaRecognitionRequest{
				MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				FileName: "document.docx",
			},
			expectedType: MediaTypeDocument,
		},
		{
			name: "Comic book (pdf)",
			req: &MediaRecognitionRequest{
				MimeType: "application/pdf",
				FileName: "comic.pdf",
			},
			expectedType: MediaTypeComicBook,
		},
		{
			name: "Book file (epub)",
			req: &MediaRecognitionRequest{
				MimeType: "application/epub+zip",
				FileName: "book.epub",
			},
			expectedType: MediaTypeBook,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			detectedType, confidence := service.detectMediaType(test.req)
			assert.Equal(t, test.expectedType, detectedType)
			assert.Greater(t, confidence, 0.0)
		})
	}
}

func TestMediaRecognitionService_GetProvidersForMediaType(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	// Test with movie type
	providers := service.getProvidersForMediaType(MediaTypeMovie)
	assert.NotNil(t, providers)
	// Should return some default providers, but since we don't have real providers registered,
	// this might be empty or have mock providers
}

func TestMediaRecognitionService_EnhanceRecognitionResult(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	result := &MediaRecognitionResult{
		MediaID:   "test_123",
		MediaType: MediaTypeMovie,
		Title:     "Test Movie",
	}

	req := &MediaRecognitionRequest{
		FilePath: "/path/to/movie.mp4",
		Metadata: map[string]string{
			"duration": "7200", // 2 hours in seconds
			"width":    "1920",
			"height":   "1080",
		},
	}

	// This should not panic and should enhance the result
	service.enhanceRecognitionResult(context.Background(), result, req)

	assert.NotNil(t, result)
}

func TestMediaRecognitionService_FindDuplicates(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	result := &MediaRecognitionResult{
		MediaID:   "test_123",
		MediaType: MediaTypeMovie,
		Title:     "Test Movie",
		IMDbID:    "tt1234567",
	}

	// Since we don't have a real database, this should return empty results without panicking
	duplicates, err := service.findDuplicates(context.Background(), result)

	// We expect this to not panic and return empty results
	assert.NotNil(t, duplicates) // Should be a valid slice, even if empty
	assert.NoError(t, err)       // Should not error with nil DB
	assert.Empty(t, duplicates)  // Should be empty since no DB
}

func TestMediaRecognitionService_TranslateMetadata(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	result := &MediaRecognitionResult{
		MediaID:     "test_123",
		MediaType:   MediaTypeMovie,
		Title:       "Test Movie",
		Description: "A test movie description",
	}

	languages := []string{"es", "fr"}

	// Mock translation service
	mockTranslation.On("TranslateText", mock.Anything, mock.AnythingOfType("services.TranslationRequest")).Return(&TranslationResult{
		OriginalText:   "Test Movie",
		TranslatedText: "Película de Prueba",
		SourceLanguage: "en",
		TargetLanguage: "es",
		Confidence:     0.9,
	}, nil).Maybe()

	translations, err := service.translateMetadata(context.Background(), result, languages)

	assert.NoError(t, err)
	assert.NotNil(t, translations)
}

func TestMediaRecognitionService_StoreRecognitionResult(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite) // nil DB for testing
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	result := &MediaRecognitionResult{
		MediaID:        "test_123",
		MediaType:      MediaTypeMovie,
		Title:          "Test Movie",
		Description:    "A test movie",
		Year:           2023,
		Confidence:     0.95,
		RecognizedAt:   time.Now(),
		ProcessingTime: 1500,
	}

	req := &MediaRecognitionRequest{
		FilePath: "/path/to/movie.mp4",
		FileHash: "abc123",
	}

	// Since we don't have a real database, this will likely fail
	err := service.storeRecognitionResult(context.Background(), result, req)

	// We expect this to not panic, but may return an error due to no database
	_ = err // We don't assert on error since database operations may fail in test
}

func TestMockRecognitionProvider(t *testing.T) {
	provider := &MockRecognitionProvider{}

	// Test GetProviderName
	provider.On("GetProviderName").Return("TestProvider")
	assert.Equal(t, "TestProvider", provider.GetProviderName())

	// Test SupportsMediaType
	provider.On("SupportsMediaType", MediaTypeMovie).Return(true)
	assert.True(t, provider.SupportsMediaType(MediaTypeMovie))

	// Test GetConfidenceThreshold
	provider.On("GetConfidenceThreshold").Return(0.8)
	assert.Equal(t, 0.8, provider.GetConfidenceThreshold())

	// Test RecognizeMedia
	req := &MediaRecognitionRequest{
		FilePath: "/test.mp4",
	}
	expectedResult := &MediaRecognitionResult{
		MediaID:    "test_123",
		MediaType:  MediaTypeMovie,
		Title:      "Test Movie",
		Confidence: 0.9,
	}

	provider.On("RecognizeMedia", mock.Anything, req).Return(expectedResult, nil)

	result, err := provider.RecognizeMedia(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test_123", result.MediaID)
	assert.Equal(t, "Test Movie", result.Title)

	provider.AssertExpectations(t)
}

// Test helper functions
func TestMediaTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected MediaType
	}{
		{"movie", MediaTypeMovie},
		{"tv", MediaTypeTV},
		{"music", MediaTypeMusic},
		{"book", MediaTypeBook},
		{"game", MediaTypeGame},
		{"software", MediaTypeSoftware},
		{"image", MediaTypeImage},
		{"document", MediaTypeDocument},
		{"unknown", MediaTypeUnknown},
	}

	for _, test := range tests {
		result := MediaType(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestMediaRecognitionResult_JSON(t *testing.T) {
	result := &MediaRecognitionResult{
		MediaID:     "test_123",
		MediaType:   MediaTypeMovie,
		Title:       "Test Movie",
		Description: "A test movie",
		Year:        2023,
		Genres:      []string{"Action", "Adventure"},
		Director:    "Test Director",
		Cast: []Person{
			{Name: "Actor 1", Role: "Lead"},
			{Name: "Actor 2", Role: "Supporting"},
		},
		Rating:         8.5,
		Confidence:     0.95,
		RecognizedAt:   time.Now(),
		ProcessingTime: 1500,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.NotNil(t, jsonData)
	assert.Contains(t, string(jsonData), "Test Movie")
	assert.Contains(t, string(jsonData), "test_123")
}

func TestMediaRecognitionService_RecognizeMediaBatch(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	requests := []*MediaRecognitionRequest{
		{
			FilePath: "/path/to/movie1.mp4",
			FileHash: "hash1",
			FileName: "Movie1.mp4",
			MimeType: "video/mp4",
		},
		{
			FilePath: "/path/to/movie2.mp4",
			FileHash: "hash2",
			FileName: "Movie2.mp4",
			MimeType: "video/mp4",
		},
	}

	// Mock cache misses for both
	mockCache.On("Get", mock.Anything, "media_recognition:hash1", mock.AnythingOfType("*services.MediaRecognitionResult")).Return(false, nil)
	mockCache.On("Get", mock.Anything, "media_recognition:hash2", mock.AnythingOfType("*services.MediaRecognitionResult")).Return(false, nil)

	// Since no providers are available, this should return empty results without panicking
	results, err := service.RecognizeMediaBatch(context.Background(), requests)

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)
	// Results should be nil since no providers are available
	assert.Nil(t, results[0])
	assert.Nil(t, results[1])
}

func TestMediaRecognitionService_GetRecognitionStats(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	// With nil DB, this should return empty stats without panicking
	stats, err := service.GetRecognitionStats(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_recognized")
	assert.Contains(t, stats, "by_type")
}

func TestMediaRecognitionService_DetectMediaType_EdgeCases(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	tests := []struct {
		name         string
		req          *MediaRecognitionRequest
		expectedType MediaType
	}{
		{
			name: "Unknown MIME type",
			req: &MediaRecognitionRequest{
				MimeType: "application/octet-stream",
				FileName: "unknown.bin",
			},
			expectedType: MediaTypeUnknown,
		},
		{
			name: "Empty MIME type",
			req: &MediaRecognitionRequest{
				MimeType: "",
				FileName: "file.txt",
			},
			expectedType: MediaTypeDocument,
		},
		{
			name: "Image file",
			req: &MediaRecognitionRequest{
				MimeType: "image/jpeg",
				FileName: "photo.jpg",
			},
			expectedType: MediaTypeImage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			detectedType, confidence := service.detectMediaType(test.req)
			assert.Equal(t, test.expectedType, detectedType)
			assert.Greater(t, confidence, 0.0)
		})
	}
}

func TestMediaRecognitionService_RecognizeMedia_ErrorHandling(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	mockCache := &MockMediaRecognitionCacheService{}
	mockTranslation := &MockMediaRecognitionTranslationService{}

	service := NewMediaRecognitionService(mockDB, mockLogger, mockCache, mockTranslation, "", "", "", "", "", "")

	// Test with invalid request
	req := &MediaRecognitionRequest{
		FilePath: "",
		FileHash: "",
	}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, "media_recognition:", mock.AnythingOfType("*services.MediaRecognitionResult")).Return(false, nil)

	// This should fail due to no providers
	result, err := service.RecognizeMedia(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no recognition providers available")
}
