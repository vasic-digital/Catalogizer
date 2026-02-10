package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

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

func newTestContext() *TestContext {
	return &TestContext{
		BaseURL: "http://localhost:8080/api/v1",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Helper to check server availability
func (tc *TestContext) isServerAvailable() bool {
	resp, err := tc.HTTPClient.Get("http://localhost:8080/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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

// =============================================================================
// INTEGRATION TEST: Complete Authentication Flow
// =============================================================================

func TestAuthenticationFlow(t *testing.T) {
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

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
		assert.NotEmpty(t, result["user"])
	})

	// Subtest: User Login
	t.Run("UserLogin", func(t *testing.T) {
		// Use default admin credentials for login test
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
		// Temporarily remove token
		savedToken := tc.AuthToken
		tc.AuthToken = ""

		resp, err := tc.makeRequest("GET", "/users/me", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should reject request without token")

		// Restore token
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
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

	// Login first
	loginAndSetToken(t, tc)

	// Subtest: List Storage Roots
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

	// Subtest: Browse Storage Path
	t.Run("BrowseStoragePath", func(t *testing.T) {
		// Assuming local storage is available
		resp, err := tc.makeRequest("GET", "/storage/list/?storage_id=local", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// May return 200 or 404 depending on storage configuration
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
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

	// Login first
	loginAndSetToken(t, tc)

	var mediaID int

	// Subtest: List Media (Empty or Existing)
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

	// Subtest: Search Media
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

	// Subtest: Get Media Details (if media exists)
	t.Run("GetMediaDetails", func(t *testing.T) {
		// Try to get media with ID 1
		resp, err := tc.makeRequest("GET", "/media/1", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		// May return 200 or 404 depending on data
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.True(t, result["success"].(bool))
			assert.NotNil(t, result["data"])

			media := result["data"].(map[string]interface{})
			assert.NotEmpty(t, media["id"])
			assert.NotEmpty(t, media["title"])

			// Store media ID for other tests
			mediaID = int(media["id"].(float64))
		}
	})

	// Subtest: Update Media (if media exists)
	t.Run("UpdateMedia", func(t *testing.T) {
		if mediaID == 0 {
			t.Skip("No media available to update")
		}

		updateData := map[string]interface{}{
			"title":       "Updated Title",
			"description": "Updated description",
			"tags":        []string{"test", "integration"},
		}

		resp, err := tc.makeRequest("PUT", fmt.Sprintf("/media/%d", mediaID), updateData)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result["success"].(bool))
	})
}

// =============================================================================
// INTEGRATION TEST: Analytics Operations
// =============================================================================

func TestAnalyticsOperations(t *testing.T) {
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

	// Login first
	loginAndSetToken(t, tc)

	// Subtest: Track Event
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

	// Subtest: Get Dashboard Metrics
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

	// Subtest: Get User Events
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
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

	// Login first
	loginAndSetToken(t, tc)

	var collectionID int

	// Subtest: List Collections
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

	// Subtest: Create Collection
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

	// Subtest: Get Collection Details
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

	// Subtest: Delete Collection
	t.Run("DeleteCollection", func(t *testing.T) {
		if collectionID == 0 {
			t.Skip("No collection to delete")
		}

		resp, err := tc.makeRequest("DELETE", fmt.Sprintf("/collections/%d", collectionID), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Subtest: List Favorites
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
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

	// Subtest: 404 Not Found
	t.Run("NotFoundEndpoint", func(t *testing.T) {
		resp, err := tc.makeRequest("GET", "/nonexistent/endpoint", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Subtest: Invalid JSON Body
	t.Run("InvalidJSONBody", func(t *testing.T) {
		req, err := http.NewRequest("POST", tc.BaseURL+"/auth/login", bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := tc.HTTPClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Subtest: Method Not Allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		resp, err := tc.makeRequest("DELETE", "/health", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

// =============================================================================
// INTEGRATION TEST: End-to-End User Journey
// =============================================================================

func TestEndToEndUserJourney(t *testing.T) {
	tc := newTestContext()
	if !tc.isServerAvailable() {
		t.Skip("Server not available - skipping integration test")
	}

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
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Step 2: Login failed")

		var loginResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResult)
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
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	resp, err := tc.makeRequest("POST", "/auth/login", loginData)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skip("Cannot login - skipping authenticated tests")
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	tc.AuthToken = result["token"].(string)
}
