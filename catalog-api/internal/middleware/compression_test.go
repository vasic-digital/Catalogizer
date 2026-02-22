package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCompressionMiddleware_Brotli(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": strings.Repeat("test", 500)}) // 2000 bytes > min size
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Content-Encoding"), "br", "should use Brotli compression when supported")
	assert.Contains(t, resp.Header().Get("Vary"), "Accept-Encoding", "should set Vary header")
	assert.True(t, len(resp.Body.Bytes()) < 2000, "response should be compressed")
}

func TestCompressionMiddleware_Gzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": strings.Repeat("test", 500)})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"), "should use gzip compression when Brotli not supported")
	assert.True(t, len(resp.Body.Bytes()) < 2000, "response should be compressed")
}

func TestCompressionMiddleware_NoCompression(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "small"}) // < min size
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Header().Get("Content-Encoding"), "small response should not be compressed")
}

func TestCompressionMiddleware_ExcludedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": strings.Repeat("metrics", 500)})
	})

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Accept-Encoding", "br")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Header().Get("Content-Encoding"), "/metrics path should be excluded from compression")
}

func TestCompressionMiddleware_ExcludedContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := DefaultCompressionConfig()
	config.ExcludedContentTypes = []string{"image/png"}
	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/image", func(c *gin.Context) {
		c.Header("Content-Type", "image/png")
		c.String(http.StatusOK, strings.Repeat("fake image data", 500))
	})

	req := httptest.NewRequest("GET", "/image", nil)
	req.Header.Set("Accept-Encoding", "br")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Header().Get("Content-Encoding"), "image/png should be excluded from compression")
}
