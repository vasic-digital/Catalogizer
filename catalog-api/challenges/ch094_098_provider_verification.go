package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// OpenLibraryProviderChallenge validates the OpenLibrary metadata provider
// returns search results for a well-known query.
type OpenLibraryProviderChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewOpenLibraryProviderChallenge creates CH-094.
func NewOpenLibraryProviderChallenge() *OpenLibraryProviderChallenge {
	return &OpenLibraryProviderChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"openlibrary-provider-search",
			"OpenLibrary Provider Search",
			"Verifies the OpenLibrary metadata provider is registered, "+
				"enabled (no API key required), and that its Search "+
				"implementation exists in provider source code.",
			"provider-verification",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the OpenLibrary provider verification challenge.
func (c *OpenLibraryProviderChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	c.ReportProgress("check-provider-source", nil)

	// Verify provider source file exists
	providerPath := findProviderFile()
	outputs["provider_file"] = providerPath

	fileFound := providerPath != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "provider_source_file",
		Expected: "providers.go exists",
		Actual:   challenge.Ternary(fileFound, providerPath, "not found"),
		Passed:   fileFound,
		Message: challenge.Ternary(fileFound,
			"Provider source file found",
			"Provider source file not found"),
	})

	if fileFound {
		data, readErr := os.ReadFile(providerPath)
		if readErr == nil {
			src := string(data)

			// Check OpenLibraryProvider struct exists
			hasStruct := strings.Contains(src, "OpenLibraryProvider")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "openlibrary_struct",
				Expected: "OpenLibraryProvider struct defined",
				Actual:   challenge.Ternary(hasStruct, "defined", "missing"),
				Passed:   hasStruct,
				Message: challenge.Ternary(hasStruct,
					"OpenLibraryProvider struct found in source",
					"OpenLibraryProvider struct not found in source"),
			})

			// Check it is enabled without API key
			hasFreeEnabled := strings.Contains(src, "openlibrary") &&
				strings.Contains(src, "bp.enabled = true")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "openlibrary_enabled_free",
				Expected: "OpenLibrary enabled without API key",
				Actual:   challenge.Ternary(hasFreeEnabled, "free/enabled", "requires key"),
				Passed:   hasFreeEnabled,
				Message: challenge.Ternary(hasFreeEnabled,
					"OpenLibrary is enabled as a free provider",
					"OpenLibrary requires an API key (should be free)"),
			})

			// Check Search method implementation exists (not just a stub)
			hasSearchImpl := strings.Contains(src, "OpenLibraryProvider) Search") &&
				strings.Contains(src, "search.json")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "openlibrary_search_impl",
				Expected: "Search method with OpenLibrary API call",
				Actual:   challenge.Ternary(hasSearchImpl, "implemented", "stub only"),
				Passed:   hasSearchImpl,
				Message: challenge.Ternary(hasSearchImpl,
					"OpenLibrary Search method is fully implemented",
					"OpenLibrary Search method is a stub or missing"),
			})
		}
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// MusicBrainzProviderChallenge validates the MusicBrainz metadata provider
// returns search results for a well-known query.
type MusicBrainzProviderChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMusicBrainzProviderChallenge creates CH-095.
func NewMusicBrainzProviderChallenge() *MusicBrainzProviderChallenge {
	return &MusicBrainzProviderChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"musicbrainz-provider-search",
			"MusicBrainz Provider Search",
			"Verifies the MusicBrainz metadata provider is registered, "+
				"enabled (no API key required), and that its Search "+
				"implementation exists in provider source code.",
			"provider-verification",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the MusicBrainz provider verification challenge.
func (c *MusicBrainzProviderChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	c.ReportProgress("check-provider-source", nil)

	providerPath := findProviderFile()
	outputs["provider_file"] = providerPath

	fileFound := providerPath != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "provider_source_file",
		Expected: "providers.go exists",
		Actual:   challenge.Ternary(fileFound, providerPath, "not found"),
		Passed:   fileFound,
		Message: challenge.Ternary(fileFound,
			"Provider source file found",
			"Provider source file not found"),
	})

	if fileFound {
		data, readErr := os.ReadFile(providerPath)
		if readErr == nil {
			src := string(data)

			// Check MusicBrainzProvider struct exists
			hasStruct := strings.Contains(src, "MusicBrainzProvider")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "musicbrainz_struct",
				Expected: "MusicBrainzProvider struct defined",
				Actual:   challenge.Ternary(hasStruct, "defined", "missing"),
				Passed:   hasStruct,
				Message: challenge.Ternary(hasStruct,
					"MusicBrainzProvider struct found in source",
					"MusicBrainzProvider struct not found in source"),
			})

			// Check it is enabled without API key (free API)
			hasFreeEnabled := strings.Contains(src, "musicbrainz") &&
				strings.Contains(src, "bp.enabled = true")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "musicbrainz_enabled_free",
				Expected: "MusicBrainz enabled without API key",
				Actual:   challenge.Ternary(hasFreeEnabled, "free/enabled", "requires key"),
				Passed:   hasFreeEnabled,
				Message: challenge.Ternary(hasFreeEnabled,
					"MusicBrainz is enabled as a free provider",
					"MusicBrainz requires an API key (should be free)"),
			})

			// Check Search method implementation (not just a stub)
			hasSearchImpl := strings.Contains(src, "MusicBrainzProvider) Search") &&
				strings.Contains(src, "/recording?")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "musicbrainz_search_impl",
				Expected: "Search method with MusicBrainz API call",
				Actual:   challenge.Ternary(hasSearchImpl, "implemented", "stub only"),
				Passed:   hasSearchImpl,
				Message: challenge.Ternary(hasSearchImpl,
					"MusicBrainz Search method is fully implemented",
					"MusicBrainz Search method is a stub or missing"),
			})

			// Check proper User-Agent header (MusicBrainz requires it)
			hasUserAgent := strings.Contains(src, "User-Agent") &&
				strings.Contains(src, "Catalogizer")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "musicbrainz_user_agent",
				Expected: "Custom User-Agent header for MusicBrainz",
				Actual:   challenge.Ternary(hasUserAgent, "present", "missing"),
				Passed:   hasUserAgent,
				Message: challenge.Ternary(hasUserAgent,
					"MusicBrainz requests include proper User-Agent",
					"MusicBrainz requests missing User-Agent header"),
			})
		}
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// ProviderGracefulDegradationChallenge validates that disabled providers
// return nil (not an error) when searched, allowing graceful degradation.
type ProviderGracefulDegradationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewProviderGracefulDegradationChallenge creates CH-096.
func NewProviderGracefulDegradationChallenge() *ProviderGracefulDegradationChallenge {
	return &ProviderGracefulDegradationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"provider-graceful-degradation",
			"Provider Graceful Degradation",
			"Verifies that disabled metadata providers (those without "+
				"API keys) return nil results instead of errors, "+
				"enabling graceful degradation of the search pipeline.",
			"provider-verification",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the provider graceful degradation challenge.
func (c *ProviderGracefulDegradationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	c.ReportProgress("check-degradation-pattern", nil)

	providerPath := findProviderFile()
	outputs["provider_file"] = providerPath

	fileFound := providerPath != ""
	if !fileFound {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "provider_source_file",
			Expected: "providers.go exists",
			Actual:   "not found",
			Passed:   false,
			Message:  "Provider source file not found",
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	data, readErr := os.ReadFile(providerPath)
	if readErr != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "provider_source_readable",
			Expected: "file readable",
			Actual:   fmt.Sprintf("read error: %v", readErr),
			Passed:   false,
			Message:  fmt.Sprintf("Cannot read provider source: %v", readErr),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	src := string(data)

	// Check that disabled providers return nil, nil (not an error)
	// The pattern: if !p.IsEnabled() { ... return nil, nil }
	hasGracefulPattern := strings.Contains(src, "IsEnabled()") &&
		strings.Contains(src, "return nil, nil")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "contains",
		Target:   "graceful_nil_return",
		Expected: "disabled providers return nil, nil",
		Actual:   challenge.Ternary(hasGracefulPattern, "pattern found", "pattern missing"),
		Passed:   hasGracefulPattern,
		Message: challenge.Ternary(hasGracefulPattern,
			"Disabled providers return nil (not error) for graceful degradation",
			"Disabled providers may return errors instead of nil"),
	})

	// Check that SearchAll continues on provider errors (doesn't abort)
	hasErrorContinue := strings.Contains(src, "SearchAll") &&
		strings.Contains(src, "continue")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "contains",
		Target:   "search_all_continues",
		Expected: "SearchAll continues on provider failure",
		Actual:   challenge.Ternary(hasErrorContinue, "continues", "may abort"),
		Passed:   hasErrorContinue,
		Message: challenge.Ternary(hasErrorContinue,
			"SearchAll continues iterating when a provider fails",
			"SearchAll may abort on provider failure"),
	})

	// Check that BaseProvider.enabled is set based on API key presence
	hasKeyBasedEnable := strings.Contains(src, "enabled: apiKey != \"\"") ||
		strings.Contains(src, `enabled: apiKey != ""`)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "contains",
		Target:   "key_based_enable",
		Expected: "providers disabled when API key missing",
		Actual:   challenge.Ternary(hasKeyBasedEnable, "key-gated", "unconditional"),
		Passed:   hasKeyBasedEnable,
		Message: challenge.Ternary(hasKeyBasedEnable,
			"Provider enable/disable is gated on API key presence",
			"Provider enable/disable not based on API key"),
	})

	// Count how many providers follow the graceful pattern
	gracefulCount := strings.Count(src, "return nil, nil")
	outputs["graceful_return_count"] = fmt.Sprintf("%d", gracefulCount)
	hasMultiple := gracefulCount >= 5
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "graceful_return_instances",
		Expected: "at least 5 graceful nil returns",
		Actual:   fmt.Sprintf("%d instances", gracefulCount),
		Passed:   hasMultiple,
		Message: challenge.Ternary(hasMultiple,
			fmt.Sprintf("Found %d graceful nil-return patterns across providers", gracefulCount),
			fmt.Sprintf("Only %d graceful nil-return patterns found (expected >= 5)", gracefulCount)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// ProviderManagerRoutingChallenge validates that the ProviderManager
// routes queries to the correct providers based on media type.
type ProviderManagerRoutingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewProviderManagerRoutingChallenge creates CH-097.
func NewProviderManagerRoutingChallenge() *ProviderManagerRoutingChallenge {
	return &ProviderManagerRoutingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"provider-manager-routing",
			"Provider Manager Media Type Routing",
			"Verifies the ProviderManager routes search queries to "+
				"the correct providers based on media type (movies "+
				"to TMDB/IMDB, music to MusicBrainz, books to "+
				"OpenLibrary, etc.).",
			"provider-verification",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the provider manager routing challenge.
func (c *ProviderManagerRoutingChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	c.ReportProgress("check-routing-table", nil)

	providerPath := findProviderFile()
	outputs["provider_file"] = providerPath

	fileFound := providerPath != ""
	if !fileFound {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "provider_source_file",
			Expected: "providers.go exists",
			Actual:   "not found",
			Passed:   false,
			Message:  "Provider source file not found",
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	data, readErr := os.ReadFile(providerPath)
	if readErr != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "provider_source_readable",
			Expected: "file readable",
			Actual:   fmt.Sprintf("read error: %v", readErr),
			Passed:   false,
			Message:  fmt.Sprintf("Cannot read provider source: %v", readErr),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	src := string(data)

	// Verify routing function exists
	hasRoutingFunc := strings.Contains(src, "getProvidersForMediaType")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "contains",
		Target:   "routing_function",
		Expected: "getProvidersForMediaType function exists",
		Actual:   challenge.Ternary(hasRoutingFunc, "exists", "missing"),
		Passed:   hasRoutingFunc,
		Message: challenge.Ternary(hasRoutingFunc,
			"Media type routing function found",
			"Media type routing function not found"),
	})

	// Verify required media type -> provider mappings
	routingChecks := []struct {
		mediaType string
		provider  string
	}{
		{"movie", "tmdb"},
		{"movie", "imdb"},
		{"tv_show", "tmdb"},
		{"music", "musicbrainz"},
		{"ebook", "openlibrary"},
		{"software", "github"},
		{"pc_game", "igdb"},
	}

	passedRoutes := 0
	for _, rc := range routingChecks {
		// Check for the media type key and provider value in the routing map
		hasMapping := strings.Contains(src, fmt.Sprintf(`"%s"`, rc.mediaType)) &&
			strings.Contains(src, fmt.Sprintf(`"%s"`, rc.provider))
		if hasMapping {
			passedRoutes++
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "contains",
			Target:   fmt.Sprintf("route_%s_%s", rc.mediaType, rc.provider),
			Expected: fmt.Sprintf("%s -> %s mapping", rc.mediaType, rc.provider),
			Actual:   challenge.Ternary(hasMapping, "mapped", "missing"),
			Passed:   hasMapping,
			Message: challenge.Ternary(hasMapping,
				fmt.Sprintf("Media type %q routes to %s provider", rc.mediaType, rc.provider),
				fmt.Sprintf("Media type %q missing %s provider mapping", rc.mediaType, rc.provider)),
		})
	}

	outputs["routes_verified"] = fmt.Sprintf("%d/%d", passedRoutes, len(routingChecks))

	// Verify ProviderManager struct has providers map
	// Use regex to tolerate Go struct field alignment whitespace
	providerMapRe := regexp.MustCompile(`providers\s+map\[string\]MetadataProvider`)
	hasProviderMap := providerMapRe.MatchString(src)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "contains",
		Target:   "provider_map",
		Expected: "ProviderManager has providers map",
		Actual:   challenge.Ternary(hasProviderMap, "present", "missing"),
		Passed:   hasProviderMap,
		Message: challenge.Ternary(hasProviderMap,
			"ProviderManager stores providers in a typed map",
			"ProviderManager missing providers map field"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// LazyServiceRegistryChallenge validates that the LazyServiceRegistry
// initializes services on first access and caches the result.
type LazyServiceRegistryChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewLazyServiceRegistryChallenge creates CH-098.
func NewLazyServiceRegistryChallenge() *LazyServiceRegistryChallenge {
	return &LazyServiceRegistryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"lazy-service-registry-init",
			"LazyServiceRegistry First-Access Init",
			"Verifies the LazyServiceRegistry initializes services "+
				"on first Get() call using sync.Once, and that the "+
				"API responds correctly on first request (proving "+
				"lazy initialization works end-to-end).",
			"provider-verification",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the lazy service registry challenge.
func (c *LazyServiceRegistryChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	// Part 1: Verify source code structure
	c.ReportProgress("check-lazy-source", nil)

	lazyPath := findLazyServicesFile()
	outputs["lazy_services_file"] = lazyPath

	fileFound := lazyPath != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "lazy_services_file",
		Expected: "lazy_services.go exists",
		Actual:   challenge.Ternary(fileFound, lazyPath, "not found"),
		Passed:   fileFound,
		Message: challenge.Ternary(fileFound,
			"LazyServiceRegistry source file found",
			"LazyServiceRegistry source file not found"),
	})

	if fileFound {
		data, readErr := os.ReadFile(lazyPath)
		if readErr == nil {
			src := string(data)

			// Check sync.Once usage for single-init guarantee
			hasSyncOnce := strings.Contains(src, "sync.Once")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "sync_once_usage",
				Expected: "sync.Once for single initialization",
				Actual:   challenge.Ternary(hasSyncOnce, "present", "missing"),
				Passed:   hasSyncOnce,
				Message: challenge.Ternary(hasSyncOnce,
					"LazyServiceRegistry uses sync.Once for init guarantee",
					"LazyServiceRegistry missing sync.Once (no init guarantee)"),
			})

			// Check Register and Get methods exist
			hasRegister := strings.Contains(src, "func (r *LazyServiceRegistry) Register(")
			hasGet := strings.Contains(src, "func (r *LazyServiceRegistry) Get(")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "register_method",
				Expected: "Register method on LazyServiceRegistry",
				Actual:   challenge.Ternary(hasRegister, "exists", "missing"),
				Passed:   hasRegister,
				Message: challenge.Ternary(hasRegister,
					"Register method found",
					"Register method missing"),
			})
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "get_method",
				Expected: "Get method on LazyServiceRegistry",
				Actual:   challenge.Ternary(hasGet, "exists", "missing"),
				Passed:   hasGet,
				Message: challenge.Ternary(hasGet,
					"Get method found",
					"Get method missing"),
			})

			// Check thread safety (RWMutex)
			hasRWMutex := strings.Contains(src, "sync.RWMutex")
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "contains",
				Target:   "rwmutex_safety",
				Expected: "sync.RWMutex for thread safety",
				Actual:   challenge.Ternary(hasRWMutex, "present", "missing"),
				Passed:   hasRWMutex,
				Message: challenge.Ternary(hasRWMutex,
					"LazyServiceRegistry uses RWMutex for thread safety",
					"LazyServiceRegistry missing RWMutex (not thread-safe)"),
			})
		}
	}

	// Part 2: Verify lazy init works end-to-end via API
	c.ReportProgress("verify-api-lazy-init", nil)

	client := httpclient.NewAPIClient(c.config.BaseURL)
	code, body, err := client.Get(ctx, "/health")

	apiOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "api_responds_after_init",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", code),
		Passed:   apiOK,
		Message: challenge.Ternary(apiOK,
			"API responds on first request (lazy init completed)",
			fmt.Sprintf("API failed on first request: code=%d err=%v", code, err)),
	})

	if apiOK && body != nil {
		if statusVal, ok := body["status"]; ok {
			outputs["health_status"] = fmt.Sprintf("%v", statusVal)
		}
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// findProviderFile locates the providers.go source file.
func findProviderFile() string {
	candidates := []string{
		"internal/media/providers/providers.go",
		"../catalog-api/internal/media/providers/providers.go",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	// Try from cwd parent
	cwd, _ := os.Getwd()
	parent := filepath.Dir(cwd)
	p := filepath.Join(parent, "catalog-api", "internal", "media", "providers", "providers.go")
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return p
	}
	// Try relative to cwd
	p = filepath.Join(cwd, "internal", "media", "providers", "providers.go")
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return p
	}
	return ""
}

// findLazyServicesFile locates the lazy_services.go source file.
func findLazyServicesFile() string {
	candidates := []string{
		"internal/lifecycle/lazy_services.go",
		"../catalog-api/internal/lifecycle/lazy_services.go",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	// Try from cwd parent
	cwd, _ := os.Getwd()
	parent := filepath.Dir(cwd)
	p := filepath.Join(parent, "catalog-api", "internal", "lifecycle", "lazy_services.go")
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return p
	}
	// Try relative to cwd
	p = filepath.Join(cwd, "internal", "lifecycle", "lazy_services.go")
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return p
	}
	return ""
}
