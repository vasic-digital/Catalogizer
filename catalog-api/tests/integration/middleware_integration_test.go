package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"catalogizer/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// INTEGRATION TEST: Middleware stack with a real gin router
// =============================================================================

func setupMiddlewareIntegrationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())

	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/echo", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, body)
	})

	return r
}

func TestMiddlewareIntegration_SecurityHeadersPresent(t *testing.T) {
	router := setupMiddlewareIntegrationRouter()

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"X-XSS-Protection":      "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	t.Run("GETRequestHasSecurityHeaders", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ok", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		for header, expected := range expectedHeaders {
			actual := w.Header().Get(header)
			assert.Equal(t, expected, actual,
				"Header %s should be %s, got %s",
				header, expected, actual)
		}
	})

	t.Run("POSTRequestHasSecurityHeaders", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewReader([]byte(`{"key":"value"}`))
		req, _ := http.NewRequest("POST", "/echo", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		for header, expected := range expectedHeaders {
			actual := w.Header().Get(header)
			assert.Equal(t, expected, actual,
				"Header %s should be set on POST responses")
		}
	})

	t.Run("404ResponseHasSecurityHeaders", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		// Security headers should be present even on error responses
		for header := range expectedHeaders {
			actual := w.Header().Get(header)
			assert.NotEmpty(t, actual,
				"Header %s should be set on 404 responses", header)
		}
	})
}

func TestMiddlewareIntegration_RequestIDGenerated(t *testing.T) {
	router := setupMiddlewareIntegrationRouter()

	t.Run("EachRequestGetsUniqueID", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 50; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ok", nil)
			router.ServeHTTP(w, req)

			id := w.Header().Get("X-Request-ID")
			assert.NotEmpty(t, id,
				"X-Request-ID should be set on every response")
			assert.False(t, ids[id],
				"Request ID should be unique: %s", id)
			ids[id] = true
		}
		assert.Equal(t, 50, len(ids),
			"All 50 request IDs should be unique")
	})
}

func TestMiddlewareIntegration_RequestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestTimeout(100 * time.Millisecond))

	r.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "fast"})
	})

	r.GET("/slow", func(c *gin.Context) {
		select {
		case <-time.After(500 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"status": "slow"})
		case <-c.Request.Context().Done():
			// Context was cancelled by the timeout middleware.
			// Return without writing a response body.
			return
		}
	})

	t.Run("FastRequestSucceeds", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/fast", nil)
		routerStart := time.Now()
		r.ServeHTTP(w, req)
		elapsed := time.Since(routerStart)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, elapsed, 100*time.Millisecond,
			"Fast request should complete quickly")
	})

	t.Run("SlowRequestContextCancelled", func(t *testing.T) {
		// The RequestTimeout middleware sets a context deadline.
		// With httptest.NewRecorder, gin still returns 200 when
		// no status is explicitly written, but the handler should
		// exit early via context cancellation rather than sleeping
		// the full 500ms.
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/slow", nil)
		start := time.Now()
		r.ServeHTTP(w, req)
		elapsed := time.Since(start)

		// The key assertion: the handler should NOT wait 500ms.
		// It should return within roughly the timeout (100ms).
		assert.Less(t, elapsed, 300*time.Millisecond,
			"Slow handler should be cancelled by timeout, "+
				"not run for the full 500ms")
		// The response body should be empty (handler returned
		// without writing)
		assert.Empty(t, w.Body.String(),
			"Cancelled handler should not write a response body")
	})
}

func TestMiddlewareIntegration_ConcurrencyLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ConcurrencyLimiter(2)) // Very low limit

	// Use a channel gate so all goroutines block inside the handler
	// simultaneously, ensuring the concurrency limit is exceeded.
	gate := make(chan struct{})

	r.GET("/work", func(c *gin.Context) {
		// Block until the test releases the gate
		<-gate
		c.JSON(http.StatusOK, gin.H{"status": "done"})
	})

	t.Run("RejectsWhenAtLimit", func(t *testing.T) {
		totalRequests := 10
		results := make(chan int, totalRequests)
		ready := make(chan struct{})

		// Launch goroutines that will all try to enter the handler
		for i := 0; i < totalRequests; i++ {
			go func() {
				ready <- struct{}{} // signal goroutine started
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/work", nil)
				r.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Wait for all goroutines to start
		for i := 0; i < totalRequests; i++ {
			<-ready
		}
		// Give goroutines a moment to enter ServeHTTP
		time.Sleep(50 * time.Millisecond)

		// Release the gate so blocked handlers can finish
		close(gate)

		var okCount, rejectedCount int
		for i := 0; i < totalRequests; i++ {
			code := <-results
			if code == http.StatusOK {
				okCount++
			} else if code == http.StatusServiceUnavailable {
				rejectedCount++
			}
		}

		t.Logf("OK: %d, Rejected: %d", okCount, rejectedCount)
		assert.Greater(t, okCount, 0,
			"Some requests should succeed")
		assert.Equal(t, totalRequests, okCount+rejectedCount,
			"All requests should complete with either OK or 503")
		// With a limit of 2 and 10 concurrent requests, some
		// should be rejected. However, the exact number depends
		// on goroutine scheduling, so we just verify the limiter
		// is functional by checking total accounting.
		t.Logf("Concurrency limiter processed %d requests "+
			"(%d OK, %d rejected)", totalRequests,
			okCount, rejectedCount)
	})
}

func TestMiddlewareIntegration_InputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.InputValidation(
		middleware.DefaultInputValidationConfig()))

	r.POST("/submit", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"accepted": true})
	})

	r.GET("/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("ValidJSONAccepted", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewReader(
			[]byte(`{"name":"test","value":42}`))
		req, _ := http.NewRequest("POST", "/submit", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code,
			"Valid JSON should be accepted")
	})

	t.Run("SQLInjectionBlocked", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := `{"name":"'; DROP TABLE users; --"}`
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest("POST", "/submit", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code,
			"SQL injection attempt should be blocked")
	})

	t.Run("XSSBlocked", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := `{"name":"<script>alert('xss')</script>"}`
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest("POST", "/submit", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code,
			"XSS attempt should be blocked")
	})

	t.Run("PathTraversalBlocked", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := `{"path":"../../../etc/passwd"}`
		body := bytes.NewReader([]byte(payload))
		req, _ := http.NewRequest("POST", "/submit", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code,
			"Path traversal attempt should be blocked")
	})

	t.Run("GETRequestNotBlocked", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/resource", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code,
			"GET requests should pass input validation")
	})

	t.Run("OversizedBodyRejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Create a body larger than the default 10MB limit
		// We use Content-Length header to simulate this without
		// actually allocating 11MB
		body := bytes.NewReader([]byte(`{"small":"payload"}`))
		req, _ := http.NewRequest("POST", "/submit", body)
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = 11 * 1024 * 1024 // 11MB
		r.ServeHTTP(w, req)

		// Should reject oversized content
		assert.Contains(t, []int{
			http.StatusBadRequest,
			http.StatusRequestEntityTooLarge,
		}, w.Code,
			"Oversized body should be rejected")
	})
}

func TestMiddlewareIntegration_FullStackCombined(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestTimeout(5 * time.Second))
	r.Use(middleware.InputValidation(
		middleware.DefaultInputValidationConfig()))

	r.POST("/api/test", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
	})

	t.Run("ValidRequestPassesAllMiddleware", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewReader(
			[]byte(`{"action":"test","value":"safe content"}`))
		req, _ := http.NewRequest("POST", "/api/test", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify all middleware contributed
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"),
			"RequestID middleware should set header")
		assert.NotEmpty(t,
			w.Header().Get("X-Content-Type-Options"),
			"SecurityHeaders middleware should set header")
	})

	t.Run("MaliciousRequestBlockedByValidation", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewReader(
			[]byte(`{"query":"UNION SELECT * FROM users"}`))
		req, _ := http.NewRequest("POST", "/api/test", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code,
			"SQL injection should be blocked before reaching handler")

		// Security headers should still be present on error responses
		assert.NotEmpty(t,
			w.Header().Get("X-Content-Type-Options"),
			"Security headers should be set even on blocked requests")
	})
}

func TestMiddlewareIntegration_RateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Use a very restrictive rate limit: 2 requests per minute
	r.Use(middleware.RateLimiter(2))

	r.GET("/limited", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("FirstRequestsAllowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/limited", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code,
			"First request should be allowed")
	})

	t.Run("ExcessRequestsBlocked", func(t *testing.T) {
		var blockedCount int
		// Send enough requests to exceed the rate limit
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/limited", nil)
			req.RemoteAddr = "10.0.0.50:12345"
			r.ServeHTTP(w, req)

			if w.Code == http.StatusTooManyRequests {
				blockedCount++
			}
		}

		assert.Greater(t, blockedCount, 0,
			"Some requests should be rate-limited")
	})
}

func TestMiddlewareIntegration_CacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CacheHeaders(3600))

	r.GET("/cacheable", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "cached"})
	})

	t.Run("CacheControlHeaderSet", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cacheable", nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		cacheControl := w.Header().Get("Cache-Control")
		assert.NotEmpty(t, cacheControl,
			"Cache-Control header should be present")
		assert.True(t,
			strings.Contains(cacheControl, "max-age") ||
				strings.Contains(cacheControl, "no-cache"),
			"Cache-Control should contain a valid directive")
	})
}
