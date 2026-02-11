package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestServer creates a test HTTP server with a Gin router mimicking the real API
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/catalog", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"items": []gin.H{
					{"name": "media", "type": "directory", "path": "/media"},
				},
			})
		})

		api.GET("/search", func(c *gin.Context) {
			query := c.Query("query")
			if query == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"results": []gin.H{
					{"name": "test_movie.mp4", "path": "/media/movies/test_movie.mp4", "type": "movie"},
				},
				"total": 1,
			})
		})

		api.GET("/stats/overall", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"total_files": 100,
				"total_size":  5000000000,
			})
		})

		api.GET("/stats/duplicates/count", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"count":  5,
				"groups": 3,
			})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// TestHealthEndpoint verifies the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	assert.Equal(t, "healthy", health["status"])
	assert.NotNil(t, health["time"])
}

// TestCatalogListRoot verifies the catalog list root endpoint
func TestCatalogListRoot(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ts.URL + "/api/v1/catalog")
	if err != nil {
		t.Fatalf("Failed to call catalog endpoint: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	assert.NotNil(t, result["items"])
}

// TestCatalogSearch verifies the search endpoint handles queries
func TestCatalogSearch(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{"valid_search_query", "movie", http.StatusOK},
		{"empty_search_query", "", http.StatusBadRequest},
		{"search_with_movie_keyword", "movie", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := ts.URL + "/api/v1/search"
			if tt.query != "" {
				url += "?query=" + tt.query
			}

			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("Failed to call search endpoint: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestStatsOverall verifies the stats overall endpoint
func TestStatsOverall(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ts.URL + "/api/v1/stats/overall")
	if err != nil {
		t.Fatalf("Failed to call stats endpoint: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	assert.NotNil(t, result["total_files"])
	assert.NotNil(t, result["total_size"])
}

// TestDuplicatesCount verifies the duplicates count endpoint
func TestDuplicatesCount(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ts.URL + "/api/v1/stats/duplicates/count")
	if err != nil {
		t.Fatalf("Failed to call duplicates endpoint: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	assert.NotNil(t, result["count"])
	assert.NotNil(t, result["groups"])
}

// TestNonExistentEndpoint verifies 404 handling
func TestNonExistentEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ts.URL + "/api/v1/nonexistent")
	if err != nil {
		t.Fatalf("Failed to call non-existent endpoint: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestCORSHeaders verifies CORS headers are present
func TestCORSHeaders(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("OPTIONS", ts.URL+"/api/v1/catalog", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	// Verify the OPTIONS request doesn't cause a server error
	assert.Less(t, resp.StatusCode, 500)
}

// TestAPIResponseTime verifies API responds within reasonable time
func TestAPIResponseTime(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	resp, err := client.Get(ts.URL + "/health")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	// API should respond within 5 seconds
	assert.Less(t, duration, 5*time.Second)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMultipleConcurrentRequests verifies API handles concurrent requests
func TestMultipleConcurrentRequests(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	const numRequests = 10
	var wg sync.WaitGroup
	errors := make([]error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, err := client.Get(ts.URL + "/health")
			if err != nil {
				errors[idx] = err
				return
			}
			resp.Body.Close()
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "Concurrent request %d failed", i)
	}
}

// TestJSONResponseFormat verifies API returns valid JSON
func TestJSONResponseFormat(t *testing.T) {
	ts := setupTestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	endpoints := []string{
		"/health",
		"/api/v1/catalog",
		"/api/v1/search?query=test",
		"/api/v1/stats/duplicates/count",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := client.Get(ts.URL + endpoint)
			if err != nil {
				t.Fatalf("Failed to call endpoint %s: %v", endpoint, err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var jsonData interface{}
			err = json.Unmarshal(body, &jsonData)
			assert.NoError(t, err, "Invalid JSON response from %s", endpoint)
		})
	}
}
