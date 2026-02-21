package challenges

import (
	"strings"

	"digital.vasic.challenges/pkg/env"
)

// BrowsingConfig holds configuration for browsing challenges,
// loaded from environment variables with sensible defaults.
type BrowsingConfig struct {
	BaseURL   string // API base URL
	Username  string // Admin login username
	Password  string // Admin login password
	WebAppURL string // Web app URL for Playwright
	WebAppDir string // Path to catalog-web directory
}

// LoadBrowsingConfig reads browsing challenge configuration
// from environment variables, falling back to defaults.
func LoadBrowsingConfig() *BrowsingConfig {
	return &BrowsingConfig{
		BaseURL:   env.GetOrDefault("BROWSING_API_URL", "http://localhost:8080"),
		Username:  env.GetOrDefault("ADMIN_USERNAME", "admin"),
		Password:  env.GetOrDefault("ADMIN_PASSWORD", ""), // ADMIN_PASSWORD must be set via environment variable
		WebAppURL: env.GetOrDefault("BROWSING_WEB_URL", "http://localhost:3000"),
		WebAppDir: env.GetOrDefault("CATALOG_WEB_DIR", "../catalog-web"),
	}
}

// invalidTitlePatterns lists title values that indicate missing
// or placeholder content in catalog entries.
var invalidTitlePatterns = []string{
	"unknown", "untitled", "placeholder", "n/a", "tbd", "",
}

// IsInvalidTitle checks whether a title matches any known
// invalid/placeholder pattern (case-insensitive).
func IsInvalidTitle(title string) bool {
	normalized := strings.TrimSpace(strings.ToLower(title))
	for _, pattern := range invalidTitlePatterns {
		if normalized == pattern {
			return true
		}
	}
	return false
}

// requiredMediaTypes lists the content categories that must
// be present in a fully cataloged collection.
var requiredMediaTypes = []string{
	"music", "tv_show", "movie", "software", "comic",
}
