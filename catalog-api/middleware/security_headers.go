package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// EnableContentSecurityPolicy enables CSP header
	EnableContentSecurityPolicy bool
	// ContentSecurityPolicy is the CSP directive (default: restrictive policy)
	ContentSecurityPolicy string
	// EnableHSTS enables Strict-Transport-Security header
	EnableHSTS bool
	// HSTSMaxAge is the max-age for HSTS in seconds (default: 1 year)
	HSTSMaxAge int
	// HSTSIncludeSubDomains includes subdomains in HSTS
	HSTSIncludeSubDomains bool
	// HSTSPreload enables HSTS preload
	HSTSPreload bool
}

// DefaultSecurityHeadersConfig returns default secure configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableContentSecurityPolicy: true,
		ContentSecurityPolicy:       "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'none'; frame-src 'none'; base-uri 'self'; form-action 'self';",
		EnableHSTS:                  true,
		HSTSMaxAge:                  31536000, // 1 year
		HSTSIncludeSubDomains:       true,
		HSTSPreload:                 false,
	}
}

// SecurityHeaders adds standard security headers to all responses.
func SecurityHeaders() gin.HandlerFunc {
	return SecurityHeadersWithConfig(DefaultSecurityHeadersConfig())
}

// SecurityHeadersWithConfig adds security headers with custom configuration.
func SecurityHeadersWithConfig(config SecurityHeadersConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic security headers (always set)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), accelerometer=(), gyroscope=(), magnetometer=(), payment=(), usb=(), vr=()")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Content Security Policy
		if config.EnableContentSecurityPolicy && config.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// HSTS (HTTP Strict Transport Security)
		if config.EnableHSTS && (c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https") {
			hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
			if config.HSTSIncludeSubDomains {
				hstsValue += "; includeSubDomains"
			}
			if config.HSTSPreload {
				hstsValue += "; preload"
			}
			c.Header("Strict-Transport-Security", hstsValue)
		}

		c.Next()
	}
}
