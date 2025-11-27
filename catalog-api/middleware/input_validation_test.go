package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInputValidation_SQLInjectionDetection(t *testing.T) {
	// Test SQL injection patterns
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid text", "Hello world", true},
		{"SQL SELECT", "SELECT * FROM users", false},
		{"SQL UNION", "1' UNION SELECT password FROM users", false},
		{"SQL COMMENT", "admin'--", false},
		{"SQL OR", "' OR '1'='1", false},
		{"SQL DROP", "DROP TABLE users", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectSQLInjection(tc.input)
			assert.Equal(t, tc.valid, !result)
		})
	}
}

func TestInputValidation_XSSDetection(t *testing.T) {
	// Test XSS patterns
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid text", "Hello world", true},
		{"Script tag", "<script>alert('xss')</script>", false},
		{"Iframe", "<iframe src='evil.com'></iframe>", false},
		{"JavaScript", "javascript:alert('xss')", false},
		{"Onload", "<img onload='alert(1)'>", false},
		{"Eval", "eval('malicious code')", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectXSS(tc.input)
			assert.Equal(t, tc.valid, !result)
		})
	}
}

func TestInputValidation_PathTraversalDetection(t *testing.T) {
	// Test path traversal patterns
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid path", "/home/user/file.txt", true},
		{"Basic traversal", "../../../etc/passwd", false},
		{"URL encoded traversal", "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd", false},
		{"Windows traversal", "..\\..\\..\\windows\\system32\\config\\sam", false},
		{"Unix traversal", "..././.../.../etc/passwd", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectPathTraversal(tc.input)
			assert.Equal(t, tc.valid, !result)
		})
	}
}

func TestInputValidation_GinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create router with validation middleware
	router := gin.New()
	router.Use(InputValidation(DefaultInputValidationConfig()))

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test SQL injection
	t.Run("Reject SQL Injection", func(t *testing.T) {
		jsonData := `{"search": "SELECT * FROM users"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "potential SQL injection")
	})

	// Test XSS
	t.Run("Reject XSS", func(t *testing.T) {
		jsonData := `{"comment": "<script>alert('xss')</script>"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "potential XSS")
	})

	// Test valid input
	t.Run("Allow valid input", func(t *testing.T) {
		jsonData := `{"name": "John Doe", "email": "john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})
}

func TestInputValidation_SizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultInputValidationConfig()
	config.MaxRequestBodySize = 100 // Very small limit

	router := gin.New()
	router.Use(InputValidation(config))

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test oversized body
	largeData := make([]byte, 200)
	for i := range largeData {
		largeData[i] = 'a'
	}
	
	jsonData := `{"data": "` + string(largeData) + `"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "request body too large")
}

func TestAdvancedRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultRateLimiterConfig()
	config.Rate = 2 // 2 requests per second
	config.Burst = 3 // Burst of 3

	router := gin.New()
	router.Use(AdvancedRateLimit(config))

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(http.StatusOK, gin.H{"call": callCount})
	})

	// Make multiple requests rapidly
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// First 3 should succeed (burst)
	// Next 2 should be rate limited
	assert.Equal(t, 3, callCount)
}

func TestInputValidation_Sanitization(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello\x00world", "helloworld"},
		{"正常文字", "正常文字"}, // Valid Unicode
	}

	for _, tc := range testCases {
		t.Run("Sanitize input", func(t *testing.T) {
			result := SanitizeInput(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInputValidation_SecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(InputValidation(DefaultInputValidationConfig()))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}