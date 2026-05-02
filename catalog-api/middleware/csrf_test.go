package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func buildCSRFRouter(t *testing.T, secret []byte) *gin.Engine {
	t.Helper()
	c, err := NewCSRF(secret)
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(c.Handler())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.POST("/write", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func TestCSRF_GETMintsToken(t *testing.T) {
	r := buildCSRFRouter(t, bytes.Repeat([]byte{0xab}, 32))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ping", nil))
	if w.Header().Get("X-CSRF-Token") == "" {
		t.Error("GET must mint a token header")
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), "__Host-csrf=") {
		t.Errorf("cookie not set: %q", w.Header().Get("Set-Cookie"))
	}
}

func TestCSRF_POSTWithoutTokenForbidden(t *testing.T) {
	r := buildCSRFRouter(t, bytes.Repeat([]byte{0xab}, 32))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/write", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestCSRF_POSTWithMatchingTokenAccepted(t *testing.T) {
	r := buildCSRFRouter(t, bytes.Repeat([]byte{0xab}, 32))

	// First, mint a token via GET.
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, httptest.NewRequest("GET", "/ping", nil))
	token := getRec.Header().Get("X-CSRF-Token")
	cookie := parseSetCookie(getRec.Header().Get("Set-Cookie"), "__Host-csrf")
	if token == "" || cookie == "" {
		t.Fatalf("mint failed: token=%q cookie=%q", token, cookie)
	}

	// Now POST with the matching header + cookie.
	req := httptest.NewRequest("POST", "/write", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-csrf", Value: cookie})
	req.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("valid request = %d %s, want 200", w.Code, w.Body.String())
	}
}

func TestCSRF_POSTWithWrongHeaderForbidden(t *testing.T) {
	r := buildCSRFRouter(t, bytes.Repeat([]byte{0xab}, 32))
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, httptest.NewRequest("GET", "/ping", nil))
	cookie := parseSetCookie(getRec.Header().Get("Set-Cookie"), "__Host-csrf")

	req := httptest.NewRequest("POST", "/write", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-csrf", Value: cookie})
	req.Header.Set("X-CSRF-Token", "not-the-right-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestCSRF_POSTWithoutCookieForbidden(t *testing.T) {
	r := buildCSRFRouter(t, bytes.Repeat([]byte{0xab}, 32))
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, httptest.NewRequest("GET", "/ping", nil))
	token := getRec.Header().Get("X-CSRF-Token")

	req := httptest.NewRequest("POST", "/write", nil)
	req.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestSubtleCompare(t *testing.T) {
	if subtleCompare("abc", "abc") != 1 {
		t.Error("equal strings must compare 1")
	}
	if subtleCompare("abc", "abd") != 0 {
		t.Error("unequal strings must compare 0")
	}
	if subtleCompare("ab", "abc") != 0 {
		t.Error("differing lengths must compare 0")
	}
}

// parseSetCookie extracts the value of the named cookie from a raw
// Set-Cookie header. Gin writes one cookie per header; this helper
// keeps the test file short.
func parseSetCookie(header, name string) string {
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return ""
	}
	kv := strings.SplitN(parts[0], "=", 2)
	if len(kv) != 2 || kv[0] != name {
		return ""
	}
	return kv[1]
}
