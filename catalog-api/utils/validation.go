package utils

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Cleaned string
}

// StringValidator provides common string validation functions
type StringValidator struct {
	MaxLength int
	MinLength int
	Pattern   *regexp.Regexp
}

// NewStringValidator creates a new string validator with constraints
func NewStringValidator(minLen, maxLen int, pattern string) *StringValidator {
	sv := &StringValidator{
		MinLength: minLen,
		MaxLength: maxLen,
	}

	if pattern != "" {
		sv.Pattern = regexp.MustCompile(pattern)
	}

	return sv
}

// Validate checks if a string meets the validation criteria
func (sv *StringValidator) Validate(input string) ValidationResult {
	result := ValidationResult{
		Valid:   true,
		Errors:  []string{},
		Cleaned: strings.TrimSpace(input),
	}

	// Check minimum length
	if sv.MinLength > 0 && len(result.Cleaned) < sv.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors,
			"Input is too short")
	}

	// Check maximum length
	if sv.MaxLength > 0 && len(result.Cleaned) > sv.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors,
			"Input is too long")
	}

	// Check pattern
	if sv.Pattern != nil && !sv.Pattern.MatchString(result.Cleaned) {
		result.Valid = false
		result.Errors = append(result.Errors,
			"Input contains invalid characters")
	}

	return result
}

// Common validation patterns
var (
	// Alphanumeric only
	AlphaNumericPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// Alphanumeric with spaces and basic punctuation
	SafeStringPattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?():;@]+$`)

	// Email validation (basic)
	EmailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// UUID validation
	UUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// Safe filename (no path traversal)
	SafeFilenamePattern = regexp.MustCompile(`^[^/\\:*?"<>|]+$`)

	// Hex color
	HexColorPattern = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

	// IP address (IPv4)
	IPv4Pattern = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	// URL (basic)
	URLPattern = regexp.MustCompile(`^https?://[\w\-\.]+(:\d+)?(/[\w\-\./?%&=]*)?$`)
)

// IsAlphanumeric checks if string contains only letters and numbers
func IsAlphanumeric(input string) bool {
	return AlphaNumericPattern.MatchString(input)
}

// IsSafeString checks if string contains only safe characters
func IsSafeString(input string) bool {
	return SafeStringPattern.MatchString(input)
}

// IsValidEmail validates email format
func IsValidEmail(input string) bool {
	return EmailPattern.MatchString(input)
}

// IsValidUUID validates UUID format
func IsValidUUID(input string) bool {
	return UUIDPattern.MatchString(input)
}

// IsSafeFilename validates filename (no path traversal)
func IsSafeFilename(input string) bool {
	// Check pattern
	if !SafeFilenamePattern.MatchString(input) {
		return false
	}

	// Check for path traversal attempts
	if strings.Contains(input, "..") || strings.Contains(input, "//") {
		return false
	}

	return true
}

// IsValidHexColor validates hex color format
func IsValidHexColor(input string) bool {
	return HexColorPattern.MatchString(input)
}

// IsValidIPv4 validates IPv4 address format
func IsValidIPv4(input string) bool {
	return IPv4Pattern.MatchString(input)
}

// IsValidURL validates URL format
func IsValidURL(input string) bool {
	return URLPattern.MatchString(input)
}

// SanitizeString performs basic sanitization
func SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Normalize Unicode
	if !utf8.ValidString(input) {
		input = strings.ToValidUTF8(input, "")
	}

	return input
}

// SanitizeHTML removes HTML tags
func SanitizeHTML(input string) string {
	// Simple tag removal
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// EscapeHTML escapes HTML special characters
func EscapeHTML(input string) string {
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, `"`, "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	return input
}

// ContainsSQLInjection checks for common SQL injection patterns
func ContainsSQLInjection(input string) bool {
	sqlPatterns := []string{
		`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|UNION)\b)`,
		`(?i)(['"]\s*;\s*(SELECT|INSERT|UPDATE|DELETE|DROP))`,
		`(?i)(/\*.*\*/)`,
		`(?i)(--.*$)`,
		`(?i)(\bWAITFOR\s+DELAY\b)`,
		`(?i)(\bSLEEP\s*\()`,
		`(?i)(\bEXEC\s*\()`,
	}

	for _, pattern := range sqlPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return true
		}
	}

	return false
}

// ContainsXSS checks for common XSS patterns
func ContainsXSS(input string) bool {
	xssPatterns := []string{
		`(?i)(<script[^>]*>.*?</script>)`,
		`(?i)(javascript\s*:)`,
		`(?i)(vbscript\s*:)`,
		`(?i)(onload\s*=|onerror\s*=|onmouseover\s*=)`,
		`(?i)(eval\s*\()`,
	}

	for _, pattern := range xssPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return true
		}
	}

	return false
}

// ContainsPathTraversal checks for path traversal attempts
func ContainsPathTraversal(input string) bool {
	// Check for common traversal patterns
	if strings.Contains(input, "../") || strings.Contains(input, "..\\") {
		return true
	}

	// Check for encoded traversal
	if strings.Contains(input, "%2e%2e") {
		return true
	}

	// Check for null byte injection
	if strings.Contains(input, "\x00") {
		return true
	}

	return false
}

// NormalizeUnicode normalizes Unicode input
func NormalizeUnicode(input string) string {
	// Remove control characters except common whitespace
	var result strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) {
			if r == '\n' || r == '\r' || r == '\t' {
				result.WriteRune(r)
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// Truncate truncates string to max length with ellipsis
func Truncate(input string, maxLen int) string {
	if len(input) <= maxLen {
		return input
	}

	if maxLen <= 3 {
		return input[:maxLen]
	}

	return input[:maxLen-3] + "..."
}

// RemoveWhitespace removes all whitespace
func RemoveWhitespace(input string) string {
	var result strings.Builder
	for _, r := range input {
		if !unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// CompactWhitespace replaces multiple whitespace with single space
func CompactWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}
