package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// DatabaseConnectivityChallenge validates that the database is connected
// and operational through the running API.
type DatabaseConnectivityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDatabaseConnectivityChallenge creates CH-014.
func NewDatabaseConnectivityChallenge() *DatabaseConnectivityChallenge {
	return &DatabaseConnectivityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"database-connectivity",
			"Database Connectivity",
			"Validates database is connected: health endpoint responds, "+
				"stats are queryable, storage roots can be created and read back",
			"e2e",
			nil, // no dependencies — first in chain
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the database connectivity challenge.
func (c *DatabaseConnectivityChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Health endpoint responds (proves database is connected)
	healthCode, healthBody, healthErr := client.Get(ctx, "/health")
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
		Target:   "health_endpoint",
		Expected: "HTTP 200 with status=healthy",
		Actual:   fmt.Sprintf("HTTP %d, status=%q", healthCode, healthStatus),
		Passed:   healthPassed,
		Message:  challenge.Ternary(healthPassed, "API health check passed (database connected)", fmt.Sprintf("Health check failed: HTTP %d err=%v", healthCode, healthErr)),
	})
	if !healthPassed {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	// Step 2: Login to get auth token
	loginResp, loginErr := client.Login(ctx, c.config.Username, c.config.Password)
	loginOK := loginErr == nil && loginResp != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "admin_login",
		Expected: "successful login with JWT token",
		Actual:   fmt.Sprintf("login_ok=%v", loginOK),
		Passed:   loginOK,
		Message:  challenge.Ternary(loginOK, "Admin login succeeded", fmt.Sprintf("Login failed: %v", loginErr)),
	})
	if !loginOK {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	// Step 3: GET /stats/overall returns valid stats (database is queryable)
	statsCode, statsBody, statsErr := client.Get(ctx, "/api/v1/stats/overall")
	statsOK := statsErr == nil && statsCode == 200 && statsBody != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_endpoint",
		Expected: "HTTP 200 with stats data",
		Actual:   fmt.Sprintf("HTTP %d", statsCode),
		Passed:   statsOK,
		Message:  challenge.Ternary(statsOK, "Stats endpoint returned data (database queryable)", fmt.Sprintf("Stats query failed: HTTP %d err=%v", statsCode, statsErr)),
	})

	// Step 4: Create a test storage root (database is writable)
	testRootName := fmt.Sprintf("db-test-%d", time.Now().UnixMilli())
	createBody := fmt.Sprintf(`{"name":%q,"protocol":"local","path":"/tmp/db-test","max_depth":1}`, testRootName)
	createCode, _, createErr := client.PostJSON(ctx, "/api/v1/storage/roots", createBody)
	createOK := createErr == nil && (createCode == 201 || createCode == 200)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "create_storage_root",
		Expected: "HTTP 201 storage root created",
		Actual:   fmt.Sprintf("HTTP %d", createCode),
		Passed:   createOK,
		Message:  challenge.Ternary(createOK, "Storage root created (database writable)", fmt.Sprintf("Create failed: HTTP %d err=%v", createCode, createErr)),
	})

	// Step 5: Read back the created storage root
	rootsCode, rootsBody, rootsErr := client.Get(ctx, "/api/v1/storage/roots")
	rootsOK := rootsErr == nil && rootsCode == 200
	foundCreated := false
	if rootsOK && rootsBody != nil {
		if roots, ok := rootsBody["roots"].([]interface{}); ok {
			for _, r := range roots {
				if rm, ok := r.(map[string]interface{}); ok {
					if nm, ok := rm["name"].(string); ok && nm == testRootName {
						foundCreated = true
						break
					}
				}
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "read_after_write",
		Expected: fmt.Sprintf("storage root %q visible in GET", testRootName),
		Actual:   fmt.Sprintf("found=%v", foundCreated),
		Passed:   foundCreated,
		Message:  challenge.Ternary(foundCreated, "Read-after-write consistency verified", "Created storage root not found in GET response"),
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
