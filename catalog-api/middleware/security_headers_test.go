package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_SetsAllHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Permissions-Policy"), "camera=()")
}

func TestSecurityHeaders_NoHSTSWithoutTLS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_HSTSWithTLS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.TLS = &tls.ConnectionState{}
	r.ServeHTTP(w, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_HeadersPresentOnAllStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		status int
	}{
		{"200 OK", http.StatusOK},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(SecurityHeaders())
			r.GET("/test", func(c *gin.Context) {
				c.String(tt.status, "response")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		})
	}
}

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	assert.True(t, config.EnableContentSecurityPolicy)
	assert.NotEmpty(t, config.ContentSecurityPolicy)
	assert.True(t, config.EnableHSTS)
	assert.Equal(t, 31536000, config.HSTSMaxAge)
	assert.True(t, config.HSTSIncludeSubDomains)
	assert.False(t, config.HSTSPreload)
}

func TestSecurityHeadersWithConfig_CSPHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := SecurityHeadersConfig{
		EnableContentSecurityPolicy: true,
		ContentSecurityPolicy:       "default-src 'none';",
		EnableHSTS:                  false,
	}

	r := gin.New()
	r.Use(SecurityHeadersWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, "default-src 'none';", w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeadersWithConfig_AdditionalHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := DefaultSecurityHeadersConfig()

	r := gin.New()
	r.Use(SecurityHeadersWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Check new headers
	assert.Equal(t, "require-corp", w.Header().Get("Cross-Origin-Embedder-Policy"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Resource-Policy"))
}

func TestSecurityHeadersWithConfig_CSPDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := SecurityHeadersConfig{
		EnableContentSecurityPolicy: false,
		EnableHSTS:                  false,
	}

	r := gin.New()
	r.Use(SecurityHeadersWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeadersWithConfig_HSTSWithForwardedProto(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := SecurityHeadersConfig{
		EnableHSTS:            true,
		HSTSMaxAge:            31536000,
		HSTSIncludeSubDomains: true,
		HSTSPreload:           true,
	}

	r := gin.New()
	r.Use(SecurityHeadersWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	r.ServeHTTP(w, req)

	hsts := w.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hsts, "max-age=")
	assert.Contains(t, hsts, "includeSubDomains")
	assert.Contains(t, hsts, "preload")
}
