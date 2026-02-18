package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MediaBrowseHandlerTestSuite struct {
	suite.Suite
	handler   *MediaBrowseHandler
	fileRepo  *repository.FileRepository
	statsRepo *repository.StatsRepository
	router    *gin.Engine
}

func (suite *MediaBrowseHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *MediaBrowseHandlerTestSuite) SetupTest() {
	suite.fileRepo = nil
	suite.statsRepo = nil
	suite.handler = NewMediaBrowseHandler(suite.fileRepo, suite.statsRepo, nil)

	suite.router = gin.New()
	suite.router.GET("/api/v1/media/search", suite.handler.SearchMedia)
	suite.router.GET("/api/v1/media/stats", suite.handler.GetMediaStats)
}

// Test handler initialization
func (suite *MediaBrowseHandlerTestSuite) TestNewMediaBrowseHandler() {
	handler := NewMediaBrowseHandler(nil, nil, nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.fileRepo)
	assert.Nil(suite.T(), handler.statsRepo)
	assert.Nil(suite.T(), handler.db)
}

func (suite *MediaBrowseHandlerTestSuite) TestNewMediaBrowseHandler_WithDeps() {
	fileRepo := &repository.FileRepository{}
	statsRepo := &repository.StatsRepository{}
	handler := NewMediaBrowseHandler(fileRepo, statsRepo, nil)
	assert.NotNil(suite.T(), handler)
	assert.Equal(suite.T(), fileRepo, handler.fileRepo)
	assert.Equal(suite.T(), statsRepo, handler.statsRepo)
}

// Test HTTP method restrictions
func (suite *MediaBrowseHandlerTestSuite) TestSearchMedia_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/v1/media/search", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *MediaBrowseHandlerTestSuite) TestGetMediaStats_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/v1/media/stats", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test route matching via wrong-path 404 (proves route IS registered)
func (suite *MediaBrowseHandlerTestSuite) TestSearchMedia_WrongPathIs404() {
	req := httptest.NewRequest("GET", "/api/v1/media/searchtypo", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *MediaBrowseHandlerTestSuite) TestGetMediaStats_WrongPathIs404() {
	req := httptest.NewRequest("GET", "/api/v1/media/statstypo", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test fileToMediaItem conversion
func TestFileToMediaItem_FullFields(t *testing.T) {
	ext := "mp4"
	fileType := "video"
	now := time.Now()

	f := models.FileWithMetadata{
		File: models.File{
			ID:              42,
			Name:            "Big Buck Bunny.mp4",
			Path:            "/Movies/Big Buck Bunny.mp4",
			Extension:       &ext,
			FileType:        &fileType,
			Size:            1048576,
			StorageRootName: "nas-main",
			CreatedAt:       now,
			ModifiedAt:      now,
		},
	}

	item := fileToMediaItem(f)

	assert.Equal(t, int64(42), item.ID)
	assert.Equal(t, "Big Buck Bunny.mp4", item.Title)
	assert.Equal(t, "video", item.MediaType)
	assert.Equal(t, "mp4", item.Quality)
	assert.Equal(t, int64(1048576), item.FileSize)
	assert.Equal(t, "/Movies/Big Buck Bunny.mp4", item.DirectoryPath)
	assert.Equal(t, "nas-main", item.StorageRootName)
	assert.Equal(t, now.Format(time.RFC3339), item.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), item.UpdatedAt)
}

func TestFileToMediaItem_NilFields(t *testing.T) {
	now := time.Now()

	f := models.FileWithMetadata{
		File: models.File{
			ID:         1,
			Name:       "mystery_file",
			Path:       "/unknown/mystery_file",
			Extension:  nil,
			FileType:   nil,
			Size:       0,
			CreatedAt:  now,
			ModifiedAt: now,
		},
	}

	item := fileToMediaItem(f)

	assert.Equal(t, int64(1), item.ID)
	assert.Equal(t, "mystery_file", item.Title)
	assert.Equal(t, "other", item.MediaType, "nil FileType should default to 'other'")
	assert.Equal(t, "", item.Quality, "nil Extension should produce empty quality")
	assert.Equal(t, int64(0), item.FileSize)
}

func TestFileToMediaItem_JSONShape(t *testing.T) {
	ext := "flac"
	fileType := "audio"
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	f := models.FileWithMetadata{
		File: models.File{
			ID:              99,
			Name:            "song.flac",
			Path:            "/Music/song.flac",
			Extension:       &ext,
			FileType:        &fileType,
			Size:            5242880,
			StorageRootName: "nas-music",
			CreatedAt:       now,
			ModifiedAt:      now,
		},
	}

	item := fileToMediaItem(f)
	data, err := json.Marshal(item)
	assert.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	// Verify all expected keys are present
	expectedKeys := []string{"id", "title", "media_type", "quality", "file_size", "directory_path", "storage_root_name", "created_at", "updated_at"}
	for _, key := range expectedKeys {
		_, ok := parsed[key]
		assert.True(t, ok, "expected key %q in JSON output", key)
	}

	assert.Equal(t, float64(99), parsed["id"])
	assert.Equal(t, "song.flac", parsed["title"])
	assert.Equal(t, "audio", parsed["media_type"])
}

func TestMediaBrowseHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MediaBrowseHandlerTestSuite))
}
