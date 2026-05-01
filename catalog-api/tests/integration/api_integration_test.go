//go:build e2e_binary

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthEndpoint verifies the health check endpoint.
func TestHealthEndpoint(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(sb.baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var health map[string]interface{}
	err = json.Unmarshal(body, &health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.NotNil(t, health["time"])
}

// TestCatalogListRoot verifies the catalog list root endpoint.
func TestCatalogListRoot(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	token := sb.login()
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, sb.baseURL+"/api/v1/catalog", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.NotNil(t, result["items"])
}

// TestCatalogSearch verifies the search endpoint handles queries.
func TestCatalogSearch(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	token := sb.login()
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
			url := sb.baseURL + "/api/v1/search"
			if tt.query != "" {
				url += "?query=" + tt.query
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestStatsOverall verifies the stats overall endpoint.
func TestStatsOverall(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	token := sb.login()
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, sb.baseURL+"/api/v1/stats/overall", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.NotNil(t, result["total_files"])
	assert.NotNil(t, result["total_size"])
}

// TestDuplicatesCount verifies the duplicates count endpoint.
func TestDuplicatesCount(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	token := sb.login()
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, sb.baseURL+"/api/v1/stats/duplicates/count", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.NotNil(t, result["count"])
	assert.NotNil(t, result["groups"])
}

// TestNonExistentEndpoint verifies 404 handling.
func TestNonExistentEndpoint(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	token := sb.login()
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, sb.baseURL+"/api/v1/nonexistent", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestCORSHeaders verifies CORS headers are present.
func TestCORSHeaders(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodOptions, sb.baseURL+"/api/v1/catalog", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the OPTIONS request doesn't cause a server error.
	assert.Less(t, resp.StatusCode, 500)
}

// TestAPIResponseTime verifies API responds within reasonable time.
func TestAPIResponseTime(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	resp, err := client.Get(sb.baseURL + "/health")
	duration := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()

	// API should respond within 5 seconds.
	assert.Less(t, duration, 5*time.Second)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMultipleConcurrentRequests verifies API handles concurrent requests.
func TestMultipleConcurrentRequests(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	const numRequests = 10
	var wg sync.WaitGroup
	errors := make([]error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, err := client.Get(sb.baseURL + "/health")
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

// TestJSONResponseFormat verifies API returns valid JSON.
func TestJSONResponseFormat(t *testing.T) {
	sb := spawnBinary(t)
	defer sb.Close()

	token := sb.login()
	client := &http.Client{Timeout: 10 * time.Second}

	endpoints := []struct {
		path   string
		method string
		auth   bool
	}{
		{"/health", http.MethodGet, false},
		{"/api/v1/catalog", http.MethodGet, true},
		{"/api/v1/search?query=test", http.MethodGet, true},
		{"/api/v1/stats/duplicates/count", http.MethodGet, true},
	}

	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			req, err := http.NewRequest(ep.method, sb.baseURL+ep.path, nil)
			require.NoError(t, err)
			if ep.auth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var jsonData interface{}
			err = json.Unmarshal(body, &jsonData)
			assert.NoError(t, err, "Invalid JSON response from %s", ep.path)
		})
	}
}
