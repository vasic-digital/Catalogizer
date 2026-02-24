package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CacheHeaders adds Cache-Control and ETag headers for cacheable responses.
// maxAge specifies the number of seconds the response can be cached by the client.
// Only successful (2xx) GET responses receive cache headers; error responses
// and non-GET methods are left untouched.
//
// When the client sends an If-None-Match header matching the computed ETag,
// the middleware returns 304 Not Modified without a response body.
func CacheHeaders(maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only buffer GET requests for ETag support.
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Wrap the writer to buffer the response body for ETag computation.
		bw := &bufferedResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		c.Writer = bw

		c.Next()

		// Restore the original writer.
		c.Writer = bw.ResponseWriter

		// Only add cache headers for successful responses.
		if bw.statusCode >= 400 {
			// Write the buffered error response through as-is.
			bw.ResponseWriter.WriteHeader(bw.statusCode)
			for k, vals := range bw.headers {
				for _, v := range vals {
					bw.ResponseWriter.Header().Set(k, v)
				}
			}
			bw.ResponseWriter.Write(bw.body.Bytes())
			return
		}

		// Compute ETag from the buffered response body.
		bodyBytes := bw.body.Bytes()
		if len(bodyBytes) > 0 {
			hash := md5.Sum(bodyBytes)
			etag := fmt.Sprintf(`"%x"`, hash)

			// Check If-None-Match from the client.
			if match := c.GetHeader("If-None-Match"); match == etag {
				c.Writer.Header().Set("ETag", etag)
				c.Writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
				c.Writer.WriteHeader(http.StatusNotModified)
				return
			}

			c.Writer.Header().Set("ETag", etag)
		}

		// Set Cache-Control header.
		c.Writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))

		// Copy any extra headers the handler set.
		for k, vals := range bw.headers {
			for _, v := range vals {
				c.Writer.Header().Set(k, v)
			}
		}

		// Write the buffered response.
		c.Writer.WriteHeader(bw.statusCode)
		c.Writer.Write(bodyBytes)
	}
}

// StaticCacheHeaders sets long-lived cache headers for static assets.
// These assets are expected to be fingerprinted (content-hashed filenames)
// so they can be cached indefinitely.
func StaticCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Next()
	}
}

// bufferedResponseWriter buffers the response body so we can compute an ETag
// before deciding whether to send a full response or 304 Not Modified.
type bufferedResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	// Don't write to the underlying writer yet — we buffer everything.
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *bufferedResponseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func (w *bufferedResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}
