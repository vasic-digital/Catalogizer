package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// Challenges CH-153 through CH-158 verify module integration:
// registry initialization, memory monitoring, health aggregation,
// resilience facades, security guardrails, and storage resolution.

// ModuleRegistryInitChallenge validates that the module registry
// initializes without errors by checking the health endpoint
// reports all subsystems as healthy.
type ModuleRegistryInitChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewModuleRegistryInitChallenge creates CH-153.
func NewModuleRegistryInitChallenge() *ModuleRegistryInitChallenge {
	return &ModuleRegistryInitChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"module-registry-init",
			"Module Registry Init",
			"Validates that the module registry initializes "+
				"without errors: health endpoint reports healthy "+
				"and system info is accessible.",
			"module-integration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the module registry init challenge.
func (c *ModuleRegistryInitChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Health check without auth proves basic module init works
	c.ReportProgress("checking-health", nil)
	healthCode, healthBody, healthErr := client.Get(
		ctx, "/health",
	)

	healthOK := healthErr == nil && healthCode == 200
	healthStatus := ""
	if healthBody != nil {
		if s, ok := healthBody["status"].(string); ok {
			healthStatus = s
		}
	}
	healthPassed := healthOK && healthStatus == "healthy"

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "registry_health",
		Expected: "HTTP 200 with status=healthy",
		Actual: fmt.Sprintf("HTTP %d, status=%q",
			healthCode, healthStatus),
		Passed: healthPassed,
		Message: challenge.Ternary(healthPassed,
			"Module registry initialized (health=healthy)",
			fmt.Sprintf("Health check failed: HTTP %d err=%v",
				healthCode, healthErr)),
	})

	if !healthPassed {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, "",
		), nil
	}

	// Verify system info is accessible (proves lazy init worked)
	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-system-info", nil)
	infoCode, _, infoErr := client.Get(
		ctx, "/api/v1/admin/system-info",
	)

	infoOK := infoErr == nil && infoCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "system_info_accessible",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", infoCode),
		Passed:   infoOK,
		Message: challenge.Ternary(infoOK,
			"System info accessible (registry init complete)",
			fmt.Sprintf("System info returned %d, err=%v",
				infoCode, infoErr)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// MemoryMonitorActiveChallenge validates that the memory
// monitor is running by checking metrics or system info for
// memory-related data.
type MemoryMonitorActiveChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMemoryMonitorActiveChallenge creates CH-154.
func NewMemoryMonitorActiveChallenge() *MemoryMonitorActiveChallenge {
	return &MemoryMonitorActiveChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"memory-monitor-active",
			"Memory Monitor Active",
			"Validates that the memory monitor is running: "+
				"system info or metrics contain memory usage data.",
			"module-integration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the memory monitor active challenge.
func (c *MemoryMonitorActiveChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-memory-info", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/admin/system-info",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "system_info_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"System info accessible",
			fmt.Sprintf("System info returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		var body map[string]interface{}
		if jsonErr := json.Unmarshal(rawBody, &body); jsonErr == nil {
			// Look for memory-related fields at any nesting level
			hasMemory := findKey(body, "memory") ||
				findKey(body, "memory_alloc") ||
				findKey(body, "heap_alloc") ||
				findKey(body, "mem_alloc") ||
				findKey(body, "go_memstats_alloc_bytes")

			outputs["has_memory_data"] = fmt.Sprintf(
				"%t", hasMemory,
			)
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "not_empty",
				Target:   "memory_data",
				Expected: "memory-related field present",
				Actual: challenge.Ternary(hasMemory,
					"present", "missing"),
				Passed: hasMemory,
				Message: challenge.Ternary(hasMemory,
					"Memory monitor data found in system info",
					"No memory data in system info (monitor "+
						"may not be wired to system-info endpoint)"),
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

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// HealthAggregatorActiveChallenge validates that the health
// aggregator is initialized by checking the health endpoint
// returns structured health data.
type HealthAggregatorActiveChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewHealthAggregatorActiveChallenge creates CH-155.
func NewHealthAggregatorActiveChallenge() *HealthAggregatorActiveChallenge {
	return &HealthAggregatorActiveChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"health-aggregator-active",
			"Health Aggregator Active",
			"Validates that the health aggregator is initialized: "+
				"health endpoint returns structured data with status.",
			"module-integration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the health aggregator active challenge.
func (c *HealthAggregatorActiveChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	c.ReportProgress("checking-health-aggregator", nil)
	code, body, err := client.Get(ctx, "/health")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "health_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Health endpoint returned 200",
			fmt.Sprintf("Health returned %d, err=%v",
				code, err)),
	})

	if codeOK && body != nil {
		hasStatus := false
		if _, ok := body["status"].(string); ok {
			hasStatus = true
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "health_status_field",
			Expected: "status field present in health response",
			Actual: challenge.Ternary(hasStatus,
				"present", "missing"),
			Passed: hasStatus,
			Message: challenge.Ternary(hasStatus,
				"Health aggregator returning structured data",
				"Health response missing status field"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// ResilienceFacadeActiveChallenge validates that circuit breakers
// are registered by verifying the API handles failures gracefully
// (no 500 errors on non-existent storage roots).
type ResilienceFacadeActiveChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewResilienceFacadeActiveChallenge creates CH-156.
func NewResilienceFacadeActiveChallenge() *ResilienceFacadeActiveChallenge {
	return &ResilienceFacadeActiveChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"resilience-facade-active",
			"Resilience Facade Active",
			"Validates that circuit breakers are registered: "+
				"the API handles failures gracefully without 500 "+
				"errors on non-existent resources.",
			"module-integration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the resilience facade active challenge.
func (c *ResilienceFacadeActiveChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	// Request a non-existent storage root to trigger the
	// resilience facade. Should get 404, not 500.
	c.ReportProgress("testing-resilience", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/storage-roots/99999",
	)

	// Accept 404 (not found), 400 (bad request), or 200
	// (if somehow valid). Reject 500 (unhandled failure).
	no500 := err == nil && code != 500
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "resilience_no_500",
		Expected: "non-500 response",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   no500,
		Message: challenge.Ternary(no500,
			fmt.Sprintf(
				"Resilience facade active: returned %d", code,
			),
			fmt.Sprintf("Server error: %d (circuit breaker "+
				"may not be active)", code)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// GuardrailEngineActiveChallenge validates that security
// guardrails are initialized by verifying path traversal
// attempts are rejected.
type GuardrailEngineActiveChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewGuardrailEngineActiveChallenge creates CH-157.
func NewGuardrailEngineActiveChallenge() *GuardrailEngineActiveChallenge {
	return &GuardrailEngineActiveChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"guardrail-engine-active",
			"Guardrail Engine Active",
			"Validates that security guardrails are initialized: "+
				"path traversal attempts are rejected with 400.",
			"module-integration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the guardrail engine active challenge.
func (c *GuardrailEngineActiveChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	// Attempt path traversal; guardrails should reject it
	c.ReportProgress("testing-guardrails", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/media/by-path?path=../../etc/passwd",
	)

	// Guardrails should return 400 or 403. Accept 404 as well
	// (path does not exist). Reject 200 (traversal succeeded)
	// and 500 (unhandled).
	guarded := err == nil &&
		(code == 400 || code == 403 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "guardrail_rejection",
		Expected: "400 or 403",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   guarded,
		Message: challenge.Ternary(guarded,
			fmt.Sprintf(
				"Guardrail engine active: returned %d", code,
			),
			fmt.Sprintf("Path traversal not rejected: %d "+
				"(guardrails may not be active)", code)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// StorageResolverActiveChallenge validates that the storage
// resolver is initialized by verifying the storage roots
// endpoint is accessible.
type StorageResolverActiveChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewStorageResolverActiveChallenge creates CH-158.
func NewStorageResolverActiveChallenge() *StorageResolverActiveChallenge {
	return &StorageResolverActiveChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"storage-resolver-active",
			"Storage Resolver Active",
			"Validates that the storage resolver is initialized: "+
				"the storage roots endpoint is accessible and "+
				"returns a valid response.",
			"module-integration",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the storage resolver active challenge.
func (c *StorageResolverActiveChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-storage-resolver", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/storage-roots",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "storage_roots_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Storage resolver active: roots endpoint returned 200",
			fmt.Sprintf("Storage roots returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// findKey searches recursively for a key in a nested map.
func findKey(m map[string]interface{}, key string) bool {
	if _, ok := m[key]; ok {
		return true
	}
	for _, v := range m {
		if sub, ok := v.(map[string]interface{}); ok {
			if findKey(sub, key) {
				return true
			}
		}
	}
	return false
}
