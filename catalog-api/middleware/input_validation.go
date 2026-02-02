package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// InputValidationConfig holds validation configuration
type InputValidationConfig struct {
	// Maximum request body size in bytes
	MaxRequestBodySize int64
	// Enable SQL injection detection
	EnableSQLInjectionDetection bool
	// Enable XSS detection
	EnableXSSDetection bool
	// Enable path traversal detection
	EnablePathTraversalDetection bool
	// Custom validation rules
	CustomRules map[string]string
}

// DefaultInputValidationConfig returns secure defaults
func DefaultInputValidationConfig() InputValidationConfig {
	return InputValidationConfig{
		MaxRequestBodySize:           10 * 1024 * 1024, // 10MB
		EnableSQLInjectionDetection:  true,
		EnableXSSDetection:           true,
		EnablePathTraversalDetection: true,
		CustomRules:                  make(map[string]string),
	}
}

// Common injection patterns
var (
	// SQL Injection patterns
	sqlInjectionPatterns = []string{
		`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|UNION|SCRIPT|OR)\b)`,
		`(?i)(\b(FROM|WHERE|AND|OR|HAVING|GROUP BY|ORDER BY)\b)`,
		`(?i)(['"]\s*;\s*(SELECT|INSERT|UPDATE|DELETE|DROP))`,
		`(?i)(/\*.*\*/)`,
		`(?i)(--.*$)`,
		`(?i)(\bWAITFOR\s+DELAY\b)`,
		`(?i)(\bBENCHMARK\b)`,
		`(?i)(\bSLEEP\b)`,
		`(?i)(\bEXEC\b\s*\(|\bXP_\w+\b)`,
	}

	// XSS patterns
	xssPatterns = []string{
		`(?i)(<script[^>]*>.*?</script>)`,
		`(?i)(<iframe[^>]*>.*?</iframe>)`,
		`(?i)(<object[^>]*>.*?</object>)`,
		`(?i)(<embed[^>]*>.*?</embed>)`,
		`(?i)(javascript\s*:|vbscript\s*:|data\s*:|onload\s*=|onerror\s*=)`,
		`(?i)(on\w+\s*=)`,
		`(?i)(eval\s*\(|alert\s*\(|confirm\s*\(|prompt\s*\()`,
	}

	// Path traversal patterns
	pathTraversalPatterns = []string{
		`\.\./|\.\.\\`,
		`%2e%2e%2f|%2e%2e\\`,
		`\.\.%2f|\.\.%5c`,
		`%2e%2e%5c|%2e%2e/`,
		`/etc/passwd|/etc/shadow|/etc/hosts`,
		`windows/system32|boot\.ini|win\.ini`,
	}
)

// SanitizeInput performs basic sanitization on input strings
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Ensure valid UTF-8
	if !utf8.ValidString(input) {
		// Remove invalid UTF-8 sequences
		valid := make([]rune, 0, len(input))
		for i, r := range input {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(input[i:])
				if size == 1 {
					continue
				}
			}
			valid = append(valid, r)
		}
		input = string(valid)
	}

	return input
}

// DetectSQLInjection checks for common SQL injection patterns
func DetectSQLInjection(input string) bool {
	for _, pattern := range sqlInjectionPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

// DetectXSS checks for common XSS patterns
func DetectXSS(input string) bool {
	for _, pattern := range xssPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

// DetectPathTraversal checks for path traversal attempts
func DetectPathTraversal(input string) bool {
	for _, pattern := range pathTraversalPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

// ValidateRequestBody validates request body against security rules
func ValidateRequestBody(config InputValidationConfig, c *gin.Context) error {
	// Check content length
	if c.Request.ContentLength > config.MaxRequestBodySize {
		return fmt.Errorf("request body too large")
	}

	// Only validate specific content types and methods that typically have bodies
	contentType := c.GetHeader("Content-Type")
	method := c.Request.Method
	
	// Skip validation for GET requests (they shouldn't have bodies)
	if method == "GET" || method == "DELETE" || method == "HEAD" {
		return nil
	}
	
	// Skip validation for non-form/json content types
	if contentType != "" && 
		!strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "application/x-www-form-urlencoded") &&
		!strings.Contains(contentType, "multipart/form-data") {
		return nil // Skip validation for other content types
	}

	// Read and parse body
	bodyBytes, err := c.GetRawData()
	if err != nil {
		return fmt.Errorf("failed to read request body")
	}

	if len(bodyBytes) > int(config.MaxRequestBodySize) {
		return fmt.Errorf("request body too large")
	}

	// Restore the request body for subsequent handlers
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// For JSON requests, validate the structure
	if strings.Contains(contentType, "application/json") {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			return fmt.Errorf("invalid JSON format")
		}

		// Recursively validate all string values
		return validateJSONValues(config, jsonData)
	}

	return nil
}

// validateJSONValues recursively validates all string values in JSON
func validateJSONValues(config InputValidationConfig, data map[string]interface{}) error {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if err := validateStringValue(config, key, v); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := validateJSONValues(config, v); err != nil {
				return err
			}
		case []interface{}:
			for i, item := range v {
				if str, ok := item.(string); ok {
					if err := validateStringValue(config, fmt.Sprintf("%s[%d]", key, i), str); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// validateStringValue validates individual string values
func validateStringValue(config InputValidationConfig, key, value string) error {
	// Sanitize input
	value = SanitizeInput(value)

	// Check for SQL injection
	if config.EnableSQLInjectionDetection && DetectSQLInjection(value) {
		return fmt.Errorf("potential SQL injection detected in field: %s", key)
	}

	// Check for XSS
	if config.EnableXSSDetection && DetectXSS(value) {
		return fmt.Errorf("potential XSS detected in field: %s", key)
	}

	// Check for path traversal
	if config.EnablePathTraversalDetection && DetectPathTraversal(value) {
		return fmt.Errorf("potential path traversal detected in field: %s", key)
	}

	// Check custom rules
	if pattern, exists := config.CustomRules[key]; exists {
		matched, _ := regexp.MatchString(pattern, value)
		if !matched {
			return fmt.Errorf("field %s does not match required pattern", key)
		}
	}

	return nil
}

// InputValidation creates a middleware for input validation
func InputValidation(config InputValidationConfig) gin.HandlerFunc {
	if config.MaxRequestBodySize == 0 {
		config = DefaultInputValidationConfig()
	}

	return func(c *gin.Context) {
		// Validate request body
		if err := ValidateRequestBody(config, c); err != nil {
			// Log the attempt
			fmt.Printf("Security validation failed: %v, IP: %s, Path: %s\n",
				err, c.ClientIP(), c.Request.URL.Path)

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_failed",
				"message": "Invalid input detected",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Next()
	}
}

// ValidateStruct validates a struct using the validator library
func ValidateStruct(obj interface{}) error {
	return validate.Struct(obj)
}

// ValidateVar validates a single field
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}
