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

// TestContext holds common test data
type TestContext struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
	UserID     int
}

func newTestContext(baseURL string) *TestContext {
	return &TestContext{
		BaseURL: baseURL + "/api/v1",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Helper to make authenticated requests
func (tc *TestContext) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, tc.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if tc.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+tc.AuthToken)
	}
	req.Header.Set("Content-Type", "application/json")

	return tc.HTTPClient.Do(req)
}

// setupUserFlowsServer creates a test server with full API endpoints for user flow testing
func setupUserFlowsServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.HandleMethodNotAllowed = true

	// State for the mock server
	var mu sync.Mutex
	users := map[string]map[string]interface{}{
		"admin": {
			"id":       1,
			"username": "admin",
			"password": "admin123",
			"role":     "admin",
			"email":    "admin@example.com",
		},
	}
	tokens := map[string]string{} // token -> username
	collections := map[int]map[string]interface{}{}
	nextCollectionID := 1
	nextUserID := 2

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	api := router.Group("/api/v1")
	{
		// Auth endpoints
		api.POST("/auth/register", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			username, _ := data["username"].(string)
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

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"user": gin.H{
					"username": username,
					"email":    data["email"],
				},
			})
		})

		api.POST("/auth/login", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
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

			token := fmt.Sprintf("test-token-%s-%d", username, time.Now().UnixNano())
			mu.Lock()
			tokens[token] = username
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"token":   token,
				"user": gin.H{
					"id":       user["id"],
					"username": user["username"],
					"role":     user["role"],
				},
			})
		})

		api.POST("/auth/logout", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// Auth middleware helper
		checkAuth := func(c *gin.Context) bool {
			auth := c.GetHeader("Authorization")
			if auth == "" || len(auth) < 8 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return false
			}
			token := auth[7:] // Remove "Bearer "
			mu.Lock()
			_, valid := tokens[token]
			mu.Unlock()
			if !valid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return false
			}
			return true
		}

		// User endpoints
		api.GET("/users/me", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"id":       1,
					"username": "admin",
					"role":     "admin",
				},
			})
		})

		// Storage endpoints
		api.GET("/storage/roots", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"roots": []gin.H{
					{"id": "local", "name": "Local Files", "protocol": "local", "enabled": true},
				},
			})
		})

		api.GET("/storage/list/", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"files": []gin.H{
					{"name": "test.mp4", "type": "file", "size": 1024000},
				},
			})
		})

		// Media endpoints
		api.GET("/media", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": []gin.H{
					{"id": 1, "title": "Test Movie", "type": "video"},
				},
				"pagination": gin.H{"page": 1, "limit": 10, "total": 1},
			})
		})

		api.GET("/media/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"id":    1,
					"title": "Test Movie",
					"type":  "video",
				},
			})
		})

		api.PUT("/media/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// Analytics endpoints
		api.POST("/analytics/track", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		api.GET("/analytics/dashboard", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"total_views":    100,
					"active_users":   5,
					"total_media":    50,
					"storage_used_gb": 10.5,
				},
			})
		})

		api.GET("/analytics/events", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    []gin.H{},
			})
		})

		// Collections endpoints
		api.GET("/collections", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			mu.Lock()
			items := make([]gin.H, 0, len(collections))
			for id, col := range collections {
				items = append(items, gin.H{"id": id, "name": col["name"]})
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
		})

		api.POST("/collections", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			mu.Lock()
			id := nextCollectionID
			nextCollectionID++
			collections[id] = data
			mu.Unlock()

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"data": gin.H{
					"id":   id,
					"name": data["name"],
				},
			})
		})

		api.GET("/collections/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": 1, "name": "Test"}})
		})

		api.DELETE("/collections/:id", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// Favorites endpoints
		api.GET("/favorites", func(c *gin.Context) {
			if !checkAuth(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{}})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// =============================================================================
// INTEGRATION TEST: Complete Authentication Flow
// =============================================================================

func TestAuthenticationFlow(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	// Subtest: User Registration
	t.Run("UserRegistration", func(t *testing.T) {
		timestamp := time.Now().Unix()
		registerData := map[string]interface{}{
			"username":       fmt.Sprintf("testuser_%d", timestamp),
			"email":          fmt.Sprintf("test_%d@example.com", timestamp),
			"password":       "SecurePassword123!",
			"full_name":      "Test User",
			"terms_accepted": true,
		}

		resp, err := tc.makeRequest("POST", "/auth/register", registerData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Registration should succeed")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["user"])
	})

	// Subtest: User Login
	t.Run("UserLogin", func(t *testing.T) {
		loginData := map[string]interface{}{
			"username":    "admin",
			"password":    "admin123",
			"remember_me": false,
		}

		resp, err := tc.makeRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Login should succeed")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotEmpty(t, result["token"], "Should receive auth token")
		assert.NotEmpty(t, result["user"], "Should receive user data")

		// Store token for subsequent tests
		tc.AuthToken = result["token"].(string)

		// Verify user object structure
		user := result["user"].(map[string]interface{})
		assert.NotEmpty(t, user["id"])
		assert.NotEmpty(t, user["username"])
		assert.NotEmpty(t, user["role"])
	})

	// Subtest: Failed Login (Invalid Credentials)
	t.Run("LoginFailedInvalidCredentials", func(t *testing.T) {
		loginData := map[string]interface{}{
			"username": "nonexistent",
			"password": "wrongpassword",
		}

		resp, err := tc.makeRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Login should fail with invalid credentials")
	})

	// Subtest: Access Protected Endpoint with Token
	t.Run("AccessProtectedEndpoint", func(t *testing.T) {
		if tc.AuthToken == "" {
			t.Skip("No auth token available")
		}

		resp, err := tc.makeRequest("GET", "/users/me", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should access protected endpoint with valid token")
	})

	// Subtest: Access Protected Endpoint without Token
	t.Run("AccessProtectedEndpointNoToken", func(t *testing.T) {
		savedToken := tc.AuthToken
		tc.AuthToken = ""

		resp, err := tc.makeRequest("GET", "/users/me", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should reject request without token")

		tc.AuthToken = savedToken
	})

	// Subtest: Logout
	t.Run("UserLogout", func(t *testing.T) {
		if tc.AuthToken == "" {
			t.Skip("No auth token available")
		}

		resp, err := tc.makeRequest("POST", "/auth/logout", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Logout should succeed")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
	})
}

// =============================================================================
// INTEGRATION TEST: Storage and File Operations
// =============================================================================

func TestStorageOperations(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	loginAndSetToken(t, tc)

	t.Run("ListStorageRoots", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/storage/roots", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["roots"])
	})

	t.Run("BrowseStoragePath", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/storage/list/?storage_id=local", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.True(t, result["success"].(bool))
			assert.NotNil(t, result["files"])
		}
	})
}

// =============================================================================
// INTEGRATION TEST: Media Operations
// =============================================================================

func TestMediaOperations(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	loginAndSetToken(t, tc)

	t.Run("ListMedia", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/media?page=1&limit=10", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["data"])
		assert.NotNil(t, result["pagination"])
	})

	t.Run("SearchMedia", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/media?search=test&type=video", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
	})

	t.Run("GetMediaDetails", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/media/1", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.True(t, result["success"].(bool))
			assert.NotNil(t, result["data"])
		}
	})

	t.Run("UpdateMedia", func(t *testing.T) {
		updateData := map[string]interface{}{
			"title":       "Updated Title",
			"description": "Updated description",
			"tags":        []string{"test", "integration"},
		}

		resp, err := tc.makeRequest("PUT", "/media/1", updateData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// =============================================================================
// INTEGRATION TEST: Analytics Operations
// =============================================================================

func TestAnalyticsOperations(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	loginAndSetToken(t, tc)

	t.Run("TrackEvent", func(t *testing.T) {
		eventData := map[string]interface{}{
			"event_type":  "media_view",
			"entity_type": "media_item",
			"entity_id":   123,
			"metadata": map[string]interface{}{
				"duration_watched": 45.2,
				"quality":          "1080p",
			},
		}

		resp, err := tc.makeRequest("POST", "/analytics/track", eventData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
	})

	t.Run("GetDashboardMetrics", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/analytics/dashboard", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["data"])
	})

	t.Run("GetUserEvents", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/analytics/events?limit=10", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["data"])
	})
}

// =============================================================================
// INTEGRATION TEST: Collections and Favorites
// =============================================================================

func TestCollectionsAndFavorites(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	loginAndSetToken(t, tc)

	var collectionID int

	t.Run("ListCollections", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/collections", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["data"])
	})

	t.Run("CreateCollection", func(t *testing.T) {
		collectionData := map[string]interface{}{
			"name":        fmt.Sprintf("Test Collection %d", time.Now().Unix()),
			"description": "Integration test collection",
			"type":        "custom",
			"tags":        []string{"test"},
			"privacy":     "private",
		}

		resp, err := tc.makeRequest("POST", "/collections", collectionData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["data"])

		collection := result["data"].(map[string]interface{})
		collectionID = int(collection["id"].(float64))
	})

	t.Run("GetCollectionDetails", func(t *testing.T) {
		if collectionID == 0 {
			t.Skip("No collection created")
		}

		resp, err := tc.makeRequest("GET", fmt.Sprintf("/collections/%d", collectionID), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
	})

	t.Run("DeleteCollection", func(t *testing.T) {
		if collectionID == 0 {
			t.Skip("No collection to delete")
		}

		resp, err := tc.makeRequest("DELETE", fmt.Sprintf("/collections/%d", collectionID), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ListFavorites", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/favorites", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
		assert.NotNil(t, result["data"])
	})
}

// =============================================================================
// INTEGRATION TEST: Error Handling
// =============================================================================

func TestErrorHandling(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	t.Run("NotFoundEndpoint", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/nonexistent/endpoint", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidJSONBody", func(t *testing.T) {
		req, err := http.NewRequest("POST", tc.BaseURL+"/auth/login", bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := tc.HTTPClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		// Use the server health URL, not the /api/v1 prefix
		req, err := http.NewRequest("DELETE", ts.URL+"/health", nil)
		require.NoError(t, err)

		resp, err := tc.HTTPClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

// =============================================================================
// INTEGRATION TEST: End-to-End User Journey
// =============================================================================

func TestEndToEndUserJourney(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	t.Run("CompleteUserJourney", func(t *testing.T) {
		// Step 1: Register new user
		timestamp := time.Now().Unix()
		registerData := map[string]interface{}{
			"username":       fmt.Sprintf("journey_user_%d", timestamp),
			"email":          fmt.Sprintf("journey_%d@example.com", timestamp),
			"password":       "SecurePassword123!",
			"full_name":      "Journey User",
			"terms_accepted": true,
		}

		resp, err := tc.makeRequest("POST", "/auth/register", registerData)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Step 1: Registration failed")

		// Step 2: Login
		loginData := map[string]interface{}{
			"username": registerData["username"],
			"password": registerData["password"],
		}

		resp, err = tc.makeRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 2: Login failed")

		var loginResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResult)
		resp.Body.Close()
		tc.AuthToken = loginResult["token"].(string)

		// Step 3: Browse storage
		resp, err = tc.makeRequest("GET", "/storage/roots", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 3: Browse storage failed")

		// Step 4: View media library
		resp, err = tc.makeRequest("GET", "/media?page=1&limit=10", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 4: View media failed")

		// Step 5: Create a collection
		collectionData := map[string]interface{}{
			"name":        fmt.Sprintf("Journey Collection %d", timestamp),
			"description": "E2E test collection",
			"type":        "custom",
		}

		resp, err = tc.makeRequest("POST", "/collections", collectionData)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Step 5: Create collection failed")

		// Step 6: Track analytics event
		eventData := map[string]interface{}{
			"event_type":  "user_journey_complete",
			"entity_type": "test",
			"entity_id":   1,
		}

		resp, err = tc.makeRequest("POST", "/analytics/track", eventData)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 6: Track event failed")

		// Step 7: View dashboard
		resp, err = tc.makeRequest("GET", "/analytics/dashboard", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 7: View dashboard failed")

		// Step 8: Logout
		resp, err = tc.makeRequest("POST", "/auth/logout", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 8: Logout failed")
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

func loginAndSetToken(t *testing.T, tc *TestContext) {
	t.Helper()
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	resp, err := tc.makeRequest("POST", "/auth/login", loginData)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Login should succeed")

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	tc.AuthToken = result["token"].(string)
}
