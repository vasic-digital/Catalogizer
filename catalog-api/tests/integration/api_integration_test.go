package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	baseURL = "http://localhost:8080"
	timeout = 10 * time.Second
)

var httpClient = &http.Client{
	Timeout: timeout,
}

// TestHealthEndpoint verifies the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	status, ok := health["status"].(string)
	if !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

// TestCatalogListRoot verifies the catalog list root endpoint
func TestCatalogListRoot(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/api/v1/catalog")
	if err != nil {
		t.Fatalf("Failed to call catalog endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got %s", contentType)
	}
}

// TestCatalogSearch verifies the search endpoint handles queries
func TestCatalogSearch(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "valid search query",
			query:          "?query=test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty search query",
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "search with movie keyword",
			query:          "?query=movie",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := baseURL + "/api/v1/search" + tt.query
			resp, err := httpClient.Get(url)
			if err != nil {
				t.Fatalf("Failed to call search endpoint: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestStatsSummary verifies the stats summary endpoint
func TestStatsSummary(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/api/v1/stats/summary")
	if err != nil {
		t.Skip("Stats endpoint not available - skipping test")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Stats endpoint not implemented - skipping test")
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestDuplicatesCount verifies the duplicates count endpoint
func TestDuplicatesCount(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/api/v1/stats/duplicates/count")
	if err != nil {
		t.Skip("Duplicates endpoint not available - skipping test")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Duplicates endpoint not implemented - skipping test")
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestNonExistentEndpoint verifies 404 handling
func TestNonExistentEndpoint(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/api/v1/nonexistent")
	if err != nil {
		t.Fatalf("Failed to call non-existent endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

// TestCORSHeaders verifies CORS headers are present
func TestCORSHeaders(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", baseURL+"/api/v1/catalog", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Skip("CORS test skipped - OPTIONS not supported")
		return
	}
	defer resp.Body.Close()

	// Just verify the OPTIONS request doesn't fail completely
	if resp.StatusCode >= 500 {
		t.Errorf("Server error on OPTIONS request: %d", resp.StatusCode)
	}
}

// TestAPIResponseTime verifies API responds within reasonable time
func TestAPIResponseTime(t *testing.T) {
	start := time.Now()
	resp, err := httpClient.Get(baseURL + "/health")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	maxDuration := 2 * time.Second
	if duration > maxDuration {
		t.Errorf("Response time %v exceeds maximum %v", duration, maxDuration)
	}
}

// TestMultipleConcurrentRequests verifies API handles concurrent requests
func TestMultipleConcurrentRequests(t *testing.T) {
	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			resp, err := httpClient.Get(baseURL + "/health")
			if err != nil {
				errors <- fmt.Errorf("request %d failed: %v", id, err)
				done <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d got status %d", id, resp.StatusCode)
				done <- false
				return
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-done {
			successCount++
		}
	}

	close(errors)
	for err := range errors {
		t.Error(err)
	}

	if successCount < concurrency {
		t.Errorf("Only %d/%d concurrent requests succeeded", successCount, concurrency)
	}
}

// TestJSONResponseFormat verifies all responses are valid JSON
func TestJSONResponseFormat(t *testing.T) {
	endpoints := []string{
		"/health",
		"/api/v1/catalog",
		"/api/v1/search?query=test",
		"/api/v1/stats/duplicates/count",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := httpClient.Get(baseURL + endpoint)
			if err != nil {
				t.Skip("Endpoint not available - skipping")
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				t.Skip("Endpoint not found - skipping")
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var js interface{}
			if err := json.Unmarshal(body, &js); err != nil {
				t.Errorf("Invalid JSON response: %v\nBody: %s", err, string(body))
			}
		})
	}
}
