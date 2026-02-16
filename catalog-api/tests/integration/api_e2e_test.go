package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupE2EServer creates a comprehensive test server for full API workflow E2E testing.
// It simulates auth, browse, search, media, download, subtitles, and conversion endpoints.
func setupE2EServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.HandleMethodNotAllowed = true

	// Shared state
	var mu sync.Mutex
	tokens := map[string]string{}       // token -> username
	users := map[string]map[string]interface{}{
		"admin": {
			"id":       1,
			"username": "admin",
			"password": "admin123",
			"role":     "admin",
			"email":    "admin@catalogizer.io",
		},
		"viewer": {
			"id":       2,
			"username": "viewer",
			"password": "viewer123",
			"role":     "viewer",
			"email":    "viewer@catalogizer.io",
		},
	}
	mediaItems := map[int]gin.H{
		1: {"id": 1, "title": "Inception", "type": "movie", "path": "/media/movies/inception.mkv", "size": 4500000000, "year": 2010, "genre": "sci-fi"},
		2: {"id": 2, "title": "Dark Side of the Moon", "type": "music", "path": "/media/music/pink_floyd/dark_side.flac", "size": 350000000, "year": 1973, "genre": "rock"},
		3: {"id": 3, "title": "Breaking Bad S01E01", "type": "series", "path": "/media/series/breaking_bad/s01e01.mkv", "size": 1200000000, "year": 2008, "genre": "drama"},
	}
	nextUserID := 3

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "version": "3.0.0", "time": time.Now().UTC()})
	})

	// Auth middleware helper
	checkAuth := func(c *gin.Context) (string, bool) {
		auth := c.GetHeader("Authorization")
		if auth == "" || len(auth) < 8 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return "", false
		}
		token := auth[7:]
		mu.Lock()
		username, valid := tokens[token]
		mu.Unlock()
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return "", false
		}
		return username, true
	}

	api := router.Group("/api/v1")
	{
		// ---- Auth endpoints ----
		api.POST("/auth/login", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
				return
			}
			username, _ := data["username"].(string)
			password, _ := data["password"].(string)

			mu.Lock()
			user, exists := users[username]
			mu.Unlock()
			if !exists || user["password"] != password {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			token := fmt.Sprintf("e2e-token-%s-%d", username, time.Now().UnixNano())
			mu.Lock()
			tokens[token] = username
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"token":   token,
				"user":    gin.H{"id": user["id"], "username": user["username"], "role": user["role"]},
			})
		})

		api.POST("/auth/register", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
				return
			}
			username, _ := data["username"].(string)
			if username == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username required"})
				return
			}
			mu.Lock()
			if _, exists := users[username]; exists {
				mu.Unlock()
				c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
				return
			}
			users[username] = map[string]interface{}{
				"id":       nextUserID,
				"username": username,
				"password": data["password"],
				"email":    data["email"],
				"role":     "user",
			}
			nextUserID++
			mu.Unlock()
			c.JSON(http.StatusCreated, gin.H{"success": true, "user": gin.H{"username": username, "email": data["email"]}})
		})

		api.POST("/auth/logout", func(c *gin.Context) {
			auth := c.GetHeader("Authorization")
			if auth != "" && len(auth) > 7 {
				token := auth[7:]
				mu.Lock()
				delete(tokens, token)
				mu.Unlock()
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		api.POST("/auth/refresh", func(c *gin.Context) {
			username, ok := checkAuth(c)
			if !ok {
				return
			}
			newToken := fmt.Sprintf("e2e-refreshed-%s-%d", username, time.Now().UnixNano())
			mu.Lock()
			tokens[newToken] = username
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "token": newToken})
		})

		// ---- User profile ----
		api.GET("/users/me", func(c *gin.Context) {
			username, ok := checkAuth(c)
			if !ok {
				return
			}
			mu.Lock()
			user := users[username]
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": user["id"], "username": user["username"], "role": user["role"], "email": user["email"]}})
		})

		// ---- Storage / Browse ----
		api.GET("/storage/roots", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"roots": []gin.H{
					{"id": "local-1", "name": "Movies", "protocol": "local", "path": "/media/movies", "enabled": true},
					{"id": "smb-1", "name": "NAS Music", "protocol": "smb", "path": "//nas/music", "enabled": true},
				},
			})
		})

		api.GET("/storage/browse", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			rootID := c.Query("root_id")
			path := c.DefaultQuery("path", "/")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"root_id": rootID,
				"path":    path,
				"entries": []gin.H{
					{"name": "inception.mkv", "type": "file", "size": 4500000000, "modified": time.Now().UTC()},
					{"name": "extras", "type": "directory", "size": 0, "modified": time.Now().UTC()},
				},
			})
		})

		// ---- Search ----
		api.GET("/search", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			query := c.Query("q")
			if query == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Search query 'q' is required"})
				return
			}
			mediaType := c.DefaultQuery("type", "")
			results := []gin.H{}
			mu.Lock()
			for _, item := range mediaItems {
				if mediaType != "" && item["type"] != mediaType {
					continue
				}
				results = append(results, item)
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "query": query, "results": results, "total": len(results)})
		})

		// ---- Media ----
		api.GET("/media", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			items := []gin.H{}
			mu.Lock()
			for _, item := range mediaItems {
				items = append(items, item)
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{
				"success":    true,
				"data":       items,
				"pagination": gin.H{"page": 1, "limit": 20, "total": len(items)},
			})
		})

		api.GET("/media/:id", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			id := c.Param("id")
			var found gin.H
			mu.Lock()
			for key, item := range mediaItems {
				if fmt.Sprintf("%d", key) == id {
					found = item
					break
				}
			}
			mu.Unlock()
			if found == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Media item not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": found})
		})

		// ---- Download ----
		api.GET("/download/:id", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			id := c.Param("id")
			mu.Lock()
			var found gin.H
			for key, item := range mediaItems {
				if fmt.Sprintf("%d", key) == id {
					found = item
					break
				}
			}
			mu.Unlock()
			if found == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
				return
			}
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", found["title"]))
			c.Header("Content-Type", "application/octet-stream")
			c.String(http.StatusOK, "mock-file-content-for-testing")
		})

		// ---- Analytics ----
		api.POST("/analytics/track", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "event_id": time.Now().UnixNano()})
		})

		api.GET("/analytics/dashboard", func(c *gin.Context) {
			if _, ok := checkAuth(c); !ok {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"total_media":    3,
					"total_size_gb":  5.8,
					"active_users":   2,
					"recent_scans":   1,
				},
			})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// E2EContext extends TestContext with additional E2E helper methods
type E2EContext struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
}

func newE2EContext(baseURL string) *E2EContext {
	return &E2EContext{
		BaseURL:    baseURL + "/api/v1",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (ec *E2EContext) doRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(method, ec.BaseURL+path, bodyReader)
	require.NoError(t, err)
	if ec.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+ec.AuthToken)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ec.HTTPClient.Do(req)
	require.NoError(t, err)
	return resp
}

func (ec *E2EContext) parseJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

// =============================================================================
// E2E TEST: Full API Workflow (Auth -> Browse -> Search -> Media -> Download)
// =============================================================================

func TestE2E_FullAPIWorkflow(t *testing.T) {
	ts := setupE2EServer(t)
	ec := newE2EContext(ts.URL)

	// Step 1: Health check
	t.Run("Step1_HealthCheck", func(t *testing.T) {
		resp, err := ec.HTTPClient.Get(ts.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "healthy", result["status"])
		assert.NotEmpty(t, result["version"])
	})

	// Step 2: Attempt protected access without auth (expect 401)
	t.Run("Step2_UnauthenticatedAccessDenied", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/media", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Step 3: Login
	t.Run("Step3_Login", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
			"username": "admin",
			"password": "admin123",
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		assert.NotEmpty(t, result["token"])
		ec.AuthToken = result["token"].(string)

		user := result["user"].(map[string]interface{})
		assert.Equal(t, "admin", user["username"])
		assert.Equal(t, "admin", user["role"])
	})

	// Step 4: Get user profile
	t.Run("Step4_GetProfile", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/users/me", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		assert.Equal(t, "admin", data["username"])
		assert.Equal(t, "admin@catalogizer.io", data["email"])
	})

	// Step 5: Browse storage roots
	t.Run("Step5_BrowseStorageRoots", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/storage/roots", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		roots := result["roots"].([]interface{})
		assert.GreaterOrEqual(t, len(roots), 1)
	})

	// Step 6: Browse a storage path
	t.Run("Step6_BrowseStoragePath", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/storage/browse?root_id=local-1&path=/", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		entries := result["entries"].([]interface{})
		assert.GreaterOrEqual(t, len(entries), 1)
	})

	// Step 7: Search for media
	t.Run("Step7_SearchMedia", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/search?q=inception", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		assert.Equal(t, "inception", result["query"])

		results := result["results"].([]interface{})
		assert.GreaterOrEqual(t, len(results), 1)
	})

	// Step 8: Search with empty query (expect error)
	t.Run("Step8_SearchEmptyQuery", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/search?q=", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Step 9: List media library
	t.Run("Step9_ListMediaLibrary", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/media", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 3)

		pagination := result["pagination"].(map[string]interface{})
		assert.NotNil(t, pagination["total"])
	})

	// Step 10: Get specific media item
	t.Run("Step10_GetMediaItem", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/media/1", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		assert.Equal(t, "Inception", data["title"])
		assert.Equal(t, "movie", data["type"])
	})

	// Step 11: Get non-existent media item (expect 404)
	t.Run("Step11_GetMediaItemNotFound", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/media/9999", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Step 12: Download a file
	t.Run("Step12_DownloadFile", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/download/1", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
		assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
	})

	// Step 13: Track analytics event
	t.Run("Step13_TrackAnalyticsEvent", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/analytics/track", map[string]interface{}{
			"event_type":  "media_view",
			"entity_type": "media_item",
			"entity_id":   1,
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	// Step 14: View analytics dashboard
	t.Run("Step14_AnalyticsDashboard", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/analytics/dashboard", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		assert.NotNil(t, data["total_media"])
	})

	// Step 15: Refresh token
	t.Run("Step15_RefreshToken", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/auth/refresh", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		assert.NotEmpty(t, result["token"])
	})

	// Step 16: Logout
	t.Run("Step16_Logout", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/auth/logout", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})
}

// =============================================================================
// E2E TEST: Multiple Users Concurrent Workflow
// =============================================================================

func TestE2E_MultiUserConcurrentWorkflow(t *testing.T) {
	ts := setupE2EServer(t)

	// Two users login and browse concurrently
	var wg sync.WaitGroup
	errors := make([]error, 2)

	users := []struct {
		username string
		password string
	}{
		{"admin", "admin123"},
		{"viewer", "viewer123"},
	}

	for i, u := range users {
		wg.Add(1)
		go func(idx int, username, password string) {
			defer wg.Done()

			ec := newE2EContext(ts.URL)

			// Login
			resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
				"username": username,
				"password": password,
			})
			var loginResult map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&loginResult)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors[idx] = fmt.Errorf("login failed for %s: status %d", username, resp.StatusCode)
				return
			}
			ec.AuthToken = loginResult["token"].(string)

			// Browse media
			resp = ec.doRequest(t, "GET", "/media", nil)
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errors[idx] = fmt.Errorf("media list failed for %s: status %d", username, resp.StatusCode)
				return
			}

			// Search
			resp = ec.doRequest(t, "GET", "/search?q=test", nil)
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errors[idx] = fmt.Errorf("search failed for %s: status %d", username, resp.StatusCode)
				return
			}

			// View media item
			resp = ec.doRequest(t, "GET", "/media/1", nil)
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errors[idx] = fmt.Errorf("media detail failed for %s: status %d", username, resp.StatusCode)
				return
			}
		}(i, u.username, u.password)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "User %d workflow failed", i)
	}
}

// =============================================================================
// E2E TEST: Registration and First Login Flow
// =============================================================================

func TestE2E_RegistrationAndFirstLogin(t *testing.T) {
	ts := setupE2EServer(t)
	ec := newE2EContext(ts.URL)

	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("newuser_%d", timestamp)
	email := fmt.Sprintf("newuser_%d@test.com", timestamp)
	password := "SecurePass123!"

	// Register
	t.Run("Register", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/auth/register", map[string]interface{}{
			"username": username,
			"email":    email,
			"password": password,
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	// Duplicate registration should fail
	t.Run("DuplicateRegistration", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/auth/register", map[string]interface{}{
			"username": username,
			"email":    email,
			"password": password,
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	// Login with new account
	t.Run("LoginNewAccount", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
			"username": username,
			"password": password,
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		ec.AuthToken = result["token"].(string)
	})

	// Access protected resource
	t.Run("AccessProtectedResource", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/users/me", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, username, data["username"])
	})
}

// =============================================================================
// E2E TEST: Search Filtering
// =============================================================================

func TestE2E_SearchWithFilters(t *testing.T) {
	ts := setupE2EServer(t)
	ec := newE2EContext(ts.URL)

	// Login first
	resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	resp.Body.Close()
	ec.AuthToken = loginResult["token"].(string)

	t.Run("SearchByType_Movie", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/search?q=all&type=movie", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		results := result["results"].([]interface{})
		for _, r := range results {
			item := r.(map[string]interface{})
			assert.Equal(t, "movie", item["type"])
		}
	})

	t.Run("SearchByType_Music", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/search?q=all&type=music", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		results := result["results"].([]interface{})
		for _, r := range results {
			item := r.(map[string]interface{})
			assert.Equal(t, "music", item["type"])
		}
	})

	t.Run("SearchAllTypes", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/search?q=all", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		total := int(result["total"].(float64))
		assert.Equal(t, 3, total)
	})
}

// =============================================================================
// E2E TEST: Download Workflow
// =============================================================================

func TestE2E_DownloadWorkflow(t *testing.T) {
	ts := setupE2EServer(t)
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

	t.Run("DownloadExistingFile", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/download/1", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
	})

	t.Run("DownloadNonExistentFile", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/download/9999", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("DownloadWithoutAuth", func(t *testing.T) {
		savedToken := ec.AuthToken
		ec.AuthToken = ""
		resp := ec.doRequest(t, "GET", "/download/1", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		ec.AuthToken = savedToken
	})
}
