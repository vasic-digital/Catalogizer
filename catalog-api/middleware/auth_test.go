package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecret = "test-middleware-secret-key"

func setupJWTMiddleware() *JWTMiddleware {
	return NewJWTMiddleware(testSecret)
}

// TestNewJWTMiddleware verifies middleware construction.
func TestNewJWTMiddleware(t *testing.T) {
	mw := NewJWTMiddleware("my-secret")
	assert.NotNil(t, mw)
	assert.Equal(t, []byte("my-secret"), mw.secretKey)
}

// TestGenerateToken_Success verifies token generation with valid parameters.
func TestGenerateToken_Success(t *testing.T) {
	mw := setupJWTMiddleware()

	token, err := mw.GenerateToken("testuser", "123", 24)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestGenerateAndValidateToken verifies round-trip token generation and validation.
func TestGenerateAndValidateToken(t *testing.T) {
	mw := setupJWTMiddleware()

	token, err := mw.GenerateToken("alice", "42", 24)
	require.NoError(t, err)

	claims, err := mw.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, "42", claims.Subject)
	assert.Equal(t, "catalog-api", claims.Issuer)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

// TestValidateToken_InvalidToken verifies that garbage tokens are rejected.
func TestValidateToken_InvalidToken(t *testing.T) {
	mw := setupJWTMiddleware()

	claims, err := mw.ValidateToken("not.a.valid.token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateToken_EmptyToken verifies that an empty token is rejected.
func TestValidateToken_EmptyToken(t *testing.T) {
	mw := setupJWTMiddleware()

	claims, err := mw.ValidateToken("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateToken_WrongSecret verifies that tokens signed with a different secret are rejected.
func TestValidateToken_WrongSecret(t *testing.T) {
	mw1 := NewJWTMiddleware("secret-one")
	mw2 := NewJWTMiddleware("secret-two")

	token, err := mw1.GenerateToken("user", "1", 24)
	require.NoError(t, err)

	claims, err := mw2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateToken_ExpiredToken verifies that expired tokens are rejected.
func TestValidateToken_ExpiredToken(t *testing.T) {
	mw := setupJWTMiddleware()

	// Create a token that is already expired
	claims := &Claims{
		Username: "expireduser",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "catalog-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	result, err := mw.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestRequireAuth_NoAuthorizationHeader verifies 401 when no Authorization header is set.
func TestRequireAuth_NoAuthorizationHeader(t *testing.T) {
	mw := setupJWTMiddleware()

	router := gin.New()
	router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Contains(t, body["error"], "Authorization header required")
}

// TestRequireAuth_InvalidHeaderFormat verifies 401 for malformed Authorization headers.
func TestRequireAuth_InvalidHeaderFormat(t *testing.T) {
	mw := setupJWTMiddleware()

	router := gin.New()
	router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"no prefix", "some-token-string"},
		{"wrong prefix", "Token abc123"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"three parts", "Bearer abc 123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// TestRequireAuth_InvalidToken verifies 401 when the token is invalid.
func TestRequireAuth_InvalidToken(t *testing.T) {
	mw := setupJWTMiddleware()

	router := gin.New()
	router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRequireAuth_ExpiredToken_HTTP verifies 401 when the token is expired.
func TestRequireAuth_ExpiredToken_HTTP(t *testing.T) {
	mw := setupJWTMiddleware()

	// Create an already-expired token
	claims := &Claims{
		Username: "expireduser",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "catalog-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	router := gin.New()
	router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRequireAuth_ValidToken_SetsContext verifies that a valid token populates the gin context.
func TestRequireAuth_ValidToken_SetsContext(t *testing.T) {
	mw := setupJWTMiddleware()

	token, err := mw.GenerateToken("bob", "99", 24)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "bob", username)

		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "99", userID)

		c.JSON(http.StatusOK, gin.H{"username": username, "user_id": userID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "bob", body["username"])
	assert.Equal(t, "99", body["user_id"])
}

// TestRequireAuth_ProtectedAndPublicRoutes verifies that protected and public routes coexist correctly.
func TestRequireAuth_ProtectedAndPublicRoutes(t *testing.T) {
	mw := setupJWTMiddleware()

	router := gin.New()
	router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public"})
	})
	router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	// Public route should work without auth
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Protected route should fail without auth
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// Protected route should work with valid auth
	token, err := mw.GenerateToken("user", "1", 24)
	require.NoError(t, err)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

// TestRequireAuth_DifferentHTTPMethods verifies auth middleware works across HTTP methods.
func TestRequireAuth_DifferentHTTPMethods(t *testing.T) {
	mw := setupJWTMiddleware()

	token, err := mw.GenerateToken("user", "1", 24)
	require.NoError(t, err)

	router := gin.New()
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": c.Request.Method})
	}

	router.GET("/resource", mw.RequireAuth(), handler)
	router.POST("/resource", mw.RequireAuth(), handler)
	router.PUT("/resource", mw.RequireAuth(), handler)
	router.DELETE("/resource", mw.RequireAuth(), handler)

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Without token: 401
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/resource", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			// With token: 200
			w2 := httptest.NewRecorder()
			req2 := httptest.NewRequest(method, "/resource", nil)
			req2.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w2, req2)
			assert.Equal(t, http.StatusOK, w2.Code)
		})
	}
}

// TestGenerateToken_DifferentExpirations verifies tokens with different expiration times.
func TestGenerateToken_DifferentExpirations(t *testing.T) {
	mw := setupJWTMiddleware()

	tests := []struct {
		name  string
		hours int
	}{
		{"1 hour", 1},
		{"24 hours", 24},
		{"168 hours (7 days)", 168},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := mw.GenerateToken("user", "1", tt.hours)
			require.NoError(t, err)

			claims, err := mw.ValidateToken(token)
			require.NoError(t, err)

			expectedExpiry := time.Now().Add(time.Duration(tt.hours) * time.Hour)
			assert.InDelta(t, expectedExpiry.Unix(), claims.ExpiresAt.Unix(), 5)
		})
	}
}

// TestTokenTampering verifies that modifying any part of the token invalidates it.
func TestTokenTampering(t *testing.T) {
	mw := setupJWTMiddleware()

	token, err := mw.GenerateToken("user", "1", 24)
	require.NoError(t, err)

	// Tamper by flipping a character
	tampered := []byte(token)
	tampered[len(tampered)/2] ^= 0xFF
	tamperedStr := string(tampered)

	claims, err := mw.ValidateToken(tamperedStr)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
