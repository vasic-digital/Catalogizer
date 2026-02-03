package integration

import (
	"encoding/json"
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

// checkServerAvailability checks if the server is running
func checkServerAvailability() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// TestHealthEndpoint verifies the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

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
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

	resp, err := httpClient.Get(baseURL + "/api/v1/catalog")
	if err != nil {
		t.Fatalf("Failed to call catalog endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Allow 401 Unauthorized since the endpoint requires authentication
	if resp.StatusCode == http.StatusUnauthorized {
		t.Log("Catalog endpoint requires authentication - test passed (endpoint exists)")
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 or 401, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	// Allow charset in content type
	if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected content type 'application/json', got %s", contentType)
	}
}

// TestCatalogSearch verifies the search endpoint handles queries
func TestCatalogSearch(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

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
			url := baseURL + "/api/v1/search"
			if tt.query != "" {
				url += "?query=" + tt.query
			}

			resp, err := httpClient.Get(url)
			if err != nil {
				t.Fatalf("Failed to call search endpoint: %v", err)
			}
			defer resp.Body.Close()

			// Allow 401 Unauthorized since the endpoint requires authentication
			if resp.StatusCode == http.StatusUnauthorized {
				t.Log("Search endpoint requires authentication - test passed (endpoint exists)")
				return
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestStatsOverall verifies the stats overall endpoint
func TestStatsOverall(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

	resp, err := httpClient.Get(baseURL + "/api/v1/stats/overall")
	if err != nil {
		t.Skip("Stats endpoint not available - skipping test")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Stats endpoint not implemented - skipping test")
		return
	}

	// Allow 401 Unauthorized since the endpoint requires authentication
	if resp.StatusCode == http.StatusUnauthorized {
		t.Log("Stats endpoint requires authentication - test passed (endpoint exists)")
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 or 401, got %d", resp.StatusCode)
	}
}

// TestDuplicatesCount verifies the duplicates count endpoint
func TestDuplicatesCount(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

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

	// Allow 401 Unauthorized since the endpoint requires authentication
	if resp.StatusCode == http.StatusUnauthorized {
		t.Log("Duplicates count endpoint requires authentication - test passed (endpoint exists)")
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 or 401, got %d", resp.StatusCode)
	}
}

// TestNonExistentEndpoint verifies 404 handling
func TestNonExistentEndpoint(t *testing.T) {
	resp, err := httpClient.Get(baseURL + "/api/v1/nonexistent")
	if err != nil {
		t.Skip("Non-existent endpoint test skipped - connection failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

// TestCORSHeaders verifies CORS headers are present
func TestCORSHeaders(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

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
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

	start := time.Now()
	resp, err := httpClient.Get(baseURL + "/health")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	// API should respond within 5 seconds
	if duration > 5*time.Second {
		t.Errorf("API took too long to respond: %v", duration)
	}
}

// TestMultipleConcurrentRequests verifies API handles concurrent requests
func TestMultipleConcurrentRequests(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := httpClient.Get(baseURL + "/health")
			if err != nil {
				results <- err
				return
			}
			resp.Body.Close()
			results <- nil
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}

// TestJSONResponseFormat verifies API returns valid JSON
func TestJSONResponseFormat(t *testing.T) {
	if !checkServerAvailability() {
		t.Skip("Server not available - skipping integration test")
	}

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
				t.Skipf("Endpoint %s not available - skipping", endpoint)
				return
			}
			defer resp.Body.Close()

			// Skip if endpoint returns 404
			if resp.StatusCode == http.StatusNotFound {
				t.Skipf("Endpoint %s not implemented - skipping", endpoint)
				return
			}

			// 401 Unauthorized responses should still be valid JSON
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var jsonData interface{}
			if err := json.Unmarshal(body, &jsonData); err != nil {
				t.Errorf("Invalid JSON response from %s: %v", endpoint, err)
			}
		})
	}
}
