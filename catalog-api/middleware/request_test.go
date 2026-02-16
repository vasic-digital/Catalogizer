package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestIDGeneratesUUID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	handler := RequestID()
	handler(c)

	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}

	parsed, err := uuid.Parse(requestID)
	if err != nil {
		t.Fatalf("expected X-Request-ID to be a valid UUID, got '%s': %v", requestID, err)
	}
	if parsed.String() != requestID {
		t.Errorf("UUID round-trip mismatch: parsed '%s', header '%s'", parsed.String(), requestID)
	}

	contextID, exists := c.Get("request_id")
	if !exists {
		t.Fatal("expected request_id to be set in context")
	}
	if contextID != requestID {
		t.Errorf("expected context request_id '%s', got '%s'", requestID, contextID)
	}
}

func TestRequestIDPreservesExistingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	existingID := "my-custom-request-id-12345"
	c.Request.Header.Set("X-Request-ID", existingID)

	handler := RequestID()
	handler(c)

	requestID := w.Header().Get("X-Request-ID")
	if requestID != existingID {
		t.Errorf("expected preserved X-Request-ID '%s', got '%s'", existingID, requestID)
	}

	contextID, exists := c.Get("request_id")
	if !exists {
		t.Fatal("expected request_id to be set in context")
	}
	if contextID != existingID {
		t.Errorf("expected context request_id '%s', got '%s'", existingID, contextID)
	}
}

func TestRequestIDUniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		handler := RequestID()
		handler(c)

		id := w.Header().Get("X-Request-ID")
		if ids[id] {
			t.Fatalf("duplicate request ID generated: '%s'", id)
		}
		ids[id] = true
	}
}

func TestCORSDefaultOrigins(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:5173")

	handler := CORS()
	handler(c)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:5173', got '%s'", origin)
	}

	credentials := w.Header().Get("Access-Control-Allow-Credentials")
	if credentials != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got '%s'", credentials)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods != "POST, OPTIONS, GET, PUT, DELETE" {
		t.Errorf("expected Access-Control-Allow-Methods 'POST, OPTIONS, GET, PUT, DELETE', got '%s'", methods)
	}

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if headers == "" {
		t.Error("expected Access-Control-Allow-Headers to be set")
	}
}

func TestCORSSecondDefaultOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:3000")

	handler := CORS()
	handler(c)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:3000', got '%s'", origin)
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	c.Request.Header.Set("Origin", "http://evil.example.com")

	handler := CORS()
	handler(c)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin for disallowed origin, got '%s'", origin)
	}
}

func TestCORSCustomOrigins(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	c.Request.Header.Set("Origin", "https://app.example.com")

	handler := CORS()
	handler(c)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://app.example.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'https://app.example.com', got '%s'", origin)
	}
}

func TestCORSEmptyOriginHeader(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// No Origin header set

	handler := CORS()
	handler(c)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin without Origin header, got '%s'", origin)
	}

	// Methods and Headers should still be set
	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Access-Control-Allow-Methods to be set even without Origin")
	}
}

func TestCORSOptionsReturns204(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:5173")

	handler := CORS()
	handler(c)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS request, got %d", w.Code)
	}

	if !c.IsAborted() {
		t.Error("expected context to be aborted for OPTIONS request")
	}
}

func TestCORSNonOptionsCallsNext(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:5173")

	handler := CORS()
	handler(c)

	if c.IsAborted() {
		t.Error("expected context to NOT be aborted for GET request")
	}
}

func TestRateLimiterCallsNext(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)

	nextCalled := false
	c.Set("_test_marker", "before")

	router := gin.New()
	router.Use(RateLimiter(60))
	router.GET("/api/test", func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if !nextCalled {
		t.Error("expected RateLimiter to call next handler")
	}
	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
}

func TestRateLimiterWithDifferentLimits(t *testing.T) {
	limits := []int{10, 60, 100, 1000}
	for _, limit := range limits {
		router := gin.New()
		router.Use(RateLimiter(limit))

		called := false
		router.GET("/test", func(c *gin.Context) {
			called = true
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if !called {
			t.Errorf("RateLimiter(%d) did not call next handler", limit)
		}
		if resp.Code != http.StatusOK {
			t.Errorf("RateLimiter(%d) expected status 200, got %d", limit, resp.Code)
		}
	}
}
