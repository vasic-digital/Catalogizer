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

// setupSubtitleServer creates a test server with subtitle workflow endpoints
func setupSubtitleServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()

	var mu sync.Mutex
	tokens := map[string]bool{}
	nextSubID := 1
	subtitles := map[int]gin.H{}
	translationJobs := map[int]gin.H{}
	nextTranslationID := 1

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
				token := fmt.Sprintf("sub-token-%d", time.Now().UnixNano())
				mu.Lock()
				tokens[token] = true
				mu.Unlock()
				c.JSON(http.StatusOK, gin.H{"success": true, "token": token})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			}
		})

		// Subtitle search
		api.GET("/subtitles/search", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			mediaID := c.Query("media_id")
			language := c.DefaultQuery("language", "en")
			if mediaID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "media_id is required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"results": []gin.H{
					{"id": "opensubtitles-1", "provider": "opensubtitles", "language": language, "title": "Test Movie", "format": "srt", "rating": 8.5, "download_count": 15000},
					{"id": "subscene-1", "provider": "subscene", "language": language, "title": "Test Movie", "format": "srt", "rating": 7.2, "download_count": 8000},
					{"id": "opensubtitles-2", "provider": "opensubtitles", "language": language, "title": "Test Movie (Forced)", "format": "ass", "rating": 6.0, "download_count": 3000},
				},
				"total": 3,
			})
		})

		// Subtitle download
		api.POST("/subtitles/download", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			subtitleID, _ := data["subtitle_id"].(string)
			mediaID, _ := data["media_id"].(string)
			if subtitleID == "" || mediaID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "subtitle_id and media_id are required"})
				return
			}

			mu.Lock()
			id := nextSubID
			nextSubID++
			sub := gin.H{
				"id":          id,
				"subtitle_id": subtitleID,
				"media_id":    mediaID,
				"language":    "en",
				"format":      "srt",
				"path":        fmt.Sprintf("/media/subtitles/%s_%d.srt", mediaID, id),
				"status":      "downloaded",
				"created_at":  time.Now().UTC(),
			}
			subtitles[id] = sub
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "data": sub})
		})

		// List downloaded subtitles for a media item
		api.GET("/subtitles/media/:media_id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			mediaID := c.Param("media_id")
			results := []gin.H{}
			mu.Lock()
			for _, sub := range subtitles {
				if sub["media_id"] == mediaID {
					results = append(results, sub)
				}
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "data": results, "total": len(results)})
		})

		// Get subtitle content
		api.GET("/subtitles/:id/content", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			sub, exists := subtitles[id]
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Subtitle not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"id":       sub["id"],
					"format":   sub["format"],
					"language": sub["language"],
					"content":  "1\n00:00:01,000 --> 00:00:04,000\nThis is a test subtitle line.\n\n2\n00:00:05,000 --> 00:00:08,000\nSecond subtitle line for testing.\n",
				},
			})
		})

		// Set default subtitle
		api.PUT("/subtitles/:id/default", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			_, exists := subtitles[id]
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Subtitle not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Default subtitle set"})
		})

		// Delete subtitle
		api.DELETE("/subtitles/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			_, exists := subtitles[id]
			if exists {
				delete(subtitles, id)
			}
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Subtitle not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subtitle deleted"})
		})

		// Subtitle translation
		api.POST("/subtitles/:id/translate", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var subID int
			fmt.Sscanf(c.Param("id"), "%d", &subID)

			mu.Lock()
			_, exists := subtitles[subID]
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Subtitle not found"})
				return
			}

			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			targetLang, _ := data["target_language"].(string)
			if targetLang == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "target_language is required"})
				return
			}

			mu.Lock()
			jobID := nextTranslationID
			nextTranslationID++
			job := gin.H{
				"id":              jobID,
				"subtitle_id":     subID,
				"source_language": "en",
				"target_language": targetLang,
				"status":          "completed",
				"progress":        100,
				"created_at":      time.Now().UTC(),
				"completed_at":    time.Now().UTC(),
			}
			translationJobs[jobID] = job

			// Create translated subtitle
			newSubID := nextSubID
			nextSubID++
			subtitles[newSubID] = gin.H{
				"id":          newSubID,
				"subtitle_id": fmt.Sprintf("translated-%d", newSubID),
				"media_id":    subtitles[subID]["media_id"],
				"language":    targetLang,
				"format":      "srt",
				"path":        fmt.Sprintf("/media/subtitles/translated_%d.srt", newSubID),
				"status":      "translated",
				"created_at":  time.Now().UTC(),
			}
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{
				"success":             true,
				"translation_job":     job,
				"translated_subtitle": gin.H{"id": newSubID, "language": targetLang},
			})
		})

		// Get translation status
		api.GET("/subtitles/translations/:job_id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var jobID int
			fmt.Sscanf(c.Param("job_id"), "%d", &jobID)

			mu.Lock()
			job, exists := translationJobs[jobID]
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Translation job not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
		})

		// Supported languages
		api.GET("/subtitles/languages", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"languages": []gin.H{
					{"code": "en", "name": "English"},
					{"code": "es", "name": "Spanish"},
					{"code": "fr", "name": "French"},
					{"code": "de", "name": "German"},
					{"code": "ja", "name": "Japanese"},
					{"code": "zh", "name": "Chinese"},
					{"code": "sr", "name": "Serbian"},
				},
			})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// =============================================================================
// E2E TEST: Subtitle Search, Download, and Translate Workflow
// =============================================================================

func TestSubtitle_FullWorkflow(t *testing.T) {
	ts := setupSubtitleServer(t)
	ec := newE2EContext(ts.URL)

	// Login
	resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	result := ec.parseJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	ec.AuthToken = result["token"].(string)

	var downloadedSubID int
	var translationJobID int

	// Step 1: Search for subtitles
	t.Run("Step1_SearchSubtitles", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/subtitles/search?media_id=1&language=en", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		results := result["results"].([]interface{})
		assert.Equal(t, 3, len(results))

		first := results[0].(map[string]interface{})
		assert.Equal(t, "opensubtitles", first["provider"])
		assert.Equal(t, "en", first["language"])
		assert.NotEmpty(t, first["rating"])
	})

	// Step 2: Search without media_id (expect error)
	t.Run("Step2_SearchWithoutMediaID", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/subtitles/search?language=en", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Step 3: Download a subtitle
	t.Run("Step3_DownloadSubtitle", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/subtitles/download", map[string]interface{}{
			"subtitle_id": "opensubtitles-1",
			"media_id":    "1",
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		downloadedSubID = int(data["id"].(float64))
		assert.Equal(t, "downloaded", data["status"])
		assert.NotEmpty(t, data["path"])
	})

	// Step 4: List subtitles for media item
	t.Run("Step4_ListSubtitlesForMedia", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/subtitles/media/1", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 1)
	})

	// Step 5: Get subtitle content
	t.Run("Step5_GetSubtitleContent", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", fmt.Sprintf("/subtitles/%d/content", downloadedSubID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		assert.Equal(t, "srt", data["format"])
		assert.NotEmpty(t, data["content"])
	})

	// Step 6: Set as default subtitle
	t.Run("Step6_SetDefaultSubtitle", func(t *testing.T) {
		resp := ec.doRequest(t, "PUT", fmt.Sprintf("/subtitles/%d/default", downloadedSubID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	// Step 7: Translate subtitle to Spanish
	t.Run("Step7_TranslateSubtitle", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", fmt.Sprintf("/subtitles/%d/translate", downloadedSubID), map[string]interface{}{
			"target_language": "es",
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		job := result["translation_job"].(map[string]interface{})
		translationJobID = int(job["id"].(float64))
		assert.Equal(t, "completed", job["status"])
		assert.Equal(t, float64(100), job["progress"])

		translated := result["translated_subtitle"].(map[string]interface{})
		assert.Equal(t, "es", translated["language"])
	})

	// Step 8: Check translation status
	t.Run("Step8_CheckTranslationStatus", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", fmt.Sprintf("/subtitles/translations/%d", translationJobID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		assert.Equal(t, "completed", data["status"])
	})

	// Step 9: Get supported languages
	t.Run("Step9_SupportedLanguages", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/subtitles/languages", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		langs := result["languages"].([]interface{})
		assert.GreaterOrEqual(t, len(langs), 5)
	})

	// Step 10: Delete subtitle
	t.Run("Step10_DeleteSubtitle", func(t *testing.T) {
		resp := ec.doRequest(t, "DELETE", fmt.Sprintf("/subtitles/%d", downloadedSubID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	// Step 11: Verify deletion
	t.Run("Step11_VerifyDeletion", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", fmt.Sprintf("/subtitles/%d/content", downloadedSubID), nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Step 12: Translate non-existent subtitle
	t.Run("Step12_TranslateNonExistent", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/subtitles/9999/translate", map[string]interface{}{
			"target_language": "fr",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Step 13: Translate without target language
	t.Run("Step13_TranslateMissingLanguage", func(t *testing.T) {
		// Download another subtitle first
		resp := ec.doRequest(t, "POST", "/subtitles/download", map[string]interface{}{
			"subtitle_id": "opensubtitles-2",
			"media_id":    "1",
		})
		var dlResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&dlResult)
		resp.Body.Close()
		newID := int(dlResult["data"].(map[string]interface{})["id"].(float64))

		resp = ec.doRequest(t, "POST", fmt.Sprintf("/subtitles/%d/translate", newID), map[string]interface{}{})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
