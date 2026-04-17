package services

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// sanitizer is the project-wide HTML sanitizer used to scrub any
// free-form metadata strings that enter the catalog from upstream
// providers (Fanart.tv, IGDB, Cover Art Archive, user-provided
// overrides). It is configured with the strictest bluemonday policy
// because our cover pipeline has no legitimate reason to accept HTML.
var sanitizer = bluemonday.StrictPolicy()

// SanitizeMetadataString returns raw stripped of any HTML markup. The
// trailing whitespace is trimmed so callers can persist the result
// directly into database columns.
func SanitizeMetadataString(raw string) string {
	return strings.TrimSpace(sanitizer.Sanitize(raw))
}

// SanitizeCoverFields clones m and scrubs every value through the
// strict HTML policy. Nil input returns nil; unchanged values share
// the allocation with the caller to avoid wasted copies.
func SanitizeCoverFields(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = SanitizeMetadataString(v)
	}
	return out
}
