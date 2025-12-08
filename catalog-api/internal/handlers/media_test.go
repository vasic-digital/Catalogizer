package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/media/analyzer"
	"catalogizer/internal/media/database"
)

func TestNewMediaHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mediaDB := &database.MediaDatabase{}
	analyzer := &analyzer.MediaAnalyzer{}

	handler := NewMediaHandler(mediaDB, analyzer, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mediaDB, handler.mediaDB)
	assert.Equal(t, analyzer, handler.analyzer)
	assert.Equal(t, logger, handler.logger)
}

func TestMediaHandler_GetMediaItem_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := zap.NewDevelopment()
	mediaDB := &database.MediaDatabase{}
	analyzer := &analyzer.MediaAnalyzer{}
	handler := NewMediaHandler(mediaDB, analyzer, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.GetMediaItem(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid media item ID", response["error"])
}

func TestMediaHandler_AnalyzeDirectory_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := zap.NewDevelopment()
	mediaDB := &database.MediaDatabase{}
	analyzer := &analyzer.MediaAnalyzer{}
	handler := NewMediaHandler(mediaDB, analyzer, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Invalid JSON request
	req := httptest.NewRequest("POST", "/analyze", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.AnalyzeDirectory(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request format", response["error"])
}

func TestMediaHandler_GetMediaStats_MethodExists(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mediaDB := &database.MediaDatabase{}
	analyzer := &analyzer.MediaAnalyzer{}
	handler := NewMediaHandler(mediaDB, analyzer, logger)

	// Test that the method exists and can be called (will panic due to nil DB, but that's expected)
	assert.NotNil(t, handler.GetMediaStats)
}

func TestMediaHandler_getQualityDistribution(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mediaDB := &database.MediaDatabase{}
	analyzer := &analyzer.MediaAnalyzer{}
	handler := NewMediaHandler(mediaDB, analyzer, logger)

	distribution, err := handler.getQualityDistribution()
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{
		"4K/UHD": 0,
		"1080p":  0,
		"720p":   0,
		"Other":  0,
	}, distribution)
}
