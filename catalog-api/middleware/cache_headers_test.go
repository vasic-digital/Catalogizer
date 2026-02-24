package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCacheHeaders_GETSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CacheHeaders(300))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	cc := w.Header().Get("Cache-Control")
	if cc != "public, max-age=300" {
		t.Errorf("expected Cache-Control 'public, max-age=300', got '%s'", cc)
	}

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("expected ETag header to be set")
	}
}

func TestCacheHeaders_POSTIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CacheHeaders(300))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"created": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	router.ServeHTTP(w, req)

	cc := w.Header().Get("Cache-Control")
	if cc != "" {
		t.Errorf("expected no Cache-Control for POST, got '%s'", cc)
	}
}

func TestCacheHeaders_ErrorIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CacheHeaders(300))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/error", nil)
	router.ServeHTTP(w, req)

	cc := w.Header().Get("Cache-Control")
	if cc != "" {
		t.Errorf("expected no Cache-Control for 404, got '%s'", cc)
	}
}

func TestCacheHeaders_IfNoneMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CacheHeaders(300))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	// First request to get ETag
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w1, req1)
	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header to be set on first request")
	}

	// Second request with If-None-Match
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("If-None-Match", etag)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotModified {
		t.Errorf("expected status 304, got %d", w2.Code)
	}
}

func TestStaticCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(StaticCacheHeaders())
	router.GET("/static/app.js", func(c *gin.Context) {
		c.String(http.StatusOK, "console.log('hello')")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/static/app.js", nil)
	router.ServeHTTP(w, req)

	cc := w.Header().Get("Cache-Control")
	expected := "public, max-age=31536000, immutable"
	if cc != expected {
		t.Errorf("expected Cache-Control '%s', got '%s'", expected, cc)
	}
}
