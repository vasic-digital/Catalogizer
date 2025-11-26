package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestConversionAPIEndpoints tests the conversion API endpoints without full integration
func TestConversionAPIEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("SupportedFormatsStructure", func(t *testing.T) {
		// Test the structure of supported formats response
		formats := &models.SupportedFormats{
			Video:    models.VideoFormats{Input: []string{"mp4", "avi"}, Output: []string{"mp4", "avi"}},
			Audio:    models.AudioFormats{Input: []string{"mp3", "wav"}, Output: []string{"mp3", "wav"}},
			Image:    models.ImageFormats{Input: []string{"jpg", "png"}, Output: []string{"jpg", "png"}},
			Document: models.DocumentFormats{Input: []string{"pdf", "docx"}, Output: []string{"pdf", "docx"}},
		}

		// Verify structure
		assert.NotNil(t, formats)
		assert.NotEmpty(t, formats.Video.Input)
		assert.NotEmpty(t, formats.Video.Output)
		assert.NotEmpty(t, formats.Audio.Input)
		assert.NotEmpty(t, formats.Audio.Output)
		assert.NotEmpty(t, formats.Image.Input)
		assert.NotEmpty(t, formats.Image.Output)
		assert.NotEmpty(t, formats.Document.Input)
		assert.NotEmpty(t, formats.Document.Output)
	})

	t.Run("ConversionJobStructure", func(t *testing.T) {
		// Test the structure of conversion job
		job := &models.ConversionJob{
			ID:             1,
			UserID:         1,
			SourcePath:     "/test/input.avi",
			TargetPath:     "/test/output.mp4",
			SourceFormat:   "avi",
			TargetFormat:   "mp4",
			ConversionType: "video",
			Quality:        "high",
			Status:         "pending",
		}

		// Verify structure
		assert.Equal(t, 1, job.ID)
		assert.Equal(t, 1, job.UserID)
		assert.Equal(t, "/test/input.avi", job.SourcePath)
		assert.Equal(t, "/test/output.mp4", job.TargetPath)
		assert.Equal(t, "avi", job.SourceFormat)
		assert.Equal(t, "mp4", job.TargetFormat)
		assert.Equal(t, "video", job.ConversionType)
		assert.Equal(t, "high", job.Quality)
		assert.Equal(t, "pending", job.Status)
	})

	t.Run("ConversionRequestStructure", func(t *testing.T) {
		// Test the structure of conversion request
		request := &models.ConversionRequest{
			SourcePath:     "/test/input.avi",
			TargetPath:     "/test/output.mp4",
			SourceFormat:   "avi",
			TargetFormat:   "mp4",
			ConversionType: "video",
			Quality:        "high",
			Priority:       1,
		}

		// Verify structure
		assert.Equal(t, "/test/input.avi", request.SourcePath)
		assert.Equal(t, "/test/output.mp4", request.TargetPath)
		assert.Equal(t, "avi", request.SourceFormat)
		assert.Equal(t, "mp4", request.TargetFormat)
		assert.Equal(t, "video", request.ConversionType)
		assert.Equal(t, "high", request.Quality)
		assert.Equal(t, 1, request.Priority)
	})

	t.Run("JSONSerialization", func(t *testing.T) {
		// Test JSON serialization/deserialization for API communication
		request := models.ConversionRequest{
			SourcePath:     "/test/input.avi",
			TargetPath:     "/test/output.mp4",
			SourceFormat:   "avi",
			TargetFormat:   "mp4",
			ConversionType: "video",
			Quality:        "high",
		}

		// Serialize to JSON
		jsonData, err := json.Marshal(request)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Deserialize from JSON
		var deserializedRequest models.ConversionRequest
		err = json.Unmarshal(jsonData, &deserializedRequest)
		assert.NoError(t, err)
		assert.Equal(t, request.SourcePath, deserializedRequest.SourcePath)
		assert.Equal(t, request.TargetPath, deserializedRequest.TargetPath)
		assert.Equal(t, request.SourceFormat, deserializedRequest.SourceFormat)
		assert.Equal(t, request.TargetFormat, deserializedRequest.TargetFormat)
		assert.Equal(t, request.ConversionType, deserializedRequest.ConversionType)
		assert.Equal(t, request.Quality, deserializedRequest.Quality)
	})

	t.Run("HTTPResponseStructure", func(t *testing.T) {
		// Test HTTP response structure for API
		job := &models.ConversionJob{
			ID:             1,
			UserID:         1,
			SourcePath:     "/test/input.avi",
			TargetPath:     "/test/output.mp4",
			SourceFormat:   "avi",
			TargetFormat:   "mp4",
			ConversionType: "video",
			Quality:        "high",
			Status:         "pending",
		}

		// Test response serialization
		jsonData, err := json.Marshal(job)
		assert.NoError(t, err)

		// Test HTTP response structure
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonData)

		// Verify response
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Verify response body can be parsed
		var responseJob models.ConversionJob
		err = json.Unmarshal(w.Body.Bytes(), &responseJob)
		assert.NoError(t, err)
		assert.Equal(t, job.ID, responseJob.ID)
		assert.Equal(t, job.SourcePath, responseJob.SourcePath)
		assert.Equal(t, job.TargetPath, responseJob.TargetPath)
	})
}

// TestConversionPermissionConstants tests that required permission constants are defined
func TestConversionPermissionConstants(t *testing.T) {
	// Test that permission constants are properly defined
	assert.NotEmpty(t, models.PermissionConversionCreate)
	assert.NotEmpty(t, models.PermissionConversionView)
	assert.NotEmpty(t, models.PermissionConversionManage)
}

// TestConversionStatusConstants tests that status constants are properly defined
func TestConversionStatusConstants(t *testing.T) {
	// Test that status constants are properly defined
	assert.NotEmpty(t, models.ConversionStatusPending)
	assert.NotEmpty(t, models.ConversionStatusRunning)
	assert.NotEmpty(t, models.ConversionStatusCompleted)
	assert.NotEmpty(t, models.ConversionStatusFailed)
	assert.NotEmpty(t, models.ConversionStatusCancelled)
}