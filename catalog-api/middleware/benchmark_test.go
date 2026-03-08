package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// BenchmarkSecurityHeaders measures the overhead of setting security headers
// on every response (5 headers unconditionally, plus HSTS on TLS).
func BenchmarkSecurityHeaders(b *testing.B) {
	handler := SecurityHeaders()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkRequestTimeout measures the overhead of wrapping the request
// context with a deadline (context.WithTimeout + Request.WithContext).
func BenchmarkRequestTimeout(b *testing.B) {
	handler := RequestTimeout(30 * time.Second)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkConcurrencyLimiter measures the acquire/release cycle of the
// semaphore-based concurrency limiter under no contention.
func BenchmarkConcurrencyLimiter(b *testing.B) {
	handler := ConcurrencyLimiter(1000)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkRequestID measures UUID generation and header/context assignment
// per request (the hot path when no X-Request-ID header is provided).
func BenchmarkRequestID(b *testing.B) {
	handler := RequestID()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkRequestIDExisting measures the fast path where an X-Request-ID
// header is already present (no UUID generation needed).
func BenchmarkRequestIDExisting(b *testing.B) {
	handler := RequestID()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("X-Request-ID", "pre-existing-request-id-value")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkFullMiddlewareChain measures the combined overhead of all four
// performance-critical middleware layers executing in sequence on a single
// request, using a real gin.Engine to wire the chain.
func BenchmarkFullMiddlewareChain(b *testing.B) {
	router := gin.New()
	router.Use(RequestID())
	router.Use(SecurityHeaders())
	router.Use(RequestTimeout(30 * time.Second))
	router.Use(ConcurrencyLimiter(1000))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkFullMiddlewareChainParallel measures the middleware chain under
// concurrent load, which is the realistic production scenario. This stresses
// the concurrency limiter's semaphore and the rate of UUID generation.
func BenchmarkFullMiddlewareChainParallel(b *testing.B) {
	router := gin.New()
	router.Use(RequestID())
	router.Use(SecurityHeaders())
	router.Use(RequestTimeout(30 * time.Second))
	router.Use(ConcurrencyLimiter(1000))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}
