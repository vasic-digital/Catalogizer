package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// BrowsingConfig holds configuration for browsing challenges,
// loaded from environment variables with sensible defaults.
type BrowsingConfig struct {
	BaseURL   string // API base URL
	Username  string // Admin login username
	Password  string // Admin login password
	WebAppURL string // Web app URL for Playwright
	WebAppDir string // Path to catalog-web directory
}

// LoadBrowsingConfig reads browsing challenge configuration
// from environment variables, falling back to defaults.
func LoadBrowsingConfig() *BrowsingConfig {
	return &BrowsingConfig{
		BaseURL:   envOrDefault("BROWSING_API_URL", "http://localhost:8080"),
		Username:  envOrDefault("ADMIN_USERNAME", "admin"),
		Password:  envOrDefault("ADMIN_PASSWORD", "admin123"),
		WebAppURL: envOrDefault("BROWSING_WEB_URL", "http://localhost:3000"),
		WebAppDir: envOrDefault("CATALOG_WEB_DIR", "../catalog-web"),
	}
}

// envOrDefault returns the value of the named environment variable
// or the provided default if unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// APIClient wraps net/http.Client with JWT authentication support
// for calling the Catalogizer REST API.
type APIClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewAPIClient creates an API client targeting the given base URL.
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Login authenticates with the API and stores the JWT token
// for subsequent requests. Returns the parsed login response.
func (c *APIClient) Login(ctx context.Context, username, password string) (map[string]interface{}, error) {
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/auth/login",
		strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login returned HTTP %d: %s", resp.StatusCode, string(data))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse login response: %w", err)
	}

	// Extract and store token
	if token, ok := result["token"].(string); ok && token != "" {
		c.token = token
	}

	return result, nil
}

// Get performs an authenticated GET request and returns the
// status code and parsed JSON object response.
func (c *APIClient) Get(ctx context.Context, path string) (int, map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return resp.StatusCode, nil, fmt.Errorf("parse response: %w", err)
	}

	return resp.StatusCode, result, nil
}

// GetArray performs an authenticated GET request and returns the
// status code and parsed JSON array response.
func (c *APIClient) GetArray(ctx context.Context, path string) (int, []interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	var result []interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return resp.StatusCode, nil, fmt.Errorf("parse response: %w", err)
	}

	return resp.StatusCode, result, nil
}

// GetRaw performs an authenticated GET and returns status code
// and raw body bytes. Used when the response could be either
// an object or array.
func (c *APIClient) GetRaw(ctx context.Context, path string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	return resp.StatusCode, data, nil
}

// Token returns the stored JWT token.
func (c *APIClient) Token() string {
	return c.token
}

// invalidTitlePatterns lists title values that indicate missing
// or placeholder content in catalog entries.
var invalidTitlePatterns = []string{
	"unknown", "untitled", "placeholder", "n/a", "tbd", "",
}

// IsInvalidTitle checks whether a title matches any known
// invalid/placeholder pattern (case-insensitive).
func IsInvalidTitle(title string) bool {
	normalized := strings.TrimSpace(strings.ToLower(title))
	for _, pattern := range invalidTitlePatterns {
		if normalized == pattern {
			return true
		}
	}
	return false
}

// requiredMediaTypes lists the content categories that must
// be present in a fully cataloged collection.
var requiredMediaTypes = []string{
	"music", "tv_show", "movie", "software", "comic",
}
