package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// BenchmarkRateLimitCheck measures the per-request cost of the in-memory
// token-bucket rate limiter. This includes mutex acquisition, token
// refill calculation, and token consumption -- the hot path on every
// incoming HTTP request.
func BenchmarkRateLimitCheck(b *testing.B) {
	handler := RateLimiter(10000)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkRateLimitCheck_Parallel measures rate limiter performance
// under concurrent goroutine contention, which stresses the mutex
// protecting the per-IP bucket map.
func BenchmarkRateLimitCheck_Parallel(b *testing.B) {
	handler := RateLimiter(10000)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			handler(c)
		}
	})
}

// BenchmarkRateLimitCheck_DistinctIPs measures rate limiter performance
// when requests arrive from many different IP addresses, which exercises
// the bucket map growth path.
func BenchmarkRateLimitCheck_DistinctIPs(b *testing.B) {
	handler := RateLimiter(10000)

	// Pre-generate distinct IP request objects
	ipCount := 256
	requests := make([]*http.Request, ipCount)
	for i := 0; i < ipCount; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.RemoteAddr = "10.0." + itoa(i/256) + "." + itoa(i%256) + ":12345"
		requests[i] = req
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = requests[i%ipCount]
		handler(c)
	}
}

// BenchmarkJWTGenerateToken measures the cost of generating a JWT token
// via the middleware's GenerateToken method, which signs claims with
// HMAC-SHA256.
func BenchmarkJWTGenerateToken(b *testing.B) {
	mw := NewJWTMiddleware("benchmark-secret-key-for-jwt")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := mw.GenerateToken("benchuser", "1", 24)
		if err != nil {
			b.Fatalf("GenerateToken failed: %v", err)
		}
	}
}

// BenchmarkJWTValidateToken measures the cost of validating an existing
// JWT token via the middleware's ValidateToken method, which parses and
// verifies the HMAC signature.
func BenchmarkJWTValidateToken(b *testing.B) {
	mw := NewJWTMiddleware("benchmark-secret-key-for-jwt")
	token, err := mw.GenerateToken("benchuser", "1", 24)
	if err != nil {
		b.Fatalf("setup: GenerateToken failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := mw.ValidateToken(token)
		if err != nil {
			b.Fatalf("ValidateToken failed: %v", err)
		}
	}
}

// BenchmarkJWTRequireAuth measures the full authentication middleware
// path: extracting the Bearer token from the Authorization header,
// parsing the JWT, validating the signature, and setting user context
// values.
func BenchmarkJWTRequireAuth(b *testing.B) {
	mw := NewJWTMiddleware("benchmark-secret-key-for-jwt")
	handler := mw.RequireAuth()

	token, err := mw.GenerateToken("benchuser", "1", 24)
	if err != nil {
		b.Fatalf("setup: GenerateToken failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkCORS measures the overhead of the CORS middleware on a
// standard GET request (non-preflight).
func BenchmarkCORS(b *testing.B) {
	handler := CORS()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// BenchmarkCORS_Preflight measures the CORS middleware on an OPTIONS
// preflight request, which aborts early with 204.
func BenchmarkCORS_Preflight(b *testing.B) {
	handler := CORS()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		handler(c)
	}
}

// itoa converts an int to a string without importing strconv, to keep
// the benchmark setup allocation-free in the hot path.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 3)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
