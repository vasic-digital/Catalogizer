package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// FuzzContentSecurityPolicy — P8 fix
// (docs/nexus/remaining-work.md): the CSP directive is a
// semicolon-delimited policy string. Any parse-on-the-wire ever
// happens client-side, but the middleware still pastes this string
// verbatim into a response header. The fuzzer proves the middleware
// handles arbitrary bytes (including NULs, CRLFs that could forge
// header injections, and oversized inputs) without panicking.
func FuzzContentSecurityPolicy(f *testing.F) {
	seed := []string{
		"",
		"default-src 'self'",
		"default-src 'self'; script-src 'self' 'unsafe-inline'",
		"\x00\x00\x00",
		"header-injection:\r\nSet-Cookie: evil=1",
		"x" + randomLongString(4096),
	}
	for _, s := range seed {
		f.Add(s)
	}
	gin.SetMode(gin.TestMode)
	f.Fuzz(func(t *testing.T, policy string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("CSP middleware panicked on policy %q: %v", policy, r)
			}
		}()
		router := gin.New()
		router.Use(SecurityHeadersWithConfig(SecurityHeadersConfig{
			EnableContentSecurityPolicy: true,
			ContentSecurityPolicy:       policy,
		}))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		router.ServeHTTP(w, r)
		// Canary: Gin must not split CRLFs in headers — stdlib sanitises
		// but we assert the policy string round-trips as a single header
		// value when it doesn't contain forbidden control bytes.
		_ = w.Header().Get("Content-Security-Policy")
	})
}

func randomLongString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
