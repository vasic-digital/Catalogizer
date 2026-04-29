package integration

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"catalogizer/internal/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// JSON helpers used across auth integration tests
// =============================================================================

// decodeJSON reads the response body and decodes it into dest, closing the body.
func decodeJSON(t *testing.T, resp *http.Response, dest interface{}) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, dest))
}

// decodeJSONSafe decodes resp body into dest without requiring *testing.T.
func decodeJSONSafe(resp *http.Response, dest interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

// decodeJSONBody decodes an http.Request body into dest.
func decodeJSONBody(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dest)
}

// =============================================================================
// INTEGRATION TEST: Auth flow with in-memory SQLite
// =============================================================================

func TestAuthIntegration_UserCreationAndLookup(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	t.Run("InsertUserAndQueryBack", func(t *testing.T) {
		username := fmt.Sprintf("authtest_%d", time.Now().UnixNano())
		email := fmt.Sprintf("%s@example.com", username)
		passwordHash := "hashed_password_placeholder"

		result, err := db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, ?, 1, 1)`,
			username, email, passwordHash,
		)
		require.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// Verify user exists
		var storedUsername, storedEmail string
		var isActive bool
		err = db.QueryRow(
			`SELECT username, email, is_active FROM users WHERE username = ?`,
			username,
		).Scan(&storedUsername, &storedEmail, &isActive)
		require.NoError(t, err)

		assert.Equal(t, username, storedUsername)
		assert.Equal(t, email, storedEmail)
		assert.True(t, isActive)
	})

	t.Run("DuplicateUsernameRejected", func(t *testing.T) {
		username := fmt.Sprintf("dupuser_%d", time.Now().UnixNano())
		email1 := fmt.Sprintf("%s_1@example.com", username)
		email2 := fmt.Sprintf("%s_2@example.com", username)

		_, err := db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash1', 1, 1)`,
			username, email1,
		)
		require.NoError(t, err)

		// Second insert with same username should fail (UNIQUE constraint)
		_, err = db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash2', 1, 1)`,
			username, email2,
		)
		assert.Error(t, err, "Duplicate username should be rejected")
		assert.Contains(t, err.Error(), "UNIQUE",
			"Error should be a unique constraint violation")
	})

	t.Run("DuplicateEmailRejected", func(t *testing.T) {
		email := fmt.Sprintf("dupemail_%d@example.com", time.Now().UnixNano())
		user1 := fmt.Sprintf("user1_%d", time.Now().UnixNano())
		user2 := fmt.Sprintf("user2_%d", time.Now().UnixNano())

		_, err := db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash1', 1, 1)`,
			user1, email,
		)
		require.NoError(t, err)

		_, err = db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash2', 1, 1)`,
			user2, email,
		)
		assert.Error(t, err, "Duplicate email should be rejected")
	})
}

func TestAuthIntegration_LoginFlowViaHTTP(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)

	t.Run("RegisterLoginAccessProtectedLogout", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		username := fmt.Sprintf("auth_flow_%d", timestamp)

		// Step 1: Register
		registerData := map[string]interface{}{
			"username":       username,
			"email":          fmt.Sprintf("%s@example.com", username),
			"password":       "Str0ngP@ssword!",
			"terms_accepted": true,
		}
		resp, err := tc.makeRequest("POST", "/auth/register", registerData)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Registration should succeed")

		// Step 2: Login with new credentials
		loginData := map[string]interface{}{
			"username": username,
			"password": "Str0ngP@ssword!",
		}
		resp, err = tc.makeRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)

		var loginResult map[string]interface{}
		decodeJSON(t, resp, &loginResult)
		require.NotEmpty(t, loginResult["token"],
			"Login should return a token")

		tc.AuthToken = loginResult["token"].(string)

		// Step 3: Access protected resource
		resp, err = tc.makeRequest("GET", "/users/me", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Should access protected endpoint with valid token")

		// Step 4: Logout
		resp, err = tc.makeRequest("POST", "/auth/logout", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Logout should succeed")
	})

	t.Run("InvalidCredentialsRejected", func(t *testing.T) {
		loginData := map[string]interface{}{
			"username": "nonexistent_user",
			"password": "wrong_password",
		}
		resp, err := tc.makeRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("EmptyTokenRejected", func(t *testing.T) {
		tc.AuthToken = ""
		resp, err := tc.makeRequest("GET", "/users/me", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("InvalidTokenRejected", func(t *testing.T) {
		tc.AuthToken = "invalid-token-value-that-does-not-exist"
		resp, err := tc.makeRequest("GET", "/users/me", nil)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthIntegration_TokenUniqueness(t *testing.T) {
	ts := setupUserFlowsServer(t)

	tokens := make(map[string]bool)
	var mu sync.Mutex

	loginCount := 20
	for i := 0; i < loginCount; i++ {
		tc := newTestContext(ts.URL)
		loginData := map[string]interface{}{
			"username": "admin",
			"password": "admin123",
		}
		resp, err := tc.makeRequest("POST", "/auth/login", loginData)
		require.NoError(t, err)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		token := result["token"].(string)

		mu.Lock()
		assert.False(t, tokens[token],
			"Token should be unique, got duplicate: %s", token)
		tokens[token] = true
		mu.Unlock()
	}
	assert.Equal(t, loginCount, len(tokens),
		"Each login should produce a unique token")
}

func TestAuthIntegration_PasswordHashNotExposed(t *testing.T) {
	ts := setupUserFlowsServer(t)
	tc := newTestContext(ts.URL)
	loginAndSetToken(t, tc)

	resp, err := tc.makeRequest("GET", "/users/me", nil)
	require.NoError(t, err)

	var result map[string]interface{}
	decodeJSON(t, resp, &result)

	data, ok := result["data"].(map[string]interface{})
	if ok {
		_, hasPassword := data["password"]
		_, hasPasswordHash := data["password_hash"]
		_, hasSalt := data["salt"]

		assert.False(t, hasPassword,
			"Password should not be in API response")
		assert.False(t, hasPasswordHash,
			"Password hash should not be in API response")
		assert.False(t, hasSalt,
			"Salt should not be in API response")
	}
}

func TestAuthIntegration_ConcurrentLogins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent login integration test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	ts := setupUserFlowsServer(t)

	concurrentLogins := 20
	var wg sync.WaitGroup
	var successCount atomic.Int64
	var mu sync.Mutex
	tokens := make([]string, 0, concurrentLogins)

	for i := 0; i < concurrentLogins; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tc := newTestContext(ts.URL)
			loginData := map[string]interface{}{
				"username": "admin",
				"password": "admin123",
			}
			resp, err := tc.makeRequest("POST", "/auth/login", loginData)
			if err != nil {
				return
			}

			if resp.StatusCode == http.StatusOK {
				successCount.Add(1)
				var result map[string]interface{}
				if decodeJSONSafe(resp, &result) == nil {
					if token, ok := result["token"].(string); ok {
						mu.Lock()
						tokens = append(tokens, token)
						mu.Unlock()
					}
				}
			} else {
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(concurrentLogins), successCount.Load(),
		"All concurrent logins should succeed")
	assert.Equal(t, concurrentLogins, len(tokens),
		"Each login should produce a token")
}

// =============================================================================
// Auth integration test for failed_login_attempts tracking
// =============================================================================

func TestAuthIntegration_FailedLoginAttemptsTracking(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	username := fmt.Sprintf("locktest_%d", time.Now().UnixNano())
	_, err := db.Exec(
		`INSERT INTO users (username, email, password_hash, is_active,
		 role_id, failed_login_attempts)
		 VALUES (?, ?, 'bcrypt_hash', 1, 1, 0)`,
		username, fmt.Sprintf("%s@example.com", username),
	)
	require.NoError(t, err)

	// Simulate incrementing failed_login_attempts
	for i := 1; i <= 5; i++ {
		_, err := db.Exec(
			`UPDATE users SET failed_login_attempts = ? WHERE username = ?`,
			i, username,
		)
		require.NoError(t, err)
	}

	var attempts int
	err = db.QueryRow(
		`SELECT failed_login_attempts FROM users WHERE username = ?`,
		username,
	).Scan(&attempts)
	require.NoError(t, err)
	assert.Equal(t, 5, attempts,
		"Failed login attempts should be tracked")

	// Simulate lock
	_, err = db.Exec(
		`UPDATE users SET is_locked = 1,
		 locked_until = datetime('now', '+30 minutes')
		 WHERE username = ?`,
		username,
	)
	require.NoError(t, err)

	var isLocked bool
	err = db.QueryRow(
		`SELECT is_locked FROM users WHERE username = ?`,
		username,
	).Scan(&isLocked)
	require.NoError(t, err)
	assert.True(t, isLocked,
		"Account should be locked after max failed attempts")

	// Simulate successful login resetting attempts
	_, err = db.Exec(
		`UPDATE users SET failed_login_attempts = 0, is_locked = 0,
		 locked_until = NULL WHERE username = ?`,
		username,
	)
	require.NoError(t, err)

	err = db.QueryRow(
		`SELECT failed_login_attempts FROM users WHERE username = ?`,
		username,
	).Scan(&attempts)
	require.NoError(t, err)
	assert.Equal(t, 0, attempts,
		"Failed attempts should reset on successful login")
}

// =============================================================================
// Helper: generate a random hex token (used in place of JWT for mock)
// =============================================================================

func generateMockToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// =============================================================================
// Helper: minimal auth-only test server for JWT-like flow
// =============================================================================

func setupAuthTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	var mu sync.Mutex
	validTokens := map[string]string{} // token -> username

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := decodeJSONBody(r, &body); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		if body.Username == "admin" && body.Password == "admin123" {
			token := generateMockToken()
			mu.Lock()
			validTokens[token] = body.Username
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w,
				`{"token":"%s","user":{"id":1,"username":"%s"}}`,
				token, body.Username)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid credentials"}`))
	})

	mux.HandleFunc("/api/v1/protected", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"missing token"}`))
			return
		}
		token := auth[7:]
		mu.Lock()
		_, valid := validTokens[token]
		mu.Unlock()
		if !valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid token"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":"protected content"}`))
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(func() { ts.Close() })
	return ts
}

func TestAuthIntegration_MinimalAuthServer(t *testing.T) {
	ts := setupAuthTestServer(t)
	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("LoginAndAccessProtected", func(t *testing.T) {
		loginBody := strings.NewReader(
			`{"username":"admin","password":"admin123"}`)
		resp, err := client.Post(
			ts.URL+"/api/v1/auth/login",
			"application/json", loginBody)
		require.NoError(t, err)

		var loginResult map[string]interface{}
		decodeJSON(t, resp, &loginResult)
		token := loginResult["token"].(string)
		assert.NotEmpty(t, token)

		req, _ := http.NewRequest("GET",
			ts.URL+"/api/v1/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ProtectedWithoutToken", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/api/v1/protected")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("ProtectedWithBogusToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET",
			ts.URL+"/api/v1/protected", nil)
		req.Header.Set("Authorization", "Bearer bogus-token-value")
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
