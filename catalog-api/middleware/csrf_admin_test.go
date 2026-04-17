package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestCSRF_W2_AdminGroupRejectsCrossOriginMutations locks in W2 from
// docs/nexus/remaining-work.md: admin routes wired behind the CSRF
// guard must refuse POST/PUT/DELETE when the X-CSRF-Token header is
// missing or stale, while GET/HEAD/OPTIONS mint a fresh token so
// clients can bootstrap a session.
func TestCSRF_W2_AdminGroupRejectsCrossOriginMutations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	guard, err := NewCSRF([]byte("this-secret-is-at-least-16-bytes-long"))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	admin := router.Group("/admin")
	admin.Use(guard.Handler())
	admin.GET("/probe", func(c *gin.Context) { c.Status(http.StatusOK) })
	admin.POST("/mutate", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Step 1: a GET mints a token.
	w1 := httptest.NewRecorder()
	r1, _ := http.NewRequest(http.MethodGet, "/admin/probe", nil)
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("probe got %d, want 200", w1.Code)
	}
	token := w1.Header().Get("X-CSRF-Token")
	if token == "" {
		t.Fatal("GET should mint X-CSRF-Token header")
	}
	var cookie string
	for _, c := range w1.Result().Cookies() {
		if c.Name == "__Host-csrf" {
			cookie = c.Value
		}
	}
	if cookie == "" {
		t.Fatal("GET should set __Host-csrf cookie")
	}

	// Step 2: a POST without header/cookie is rejected.
	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest(http.MethodPost, "/admin/mutate", strings.NewReader("{}"))
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusForbidden {
		t.Errorf("raw POST got %d, want 403", w2.Code)
	}

	// Step 3: a POST with mismatched header is rejected.
	w3 := httptest.NewRecorder()
	r3, _ := http.NewRequest(http.MethodPost, "/admin/mutate", strings.NewReader("{}"))
	r3.AddCookie(&http.Cookie{Name: "__Host-csrf", Value: cookie})
	r3.Header.Set("X-CSRF-Token", "tampered-value-not-matching-cookie")
	router.ServeHTTP(w3, r3)
	if w3.Code != http.StatusForbidden {
		t.Errorf("mismatched POST got %d, want 403", w3.Code)
	}

	// Step 4: a POST with the minted token + cookie passes.
	w4 := httptest.NewRecorder()
	r4, _ := http.NewRequest(http.MethodPost, "/admin/mutate", strings.NewReader("{}"))
	r4.AddCookie(&http.Cookie{Name: "__Host-csrf", Value: cookie})
	r4.Header.Set("X-CSRF-Token", token)
	router.ServeHTTP(w4, r4)
	if w4.Code != http.StatusOK {
		t.Errorf("valid POST got %d, want 200", w4.Code)
	}
}
