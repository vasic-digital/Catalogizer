package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

// CompressionLevel represents the compression level for Brotli and Gzip.
type CompressionLevel int

const (
	// CompressionLevelDefault uses default compression level.
	CompressionLevelDefault CompressionLevel = iota
	// CompressionLevelBestSpeed uses best speed compression.
	CompressionLevelBestSpeed
	// CompressionLevelBestCompression uses best compression ratio.
	CompressionLevelBestCompression
)

// CompressionConfig holds configuration for compression middleware.
type CompressionConfig struct {
	// MinSize is the minimum response size in bytes to compress.
	MinSize int
	// BrotliLevel is the compression level for Brotli.
	BrotliLevel CompressionLevel
	// GzipLevel is the compression level for Gzip.
	GzipLevel CompressionLevel
	// ExcludedContentTypes are content types that should not be compressed.
	ExcludedContentTypes []string
	// ExcludedPaths are URL paths that should not be compressed.
	ExcludedPaths []string
}

// DefaultCompressionConfig returns a default compression configuration.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		MinSize:              1024,
		BrotliLevel:          CompressionLevelDefault,
		GzipLevel:            CompressionLevelDefault,
		ExcludedContentTypes: []string{"image/", "video/", "audio/", "application/octet-stream"},
		ExcludedPaths:        []string{"/metrics"},
	}
}

// CompressionMiddleware returns a Gin middleware that compresses HTTP responses
// using Brotli (preferred) or Gzip based on Accept-Encoding header.
func CompressionMiddleware(config CompressionConfig) gin.HandlerFunc {
	// Pool of Brotli writers to reuse.
	brotliWriterPool := sync.Pool{
		New: func() interface{} {
			var level int
			switch config.BrotliLevel {
			case CompressionLevelBestSpeed:
				level = brotli.BestSpeed
			case CompressionLevelBestCompression:
				level = brotli.BestCompression
			default:
				level = brotli.DefaultCompression
			}
			return brotli.NewWriterLevel(io.Discard, level)
		},
	}

	// Pool of Gzip writers to reuse.
	gzipWriterPool := sync.Pool{
		New: func() interface{} {
			var level int
			switch config.GzipLevel {
			case CompressionLevelBestSpeed:
				level = gzip.BestSpeed
			case CompressionLevelBestCompression:
				level = gzip.BestCompression
			default:
				level = gzip.DefaultCompression
			}
			w, _ := gzip.NewWriterLevel(io.Discard, level)
			return w
		},
	}

	return func(c *gin.Context) {
		// Check if path is excluded.
		for _, excludedPath := range config.ExcludedPaths {
			if strings.HasPrefix(c.Request.URL.Path, excludedPath) {
				c.Next()
				return
			}
		}

		// Check Accept-Encoding header.
		acceptEncoding := c.GetHeader("Accept-Encoding")
		supportsBrotli := strings.Contains(acceptEncoding, "br")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		// If client doesn't support compression, skip.
		if !supportsBrotli && !supportsGzip {
			c.Next()
			return
		}

		// Store original writer.
		originalWriter := c.Writer

		// Create a response writer wrapper that captures the response.
		crw := &compressResponseWriter{
			ResponseWriter: originalWriter,
			config:         config,
			supportsBrotli: supportsBrotli,
			supportsGzip:   supportsGzip,
			brotliPool:     &brotliWriterPool,
			gzipPool:       &gzipWriterPool,
		}
		c.Writer = crw

		defer func() {
			crw.Flush()
			c.Writer = originalWriter
		}()

		c.Next()
	}
}

// compressResponseWriter wraps the original ResponseWriter to compress the response.
type compressResponseWriter struct {
	gin.ResponseWriter
	config         CompressionConfig
	supportsBrotli bool
	supportsGzip   bool
	brotliPool     *sync.Pool
	gzipPool       *sync.Pool

	brotliWriter *brotli.Writer
	gzipWriter   *gzip.Writer
	buffer       []byte
	statusCode   int
	wroteHeader  bool
}

func (w *compressResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *compressResponseWriter) Write(data []byte) (int, error) {
	// Check content type for exclusion.
	contentType := w.Header().Get("Content-Type")
	for _, excludedType := range w.config.ExcludedContentTypes {
		if strings.HasPrefix(contentType, excludedType) {
			return w.ResponseWriter.Write(data)
		}
	}

	// If response is too small, don't compress.
	if len(data) < w.config.MinSize && len(w.buffer) == 0 {
		return w.ResponseWriter.Write(data)
	}

	// Buffer the data to determine total size.
	w.buffer = append(w.buffer, data...)
	return len(data), nil
}

func (w *compressResponseWriter) Flush() {
	if len(w.buffer) == 0 {
		return
	}

	// Determine total size.
	totalSize := len(w.buffer)

	// If total size is still below threshold, write uncompressed.
	if totalSize < w.config.MinSize {
		w.ResponseWriter.Write(w.buffer)
		w.buffer = nil
		return
	}

	// Choose compression algorithm: Brotli preferred over Gzip.
	var writer io.Writer = w.ResponseWriter
	var encoding string

	if w.supportsBrotli {
		// Use Brotli compression.
		bw := w.brotliPool.Get().(*brotli.Writer)
		bw.Reset(w.ResponseWriter)
		defer func() {
			bw.Close()
			w.brotliPool.Put(bw)
		}()
		writer = bw
		encoding = "br"
		w.brotliWriter = bw
	} else if w.supportsGzip {
		// Use Gzip compression.
		gw := w.gzipPool.Get().(*gzip.Writer)
		gw.Reset(w.ResponseWriter)
		defer func() {
			gw.Close()
			w.gzipPool.Put(gw)
		}()
		writer = gw
		encoding = "gzip"
		w.gzipWriter = gw
	}

	// Set Content-Encoding header.
	if encoding != "" {
		w.Header().Set("Content-Encoding", encoding)
	}
	// Set Vary header for caching.
	w.Header().Add("Vary", "Accept-Encoding")

	// Write the buffered data through the compression writer.
	writer.Write(w.buffer)
	w.buffer = nil
}

func (w *compressResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
