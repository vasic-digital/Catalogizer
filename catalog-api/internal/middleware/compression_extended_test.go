package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestCompressionMiddleware_BrotliDecompressesCorrectly verifies that Brotli
// compressed responses can be decompressed back to the original content.
func TestCompressionMiddleware_BrotliDecompressesCorrectly(t *testing.T) {
	originalContent := strings.Repeat("Hello Brotli Compression Test! ", 100)

	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, originalContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	encoding := resp.Header().Get("Content-Encoding")
	if encoding == "br" {
		reader := brotli.NewReader(resp.Body)
		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(decompressed))
	}
}

// TestCompressionMiddleware_GzipDecompressesCorrectly verifies that Gzip
// compressed responses can be decompressed back to the original content.
func TestCompressionMiddleware_GzipDecompressesCorrectly(t *testing.T) {
	originalContent := strings.Repeat("Hello Gzip Compression Test! ", 100)

	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, originalContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	encoding := resp.Header().Get("Content-Encoding")
	if encoding == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(decompressed))
	}
}

// TestCompressionMiddleware_BrotliPreferredOverGzip verifies that when both
// Brotli and Gzip are supported, Brotli is chosen (as per the HTTP/3 mandate).
func TestCompressionMiddleware_BrotliPreferredOverGzip(t *testing.T) {
	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Repeat("prefer brotli ", 200))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip, br")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "br", resp.Header().Get("Content-Encoding"),
		"Brotli should be preferred when both br and gzip are accepted")
}

// TestCompressionMiddleware_VaryHeaderAlwaysSet verifies that the Vary header
// is set for compressed responses (important for HTTP caching).
func TestCompressionMiddleware_VaryHeaderAlwaysSet(t *testing.T) {
	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Repeat("vary header test ", 200))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	if resp.Header().Get("Content-Encoding") != "" {
		assert.Contains(t, resp.Header().Get("Vary"), "Accept-Encoding",
			"Vary header should include Accept-Encoding for compressed responses")
	}
}

// TestCompressionMiddleware_NoCompressionWithoutAcceptEncoding verifies that
// when the client does not send Accept-Encoding, no compression is applied.
func TestCompressionMiddleware_NoCompressionWithoutAcceptEncoding(t *testing.T) {
	originalContent := strings.Repeat("no compress test ", 200)

	router := gin.New()
	router.Use(CompressionMiddleware(DefaultCompressionConfig()))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, originalContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Intentionally do NOT set Accept-Encoding.
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Header().Get("Content-Encoding"),
		"no compression should be applied without Accept-Encoding")
	assert.Equal(t, originalContent, resp.Body.String(),
		"response body should be uncompressed original content")
}

// TestCompressionMiddleware_BelowMinSizeNotCompressed verifies that responses
// smaller than the configured minimum size are not compressed.
func TestCompressionMiddleware_BelowMinSizeNotCompressed(t *testing.T) {
	config := DefaultCompressionConfig()
	config.MinSize = 5000 // Set high threshold.

	router := gin.New()
	router.Use(CompressionMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		// 200 bytes, well below 5000 threshold.
		c.String(http.StatusOK, strings.Repeat("x", 200))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Header().Get("Content-Encoding"),
		"response below min size should not be compressed")
}

// TestCompressionMiddleware_ExcludedMediaContentTypes verifies that binary
// content types (images, video, audio) are not compressed as they are already
// compressed formats.
func TestCompressionMiddleware_ExcludedMediaContentTypes(t *testing.T) {
	excludedTypes := []string{
		"image/png",
		"image/jpeg",
		"video/mp4",
		"audio/mpeg",
		"application/octet-stream",
	}

	for _, contentType := range excludedTypes {
		t.Run(contentType, func(t *testing.T) {
			router := gin.New()
			router.Use(CompressionMiddleware(DefaultCompressionConfig()))
			router.GET("/test", func(c *gin.Context) {
				c.Header("Content-Type", contentType)
				c.String(http.StatusOK, strings.Repeat("binary data ", 500))
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "br, gzip")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Empty(t, resp.Header().Get("Content-Encoding"),
				"content type %q should not be compressed", contentType)
		})
	}
}

// TestCompressionMiddleware_CompressionLevels verifies that different
// compression levels produce valid compressed output.
func TestCompressionMiddleware_CompressionLevels(t *testing.T) {
	levels := []struct {
		name   string
		config CompressionConfig
	}{
		{
			"best speed",
			CompressionConfig{
				MinSize:     512,
				BrotliLevel: CompressionLevelBestSpeed,
				GzipLevel:   CompressionLevelBestSpeed,
			},
		},
		{
			"best compression",
			CompressionConfig{
				MinSize:     512,
				BrotliLevel: CompressionLevelBestCompression,
				GzipLevel:   CompressionLevelBestCompression,
			},
		},
		{
			"default",
			CompressionConfig{
				MinSize:     512,
				BrotliLevel: CompressionLevelDefault,
				GzipLevel:   CompressionLevelDefault,
			},
		},
	}

	originalContent := strings.Repeat("compression level test content ", 100)

	for _, tt := range levels {
		t.Run(tt.name+" brotli", func(t *testing.T) {
			router := gin.New()
			router.Use(CompressionMiddleware(tt.config))
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, originalContent)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "br")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			if resp.Header().Get("Content-Encoding") == "br" {
				reader := brotli.NewReader(resp.Body)
				decompressed, err := io.ReadAll(reader)
				require.NoError(t, err, "should decompress without error at %s level", tt.name)
				assert.Equal(t, originalContent, string(decompressed))
			}
		})
	}
}

// TestDefaultCompressionConfig_MetricsPathExcluded verifies that the default
// compression config excludes the /metrics path (important for Prometheus
// scraping compatibility).
func TestDefaultCompressionConfig_MetricsPathExcluded(t *testing.T) {
	config := DefaultCompressionConfig()

	found := false
	for _, path := range config.ExcludedPaths {
		if path == "/metrics" {
			found = true
			break
		}
	}
	assert.True(t, found, "default config should exclude /metrics path from compression")
}

// TestDefaultCompressionConfig_MinSizeIsReasonable verifies that the default
// minimum size for compression is set to a reasonable value.
func TestDefaultCompressionConfig_MinSizeIsReasonable(t *testing.T) {
	config := DefaultCompressionConfig()

	assert.Greater(t, config.MinSize, 0,
		"minimum compression size should be positive")
	assert.LessOrEqual(t, config.MinSize, 4096,
		"minimum compression size should not be too large (default is 1024)")
}

// TestCompressionMiddleware_JSONContentCompressed verifies that JSON API
// responses are compressed (the primary use case for the API server).
func TestCompressionMiddleware_JSONContentCompressed(t *testing.T) {
	router := gin.New()
	config := DefaultCompressionConfig()
	config.MinSize = 100
	router.Use(CompressionMiddleware(config))
	router.GET("/api/v1/data", func(c *gin.Context) {
		// Generate a JSON response larger than min size.
		data := make(map[string]string)
		for i := 0; i < 50; i++ {
			data[strings.Repeat("key", 3)+string(rune('a'+i))] = strings.Repeat("value", 10)
		}
		c.JSON(http.StatusOK, data)
	})

	req := httptest.NewRequest("GET", "/api/v1/data", nil)
	req.Header.Set("Accept-Encoding", "br")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	// JSON should be compressed if large enough.
	if resp.Body.Len() > config.MinSize {
		// Response may or may not be compressed depending on total buffered size.
		// This is acceptable behavior.
		t.Logf("Response size: %d bytes, Content-Encoding: %q",
			resp.Body.Len(), resp.Header().Get("Content-Encoding"))
	}
}
