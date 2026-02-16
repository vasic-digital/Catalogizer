package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupConversionServer creates a test server with media conversion endpoints
func setupConversionServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()

	var mu sync.Mutex
	tokens := map[string]bool{}
	nextJobID := 1
	jobs := map[int]gin.H{}
	presets := []gin.H{
		{"id": "mp4-h264-1080p", "name": "MP4 H.264 1080p", "format": "mp4", "video_codec": "h264", "resolution": "1920x1080", "bitrate": 8000},
		{"id": "mp4-h265-4k", "name": "MP4 H.265 4K", "format": "mp4", "video_codec": "hevc", "resolution": "3840x2160", "bitrate": 20000},
		{"id": "webm-vp9-720p", "name": "WebM VP9 720p", "format": "webm", "video_codec": "vp9", "resolution": "1280x720", "bitrate": 4000},
		{"id": "mp3-320", "name": "MP3 320kbps", "format": "mp3", "audio_codec": "mp3", "bitrate": 320},
		{"id": "flac-lossless", "name": "FLAC Lossless", "format": "flac", "audio_codec": "flac", "bitrate": 0},
	}

	checkAuth := func(c *gin.Context) bool {
		auth := c.GetHeader("Authorization")
		if auth == "" || len(auth) < 8 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return false
		}
		token := auth[7:]
		mu.Lock()
		valid := tokens[token]
		mu.Unlock()
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return false
		}
		return true
	}

	api := router.Group("/api/v1")
	{
		api.POST("/auth/login", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			username, _ := data["username"].(string)
			password, _ := data["password"].(string)
			if username == "admin" && password == "admin123" {
				token := fmt.Sprintf("conv-token-%d", time.Now().UnixNano())
				mu.Lock()
				tokens[token] = true
				mu.Unlock()
				c.JSON(http.StatusOK, gin.H{"success": true, "token": token})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			}
		})

		// Get conversion presets
		api.GET("/conversion/presets", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": presets})
		})

		// Get supported formats
		api.GET("/conversion/formats", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"video":   []string{"mp4", "mkv", "webm", "avi", "mov"},
				"audio":   []string{"mp3", "flac", "aac", "ogg", "wav"},
			})
		})

		// Create conversion job
		api.POST("/conversion/jobs", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			sourcePath, _ := data["source_path"].(string)
			targetFormat, _ := data["target_format"].(string)
			presetID, _ := data["preset_id"].(string)

			if sourcePath == "" || targetFormat == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "source_path and target_format are required"})
				return
			}

			mu.Lock()
			id := nextJobID
			nextJobID++
			job := gin.H{
				"id":            id,
				"source_path":   sourcePath,
				"target_format": targetFormat,
				"preset_id":     presetID,
				"status":        "pending",
				"progress":      0,
				"created_at":    time.Now().UTC(),
				"started_at":    nil,
				"completed_at":  nil,
				"error_message": nil,
				"output_path":   "",
				"output_size":   0,
			}
			jobs[id] = job
			mu.Unlock()

			c.JSON(http.StatusCreated, gin.H{"success": true, "data": job})
		})

		// List conversion jobs
		api.GET("/conversion/jobs", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			status := c.Query("status")
			mu.Lock()
			result := []gin.H{}
			for _, job := range jobs {
				if status != "" && job["status"] != status {
					continue
				}
				result = append(result, job)
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "total": len(result)})
		})

		// Get job status
		api.GET("/conversion/jobs/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			job, exists := jobs[id]
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
		})

		// Start job processing
		api.POST("/conversion/jobs/:id/start", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			job, exists := jobs[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			if job["status"] != "pending" {
				mu.Unlock()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Job is not in pending state"})
				return
			}
			now := time.Now().UTC()
			job["status"] = "processing"
			job["progress"] = 50
			job["started_at"] = now
			jobs[id] = job
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
		})

		// Complete job (simulate)
		api.POST("/conversion/jobs/:id/complete", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			job, exists := jobs[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			if job["status"] != "processing" {
				mu.Unlock()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Job is not in processing state"})
				return
			}
			now := time.Now().UTC()
			job["status"] = "completed"
			job["progress"] = 100
			job["completed_at"] = now
			job["output_path"] = fmt.Sprintf("/media/converted/output_%d.%s", id, job["target_format"])
			job["output_size"] = 2500000000
			jobs[id] = job
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
		})

		// Cancel job
		api.POST("/conversion/jobs/:id/cancel", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			job, exists := jobs[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			if job["status"] == "completed" || job["status"] == "cancelled" {
				mu.Unlock()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel completed or already cancelled job"})
				return
			}
			job["status"] = "cancelled"
			jobs[id] = job
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
		})

		// Retry failed job
		api.POST("/conversion/jobs/:id/retry", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			job, exists := jobs[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			if job["status"] != "failed" && job["status"] != "cancelled" {
				mu.Unlock()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Only failed or cancelled jobs can be retried"})
				return
			}
			job["status"] = "pending"
			job["progress"] = 0
			job["error_message"] = nil
			jobs[id] = job
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
		})

		// Delete job
		api.DELETE("/conversion/jobs/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			_, exists := jobs[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
				return
			}
			delete(jobs, id)
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Job deleted"})
		})

		// Conversion stats
		api.GET("/conversion/stats", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			mu.Lock()
			pending, processing, completed, failed, cancelled := 0, 0, 0, 0, 0
			for _, job := range jobs {
				switch job["status"] {
				case "pending":
					pending++
				case "processing":
					processing++
				case "completed":
					completed++
				case "failed":
					failed++
				case "cancelled":
					cancelled++
				}
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"total":      pending + processing + completed + failed + cancelled,
					"pending":    pending,
					"processing": processing,
					"completed":  completed,
					"failed":     failed,
					"cancelled":  cancelled,
				},
			})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// =============================================================================
// E2E TEST: Conversion Job Full Lifecycle
// =============================================================================

func TestConversion_FullLifecycle(t *testing.T) {
	ts := setupConversionServer(t)
	ec := newE2EContext(ts.URL)

	// Login
	resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	result := ec.parseJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	ec.AuthToken = result["token"].(string)

	var jobID int

	// Step 1: Get available presets
	t.Run("Step1_GetPresets", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/presets", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		presets := result["data"].([]interface{})
		assert.Equal(t, 5, len(presets))
	})

	// Step 2: Get supported formats
	t.Run("Step2_GetFormats", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/formats", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		video := result["video"].([]interface{})
		audio := result["audio"].([]interface{})
		assert.GreaterOrEqual(t, len(video), 3)
		assert.GreaterOrEqual(t, len(audio), 3)
	})

	// Step 3: Create conversion job
	t.Run("Step3_CreateJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/conversion/jobs", map[string]interface{}{
			"source_path":   "/media/movies/inception.mkv",
			"target_format": "mp4",
			"preset_id":     "mp4-h264-1080p",
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		jobID = int(data["id"].(float64))
		assert.Equal(t, "pending", data["status"])
		assert.Equal(t, float64(0), data["progress"])
		assert.Equal(t, "/media/movies/inception.mkv", data["source_path"])
	})

	// Step 4: Create job with missing fields
	t.Run("Step4_CreateJobMissingFields", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/conversion/jobs", map[string]interface{}{
			"source_path": "/media/test.mkv",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Step 5: Check initial job status
	t.Run("Step5_CheckJobStatus", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", fmt.Sprintf("/conversion/jobs/%d", jobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "pending", data["status"])
	})

	// Step 6: Start job processing
	t.Run("Step6_StartJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/start", jobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "processing", data["status"])
		assert.NotNil(t, data["started_at"])
	})

	// Step 7: Cannot start already processing job
	t.Run("Step7_CannotStartProcessingJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/start", jobID), nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Step 8: Complete the job
	t.Run("Step8_CompleteJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/complete", jobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "completed", data["status"])
		assert.Equal(t, float64(100), data["progress"])
		assert.NotEmpty(t, data["output_path"])
		assert.NotNil(t, data["completed_at"])
	})

	// Step 9: Cannot cancel completed job
	t.Run("Step9_CannotCancelCompleted", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/cancel", jobID), nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Step 10: Check conversion stats
	t.Run("Step10_CheckStats", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/stats", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["completed"])
	})

	// Step 11: List all jobs
	t.Run("Step11_ListJobs", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/jobs", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		total := int(result["total"].(float64))
		assert.GreaterOrEqual(t, total, 1)
	})

	// Step 12: List by status filter
	t.Run("Step12_ListJobsByStatus", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/jobs?status=completed", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].([]interface{})
		for _, j := range data {
			job := j.(map[string]interface{})
			assert.Equal(t, "completed", job["status"])
		}
	})

	// Step 13: Delete job
	t.Run("Step13_DeleteJob", func(t *testing.T) {
		resp := ec.doRequest(t, "DELETE", fmt.Sprintf("/conversion/jobs/%d", jobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	// Step 14: Verify deletion
	t.Run("Step14_VerifyDeletion", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", fmt.Sprintf("/conversion/jobs/%d", jobID), nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// E2E TEST: Conversion Job Cancel and Retry
// =============================================================================

func TestConversion_CancelAndRetry(t *testing.T) {
	ts := setupConversionServer(t)
	ec := newE2EContext(ts.URL)

	// Login
	resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	resp.Body.Close()
	ec.AuthToken = loginResult["token"].(string)

	// Create a job
	resp = ec.doRequest(t, "POST", "/conversion/jobs", map[string]interface{}{
		"source_path":   "/media/series/episode.mkv",
		"target_format": "webm",
		"preset_id":     "webm-vp9-720p",
	})
	var createResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResult)
	resp.Body.Close()
	jobID := int(createResult["data"].(map[string]interface{})["id"].(float64))

	t.Run("CancelPendingJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/cancel", jobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "cancelled", data["status"])
	})

	t.Run("RetryCancelledJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/retry", jobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "pending", data["status"])
		assert.Equal(t, float64(0), data["progress"])
	})

	t.Run("CannotRetryPendingJob", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/retry", jobID), nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// E2E TEST: Multiple Concurrent Conversion Jobs
// =============================================================================

func TestConversion_MultipleJobs(t *testing.T) {
	ts := setupConversionServer(t)
	ec := newE2EContext(ts.URL)

	// Login
	resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	resp.Body.Close()
	ec.AuthToken = loginResult["token"].(string)

	jobCount := 5
	jobIDs := make([]int, jobCount)

	// Create multiple jobs
	for i := 0; i < jobCount; i++ {
		resp := ec.doRequest(t, "POST", "/conversion/jobs", map[string]interface{}{
			"source_path":   fmt.Sprintf("/media/batch/file_%d.mkv", i),
			"target_format": "mp4",
			"preset_id":     "mp4-h264-1080p",
		})
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		jobIDs[i] = int(result["data"].(map[string]interface{})["id"].(float64))
	}

	t.Run("AllJobsCreated", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/jobs", nil)
		result := ec.parseJSON(t, resp)
		total := int(result["total"].(float64))
		assert.GreaterOrEqual(t, total, jobCount)
	})

	t.Run("ProcessAllJobs", func(t *testing.T) {
		for _, id := range jobIDs {
			// Start
			resp := ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/start", id), nil)
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Complete
			resp = ec.doRequest(t, "POST", fmt.Sprintf("/conversion/jobs/%d/complete", id), nil)
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("VerifyAllCompleted", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/conversion/stats", nil)
		result := ec.parseJSON(t, resp)
		data := result["data"].(map[string]interface{})
		assert.GreaterOrEqual(t, int(data["completed"].(float64)), jobCount)
	})
}
