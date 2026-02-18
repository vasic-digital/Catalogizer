package services

import (
	"catalogizer/database"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewCacheService(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()

	service := NewCacheService(mockDB, mockLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockLogger, service.logger)
}

func TestCacheService_Set(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	// With nil DB, this should not panic and should return nil
	testData := map[string]string{"key": "value"}
	err := service.Set(context.Background(), "test_key", testData, time.Hour)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_Get(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	var result map[string]string
	found, err := service.Get(context.Background(), "test_key", &result)

	// Should not panic and should return not found with nil DB
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestCacheService_Delete(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.Delete(context.Background(), "test_key")

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_Clear(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.Clear(context.Background(), "")

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_Clear_WithPattern(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.Clear(context.Background(), "test_%")

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_SetMediaMetadata(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	testData := map[string]string{"title": "Test Movie"}
	err := service.SetMediaMetadata(context.Background(), 123, "movie", "tmdb", testData, 0.95)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetMediaMetadata(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	var result map[string]string
	found, quality, err := service.GetMediaMetadata(context.Background(), 123, "movie", "tmdb", &result)

	// Should not panic and should return not found with nil DB
	assert.False(t, found)
	assert.Equal(t, 0.0, quality)
	assert.NoError(t, err)
}

func TestCacheService_SetAPIResponse(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	requestData := map[string]string{"query": "test"}
	responseData := map[string]interface{}{"results": []string{"item1", "item2"}}

	err := service.SetAPIResponse(context.Background(), "tmdb", "/search/movie", requestData, responseData, 200, time.Hour)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetAPIResponse(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	requestData := map[string]string{"query": "test"}
	var result map[string]interface{}

	found, statusCode, err := service.GetAPIResponse(context.Background(), "tmdb", "/search/movie", requestData, &result)

	// Should not panic and should return not found with nil DB
	assert.False(t, found)
	assert.Equal(t, 0, statusCode)
	assert.NoError(t, err)
}

func TestCacheService_SetThumbnail(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.SetThumbnail(context.Background(), 123, 30, "http://example.com/thumb.jpg", 320, 240, 10240)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetThumbnail(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	thumbnail, err := service.GetThumbnail(context.Background(), 123, 30, 320, 240)

	// Should not panic and should return nil with nil DB
	assert.Nil(t, thumbnail)
	assert.NoError(t, err)
}

func TestCacheService_SetTranslation(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.SetTranslation(context.Background(), "Hello", "en", "es", "google", "Hola")

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetTranslation(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	translation, found, err := service.GetTranslation(context.Background(), "Hello", "en", "es", "google")

	// Should not panic and should return not found with nil DB
	assert.Empty(t, translation)
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestCacheService_SetSubtitle(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	content := "Test subtitle content"
	subtitle := &SubtitleTrack{
		Language: "en",
		Source:   "opensubtitles",
		Content:  &content,
	}

	err := service.SetSubtitle(context.Background(), 123, "en", "opensubtitles", subtitle)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetSubtitle(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	subtitle, found, err := service.GetSubtitle(context.Background(), 123, "en", "opensubtitles")

	// Should not panic and should return nil/false with nil DB
	assert.Nil(t, subtitle)
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestCacheService_SetLyrics(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	lyrics := &LyricsData{
		Source:  "genius",
		Content: "Test lyrics content",
	}

	err := service.SetLyrics(context.Background(), "Test Artist", "Test Song", "genius", lyrics)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetLyrics(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	lyrics, found, err := service.GetLyrics(context.Background(), "Test Artist", "Test Song", "genius")

	// Should not panic and should return nil/false with nil DB
	assert.Nil(t, lyrics)
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestCacheService_SetCoverArt(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	url := "http://example.com/cover.jpg"
	width := 500
	height := 500
	size := int64(51200)
	coverArt := &CoverArt{
		Source: "itunes",
		URL:    &url,
		Width:  &width,
		Height: &height,
		Size:   &size,
	}

	err := service.SetCoverArt(context.Background(), "Test Artist", "Test Album", "itunes", coverArt)

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_GetCoverArt(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	coverArt, found, err := service.GetCoverArt(context.Background(), "Test Artist", "Test Album", "itunes")

	// Should not panic and should return nil/false with nil DB
	assert.Nil(t, coverArt)
	assert.False(t, found)
	assert.NoError(t, err)
}

func TestCacheService_GetStats(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	stats, err := service.GetStats(context.Background())

	// Should not panic and should return empty stats with nil DB
	assert.NotNil(t, stats)
	assert.NoError(t, err)
}

func TestCacheService_CleanupExpired(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.CleanupExpired(context.Background())

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_Warmup(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.Warmup(context.Background())

	// Should not panic and should return nil (no-op implementation)
	assert.NoError(t, err)
}

func TestCacheService_InvalidateByPattern(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	err := service.InvalidateByPattern(context.Background(), "test_%")

	// Should not panic and should succeed with nil DB (no-op)
	assert.NoError(t, err)
}

func TestCacheService_HashString(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	hash1 := service.hashString("test")
	hash2 := service.hashString("test")
	hash3 := service.hashString("different")

	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
	assert.NotEmpty(t, hash1)
}

func TestCacheService_HashRequest(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewCacheService(mockDB, mockLogger)

	request1 := map[string]string{"query": "test"}
	request2 := map[string]string{"query": "test"}
	request3 := map[string]string{"query": "different"}

	hash1, err1 := service.hashRequest(request1)
	hash2, err2 := service.hashRequest(request2)
	hash3, err3 := service.hashRequest(request3)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
	assert.NotEmpty(t, hash1)
}
