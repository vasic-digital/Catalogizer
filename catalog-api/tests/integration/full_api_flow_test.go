package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupFullAPITestServer creates a comprehensive test HTTP server that mimics
// the full set of API endpoints exercised by the flow test. All endpoints
// return valid JSON with the expected field structure.
func setupFullAPITestServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// --- Health check ---
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"time":         time.Now().UTC(),
			"version":      "test",
			"build_number": "0",
			"build_date":   "unknown",
		})
	})

	// --- Auth ---
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		username, _ := body["username"].(string)
		password, _ := body["password"].(string)
		if username == "admin" && password == "admin123" {
			c.JSON(http.StatusOK, gin.H{
				"token": "test-jwt-token-abc123",
				"user": gin.H{
					"id":       1,
					"username": "admin",
					"role":     "admin",
				},
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		}
	})

	router.GET("/api/v1/auth/me", func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":       1,
			"username": "admin",
			"role":     "admin",
		})
	})

	// --- Storage roots ---
	router.GET("/api/v1/storage-roots", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"roots": []gin.H{
				{
					"id":       1,
					"name":     "Media NAS",
					"path":     "/media/nas",
					"protocol": "smb",
					"enabled":  true,
				},
				{
					"id":       2,
					"name":     "Local Media",
					"path":     "/media/local",
					"protocol": "local",
					"enabled":  true,
				},
			},
		})
	})

	// --- Media endpoints ---
	router.GET("/api/v1/media/recent", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items": []gin.H{
				{
					"id":         1,
					"title":      "Recent Movie",
					"type":       "movie",
					"path":       "/media/movies/recent.mkv",
					"created_at": time.Now().Add(-24 * time.Hour).UTC(),
				},
			},
			"total": 1,
		})
	})

	router.GET("/api/v1/media/popular", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items": []gin.H{
				{
					"id":         2,
					"title":      "Popular Series",
					"type":       "series",
					"path":       "/media/tv/popular_s01e01.mkv",
					"play_count": 42,
				},
			},
			"total": 1,
		})
	})

	router.GET("/api/v1/media/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"id":    3,
					"title": fmt.Sprintf("Search Result: %s", query),
					"type":  "movie",
					"path":  "/media/movies/result.mkv",
				},
			},
			"total": 1,
			"query": query,
		})
	})

	router.GET("/api/v1/media/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"total_files":  1500,
			"total_size":   75000000000,
			"media_count":  800,
			"movie_count":  350,
			"series_count": 150,
			"music_count":  300,
			"scan_status":  "idle",
			"last_scan_at": time.Now().Add(-1 * time.Hour).UTC(),
		})
	})

	// --- Entities ---
	router.GET("/api/v1/entities", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items": []gin.H{
				{"id": 1, "title": "Inception", "type": "movie", "year": 2010},
				{"id": 2, "title": "Breaking Bad", "type": "tv_show", "year": 2008},
			},
			"total": 2,
			"page":  1,
		})
	})

	// --- Admin: backups ---
	router.GET("/api/v1/backups", func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"backups": []gin.H{
				{
					"id":         1,
					"filename":   "backup_2026-03-30.db",
					"size":       52428800,
					"created_at": time.Now().Add(-24 * time.Hour).UTC(),
				},
			},
			"total": 1,
		})
	})

	// --- Configuration ---
	router.GET("/api/v1/configuration", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":  "test",
			"database": "sqlite",
			"features": gin.H{
				"http3":     true,
				"brotli":    true,
				"websocket": true,
			},
		})
	})

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// TestFullAPIFlow exercises the full API flow end-to-end as a table-driven
// test with subtests. Each subtest targets a specific API operation and
// verifies the response status, content type, JSON validity, and expected
// fields.
func TestFullAPIFlow(t *testing.T) {
	ts := setupFullAPITestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	// --- Step 1: Health check (no auth required) ---
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/api/v1/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertJSONContentType(t, resp)

		body := readBody(t, resp)
		data := parseJSON(t, body)

		assert.Equal(t, "healthy", data["status"])
		assert.NotNil(t, data["time"], "health response should include time")
		assert.NotNil(t, data["version"], "health response should include version")
	})

	// --- Step 2: Login ---
	var authToken string
	t.Run("Login", func(t *testing.T) {
		loginBody := strings.NewReader(`{"username":"admin","password":"admin123"}`)
		resp, err := client.Post(ts.URL+"/api/v1/auth/login", "application/json", loginBody)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertJSONContentType(t, resp)

		body := readBody(t, resp)
		data := parseJSON(t, body)

		token, ok := data["token"].(string)
		require.True(t, ok, "login response should contain a token string")
		assert.NotEmpty(t, token, "token should not be empty")
		authToken = token

		user, ok := data["user"].(map[string]interface{})
		assert.True(t, ok, "login response should contain a user object")
		assert.NotNil(t, user["username"], "user should have a username")
	})

	// Build auth headers for subsequent requests
	authHeaders := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")
	}

	// --- Steps 3-8: Table-driven authenticated endpoint tests ---
	type endpointTest struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedFields []string
	}

	tests := []endpointTest{
		{
			name:           "ListStorageRoots",
			method:         "GET",
			path:           "/api/v1/storage-roots",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"roots"},
		},
		{
			name:           "BrowseRecentMedia",
			method:         "GET",
			path:           "/api/v1/media/recent",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"items", "total"},
		},
		{
			name:           "BrowsePopularMedia",
			method:         "GET",
			path:           "/api/v1/media/popular",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"items", "total"},
		},
		{
			name:           "SearchMedia",
			method:         "GET",
			path:           "/api/v1/media/search?q=movie",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"results", "total", "query"},
		},
		{
			name:           "GetEntityStats",
			method:         "GET",
			path:           "/api/v1/media/stats",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"total_files", "total_size", "media_count"},
		},
		{
			name:           "ListEntities",
			method:         "GET",
			path:           "/api/v1/entities",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"items", "total"},
		},
		{
			name:           "ListBackups_Admin",
			method:         "GET",
			path:           "/api/v1/backups",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"backups"},
		},
		{
			name:           "GetConfiguration",
			method:         "GET",
			path:           "/api/v1/configuration",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"version"},
		},
		{
			name:           "GetAuthMe",
			method:         "GET",
			path:           "/api/v1/auth/me",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "username", "role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			require.NoError(t, err)
			authHeaders(req)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"endpoint %s should return status %d", tt.path, tt.expectedStatus)
			assertJSONContentType(t, resp)

			body := readBody(t, resp)
			data := parseJSON(t, body)

			for _, field := range tt.expectedFields {
				assert.NotNil(t, data[field],
					"response from %s should contain field %q", tt.path, field)
			}
		})
	}
}

// TestFullAPIFlow_InvalidAuth verifies that protected endpoints reject
// unauthenticated requests.
func TestFullAPIFlow_InvalidAuth(t *testing.T) {
	ts := setupFullAPITestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	protectedEndpoints := []string{
		"/api/v1/auth/me",
		"/api/v1/backups",
	}

	for _, endpoint := range protectedEndpoints {
		t.Run("Unauthorized_"+endpoint, func(t *testing.T) {
			resp, err := client.Get(ts.URL + endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
				"endpoint %s should reject unauthenticated requests", endpoint)
			assertJSONContentType(t, resp)

			body := readBody(t, resp)
			data := parseJSON(t, body)
			assert.NotNil(t, data["error"], "error response should contain 'error' field")
		})
	}
}

// TestFullAPIFlow_InvalidLogin verifies that login rejects bad credentials.
func TestFullAPIFlow_InvalidLogin(t *testing.T) {
	ts := setupFullAPITestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"WrongPassword", "admin", "wrongpassword"},
		{"WrongUsername", "nonexistent", "admin123"},
		{"BothWrong", "nobody", "nothing"},
		{"EmptyCredentials", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginBody := fmt.Sprintf(`{"username":%q,"password":%q}`, tt.username, tt.password)
			resp, err := client.Post(
				ts.URL+"/api/v1/auth/login",
				"application/json",
				strings.NewReader(loginBody),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
				"login with invalid credentials should return 401")
			assertJSONContentType(t, resp)

			body := readBody(t, resp)
			data := parseJSON(t, body)
			assert.NotNil(t, data["error"], "error response should contain 'error' field")
		})
	}
}

// TestFullAPIFlow_SearchValidation verifies search endpoint input validation.
func TestFullAPIFlow_SearchValidation(t *testing.T) {
	ts := setupFullAPITestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("MissingQuery", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/api/v1/media/search")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"search without query should return 400")
	})

	t.Run("EmptyQuery", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/api/v1/media/search?q=")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"search with empty query should return 400")
	})

	t.Run("ValidQuery", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/api/v1/media/search?q=test")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body := readBody(t, resp)
		data := parseJSON(t, body)
		assert.Equal(t, "test", data["query"], "response should echo back the query")
	})
}

// TestFullAPIFlow_NonExistentEndpoint verifies 404 handling.
func TestFullAPIFlow_NonExistentEndpoint(t *testing.T) {
	ts := setupFullAPITestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ts.URL + "/api/v1/does-not-exist")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestFullAPIFlow_ResponseTimes verifies that all endpoints respond within
// a reasonable timeframe.
func TestFullAPIFlow_ResponseTimes(t *testing.T) {
	ts := setupFullAPITestServer(t)
	client := &http.Client{Timeout: 10 * time.Second}

	endpoints := []string{
		"/api/v1/health",
		"/api/v1/storage-roots",
		"/api/v1/media/recent",
		"/api/v1/media/popular",
		"/api/v1/media/search?q=test",
		"/api/v1/media/stats",
		"/api/v1/entities",
		"/api/v1/configuration",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			start := time.Now()
			resp, err := client.Get(ts.URL + endpoint)
			duration := time.Since(start)

			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Less(t, duration, 5*time.Second,
				"endpoint %s should respond within 5 seconds", endpoint)
			assert.Less(t, resp.StatusCode, 500,
				"endpoint %s should not return a server error", endpoint)
		})
	}
}

// --- Helper functions ---

// assertJSONContentType verifies the response Content-Type contains application/json.
func assertJSONContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json",
		"Content-Type should be application/json, got: %s", contentType)
}

// readBody reads and returns the full response body. Fails the test on error.
func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")
	return body
}

// parseJSON unmarshals a JSON byte slice into a map. Fails the test on error.
func parseJSON(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	require.NoError(t, err, "response body should be valid JSON: %s", string(body))
	return data
}
